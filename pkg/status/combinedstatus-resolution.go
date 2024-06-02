/*
Copyright 2024 The KubeStellar Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package status

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/cel-go/common/types/ref"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

// combinedStatusResolution is a struct that represents the resolution of a
// combinedstatus. A combinedstatus resolution is associated with a (binding,
// object) tuple.
// A resolution's logic is as follows:
// - Composition:
//   - A combinedstatus resolution is composed of multiple statuscollector
//     data.
//   - statuscollector data is composed of collector-clauses. Each
//     collector-clause has an expression that is evaluated against every
//     workstatus that is associated with the statuscollector, and the result
//     is cached.
//
// - Flow:
//   - When a statuscollector is added to the combinedstatus resolution,
//     every clause in the statuscollector is evaluated against every workstatus
//     that is associated with the statuscollector.
//   - When a workstatus is added to / modified in the combinedstatus
//     resolution, every clause in every statuscollector is evaluated against
//     the workstatus. The caching of collector-clause -> evaluation makes
//     calculating combinedFields efficient if present.
type combinedStatusResolution struct {
	sync.RWMutex

	// statusCollectorNameToData is a map of status collector names to
	// their corresponding data.
	// If a status collector name is mapped to nil, it means that the status
	// collector has not been processed.
	statusCollectorNameToData map[string]*statusCollectorData
}

// statusCollectorData is a struct that represents the data of a status
// collector in a combinedstatus resolution.
type statusCollectorData struct {
	*v1alpha1.StatusCollectorSpec

	// workStatusRefToData is a map of workstatus-hosting WEC name to the
	// evaluation of the workstatus against the statuscollector's clauses.
	workStatusRefToData map[string]*workStatusData
}

// workStatusData is a struct that represents the evaluation of a workstatus
// against a statuscollector's clauses.
type workStatusData struct {
	// groupByEval is a map of groupBy expression names to their evaluated values.
	groupByEval map[string]ref.Val
	// combinedFieldsEval is a map of combinedFields expression names to their
	// evaluated values. CombinedField types:
	// - COUNT: the number of workstatuses per groupBy value. If groupBy is
	//   empty, the count is the number of workstatuses.
	// - SUM: the sum of the values of the workstatuses groupBy values.
	// - AVG: the average of the values of the workstatuses groupBy values.
	// - MIN: the minimum value of the workstatuses groupBy values.
	// - MAX: the maximum value of the workstatuses groupBy values.
	// If a combinedField's eval is nil, its TYPE is COUNT.
	combinedFieldsEval map[string]ref.Val

	selectEval map[string]ref.Val
}

// setStatusCollectors sets ALL the statuscollectors relevant to the
// combinedstatus resolution.
// The given map is expected not to be mutated during this call, its values
// contain valid expressions, and are immutable.
// If a value is nil, it is assumed that the associated statuscollector has not
// been processed.
// The function returns a tuple (removedSome, addedSome):
//
// - removedSome: true if one or more statuscollectors were removed.
//
// - addedSome: true if one or more statuscollectors were added.
func (c *combinedStatusResolution) setStatusCollectors(statusCollectorNameToSpec map[string]*v1alpha1.StatusCollectorSpec) (bool, bool) {
	c.Lock()
	defer c.Unlock()

	removedSome, addedSome := false, false

	// remove statuscollector data that are not relevant anymore and update the
	// statuscollector data that are. If one of the latter is updated, mark it as added
	for statusCollectorName, statusCollectorData := range c.statusCollectorNameToData {
		statusCollectorSpec, ok := statusCollectorNameToSpec[statusCollectorName]
		if !ok {
			delete(c.statusCollectorNameToData, statusCollectorName)
			removedSome = true
			continue
		}

		if statusCollectorSpecsMatch(statusCollectorData.StatusCollectorSpec, statusCollectorSpec) {
			c.statusCollectorNameToData[statusCollectorName].StatusCollectorSpec = statusCollectorSpec
			addedSome = true
		}
	}

	// add new statuscollector data
	for statusCollectorName, statusCollectorSpec := range statusCollectorNameToSpec {
		if _, ok := c.statusCollectorNameToData[statusCollectorName]; !ok {
			c.statusCollectorNameToData[statusCollectorName] = &statusCollectorData{
				StatusCollectorSpec: statusCollectorSpec,
				workStatusRefToData: make(map[string]*workStatusData),
			}

			addedSome = true
		}
	}

	return removedSome, addedSome
}

// updateStatusCollector updates the status collector data in the
// combinedstatus resolution. If the status collector is not relevant to the
// latter, the function returns false. The function returns true if the status
// collector data is updated. The given spec is assumed to be valid and
// immutable.
func (c *combinedStatusResolution) updateStatusCollector(statusCollectorName string,
	statusCollectorSpec *v1alpha1.StatusCollectorSpec) bool {
	c.Lock()
	defer c.Unlock()

	scData, ok := c.statusCollectorNameToData[statusCollectorName]
	if !ok {
		return false // statusCollector is irrelevant to this combinedstatus resolution
	}

	if statusCollectorSpecsMatch(scData.StatusCollectorSpec, statusCollectorSpec) {
		return false // statusCollector data is already up-to-date
	}

	// status collector clauses need to be updated, therefore update fields
	// and invalidate all cached workstatus evaluations by resetting the map
	c.statusCollectorNameToData[statusCollectorName] = &statusCollectorData{
		StatusCollectorSpec: statusCollectorSpec,
		workStatusRefToData: make(map[string]*workStatusData),
	}

	return true
}

// evaluateWorkStatus evaluates the workstatus per relevant statuscollector.
// The function returns true if any statuscollector data is updated.
func (c *combinedStatusResolution) evaluateWorkStatus(ctx context.Context, celEvaluator *celEvaluator,
	workStatusWECName string, workStatusContent *runtime.RawExtension) (bool, error) {
	c.Lock()
	defer c.Unlock()

	logger := klog.FromContext(ctx)

	updated := false
	for statusCollectorName, scData := range c.statusCollectorNameToData {
		changed, err := evaluateWorkStatusAgainstStatusCollector(celEvaluator, workStatusWECName, workStatusContent,
			scData)
		if err != nil {
			logger.Error(err, "failed to evaluate workstatus against statuscollector",
				"statusCollectorName", statusCollectorName)
			return false, nil
		}

		updated = updated || changed
	}

	return updated, nil
}

// evaluateWorkStatusAgainstStatusCollector evaluates the workstatus against
// the statuscollector clauses and caches the evaluations. If the workstatus
// does not match the filter, no other clause is evaluated.
// The function returns true if an evaluation is updated.
// If any evaluation fails, the function returns an error.
// The function assumes that the caller holds a lock over the combinedstatus
// resolution.
func evaluateWorkStatusAgainstStatusCollector(celEvaluator *celEvaluator, workStatusWECName string,
	workStatusContent *runtime.RawExtension, scData *statusCollectorData) (bool, error) {
	wsData, exists := scData.workStatusRefToData[workStatusWECName]
	if !exists {
		wsData = &workStatusData{}
		scData.workStatusRefToData[workStatusWECName] = wsData
	}

	// evaluate all clauses
	if scData.Filter != nil {
		eval, err := celEvaluator.Evaluate(*scData.Filter, workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate filter expression: %w", err)
		}

		if eval.Type().TypeName() != "bool" {
			return false, fmt.Errorf("filter expression must evaluate to a boolean")
		}

		if !eval.Value().(bool) && !exists { // workstatus not previously evaluated and still irrelevant
			return false, nil
		}
	}

	updated := false

	// evaluate select
	selectEvals := make(map[string]ref.Val)
	for _, selectNamedExp := range scData.Select {
		eval, err := celEvaluator.Evaluate(selectNamedExp.Def, workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate select expression: %w", err)
		}

		updated = updated || eval.Value() != wsData.selectEval[selectNamedExp.Name]
		selectEvals[selectNamedExp.Name] = eval
	}

	// evaluate groupBy
	groupByEvals := make(map[string]ref.Val)
	for _, groupByNamedExp := range scData.GroupBy {
		eval, err := celEvaluator.Evaluate(groupByNamedExp.Def, workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate groupBy expression: %w", err)
		}

		updated = updated || eval.Value() != wsData.groupByEval[groupByNamedExp.Name]
		groupByEvals[groupByNamedExp.Name] = eval
	}

	// evaluate combinedFields
	combinedFieldEvals := make(map[string]ref.Val)
	for _, combinedFieldNamedAgg := range scData.CombinedFields {
		if combinedFieldNamedAgg.Type == v1alpha1.AggregatorTypeCount {
			// count does not require a subject - mark the evaluation with the type
			currentEval, exists := wsData.combinedFieldsEval[combinedFieldNamedAgg.Name]
			updated = updated || !exists || currentEval != nil

			combinedFieldEvals[combinedFieldNamedAgg.Name] = nil
			continue
		}

		// evaluate subject which should not be nil since the statuscollector is valid
		eval, err := celEvaluator.Evaluate(*combinedFieldNamedAgg.Subject, workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate combinedFields expression: %w", err)
		}

		updated = updated || eval.Value() != wsData.combinedFieldsEval[combinedFieldNamedAgg.Name]
		combinedFieldEvals[combinedFieldNamedAgg.Name] = eval
	}

	// update the workstatus data
	wsData.selectEval = selectEvals
	wsData.groupByEval = groupByEvals
	wsData.combinedFieldsEval = combinedFieldEvals

	return updated, nil
}
