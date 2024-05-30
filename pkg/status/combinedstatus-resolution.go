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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
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

	// workStatusRefToData is a map of workstatus references to their
	// corresponding work status data.
	// A workstatus reference is identical to its source object's identifier,
	// apart from the namespace, which is set to the associated clusterName.
	// If a workstatus reference is mapped to nil, it means that the workstatus
	// has not been evaluated against the statuscollector's clauses.
	workStatusRefToData map[util.ObjectIdentifier]workStatusData
}

// workStatusData is a struct that represents the evaluation of a workstatus
// against a statuscollector's clauses.
type workStatusData struct {
	// filterEval is a boolean that represents the evaluation of the filter
	// expression against the workstatus. If true, the workstatus is included
	// in the statuscollector's result.
	filterEval bool

	// groupByEval is a map of groupBy expression names to their evaluated values.
	groupByEval map[string]any
	// combinedFieldsEval is a map of combinedFields expression names to their
	// evaluated values. CombinedField types:
	// - COUNT: the number of workstatuses per groupBy value. If groupBy is
	//   empty, the count is the number of workstatuses.
	// - SUM: the sum of the values of the workstatuses groupBy values.
	// - AVG: the average of the values of the workstatuses groupBy values.
	// - MIN: the minimum value of the workstatuses groupBy values.
	// - MAX: the maximum value of the workstatuses groupBy values.
	combinedFieldsEval map[string]any

	selectEval map[string]any
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
				workStatusRefToData: make(map[util.ObjectIdentifier]workStatusData),
			}

			addedSome = true
		}
	}

	return removedSome, addedSome
}

// getIdentifier returns the identifier of the combinedstatus resolution.
// TODO: implement
func (c *combinedStatusResolution) getIdentifier() util.ObjectIdentifier {
	return util.ObjectIdentifier{}
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
		workStatusRefToData: make(map[util.ObjectIdentifier]workStatusData),
	}

	return true
}

// evaluateWorkStatus evaluates the workstatus per relevant statuscollector.
// The function returns true if any statuscollector data is updated.
func (c *combinedStatusResolution) evaluateWorkStatus(ctx context.Context, celEvaluator *celEvaluator,
	workStatusIdentifier util.ObjectIdentifier, workStatusContent *runtime.RawExtension) (bool, error) {
	c.Lock()
	defer c.Unlock()

	logger := klog.FromContext(ctx)

	updated := false
	for statusCollectorName, scData := range c.statusCollectorNameToData {
		changed, err := c.evaluateWorkStatusAgainstStatusCollector(celEvaluator, workStatusIdentifier,
			workStatusContent, scData)
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
// the statuscollector clauses and caches the evaluations. The function returns
// true if an evaluation is updated.
// If any evaluation fails, the function returns an error.
func (c *combinedStatusResolution) evaluateWorkStatusAgainstStatusCollector(celEvaluator *celEvaluator,
	workStatusIdentifier util.ObjectIdentifier, workStatusContent *runtime.RawExtension,
	scData *statusCollectorData) (bool, error) {
	wsData, ok := scData.workStatusRefToData[workStatusIdentifier]
	if !ok {
		wsData = workStatusData{}
		scData.workStatusRefToData[workStatusIdentifier] = wsData
	}

	updated := false
	// evaluate all clauses
	if scData.Filter != nil {
		eval, err := celEvaluator.Evaluate(string(*scData.Filter), workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate filter expression: %w", err)
		}

		if eval.Type().TypeName() != "bool" {
			return false, fmt.Errorf("filter expression must evaluate to a boolean")
		}

		fiterEval := eval.Value().(bool)
		updated = fiterEval != wsData.filterEval
		wsData.filterEval = fiterEval

		if !fiterEval {
			// workstatus does not match the filter
			return updated, nil
		}
	}

	// evaluate select
	selectEvals := make(map[string]any)
	for _, selectNamedExp := range scData.Select {
		eval, err := celEvaluator.Evaluate(string(selectNamedExp.Def), workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate select expression: %w", err)
		}

		selectEval := eval.Value()
		updated = updated || selectEval != wsData.selectEval[selectNamedExp.Name]
		selectEvals[selectNamedExp.Name] = selectEval
	}

	// evaluate groupBy
	groupByEvals := make(map[string]any)
	for _, groupByNamedExp := range scData.GroupBy {
		eval, err := celEvaluator.Evaluate(string(groupByNamedExp.Def), workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate groupBy expression: %w", err)
		}

		groupByEval := eval.Value()
		updated = updated || groupByEval != wsData.groupByEval[groupByNamedExp.Name]
		groupByEvals[groupByNamedExp.Name] = groupByEval
	}

	// evaluate combinedFields
	combinedFieldEvals := make(map[string]any)
	for _, combinedFieldNamedAgg := range scData.CombinedFields {
		if combinedFieldNamedAgg.Type == v1alpha1.AggregatorTypeCount {
			// count does not require a subject - mark the evaluation with the type
			updated = updated || wsData.combinedFieldsEval[combinedFieldNamedAgg.Name] != v1alpha1.AggregatorTypeCount
			combinedFieldEvals[combinedFieldNamedAgg.Name] = v1alpha1.AggregatorTypeCount
			continue
		}

		// evaluate subject which should not be nil since the statuscollector is valid
		eval, err := celEvaluator.Evaluate(string(*combinedFieldNamedAgg.Subject), workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate combinedFields expression: %w", err)
		}

		combinedFieldEval := eval.Value()
		updated = updated || combinedFieldEval != wsData.combinedFieldsEval[combinedFieldNamedAgg.Name]
		combinedFieldEvals[combinedFieldNamedAgg.Name] = combinedFieldEval
	}

	// update the workstatus data
	wsData.selectEval = selectEvals
	wsData.groupByEval = groupByEvals
	wsData.combinedFieldsEval = combinedFieldEvals

	return updated, nil
}
