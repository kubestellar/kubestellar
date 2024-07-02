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
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/cel-go/common/types/ref"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime2 "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const nilValue = "nil"

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

	name string
	ns   string

	// statusCollectorNameToData is a map of status collector names to
	// their corresponding data.
	statusCollectorNameToData map[string]*statusCollectorData
	// collectionDestinations is a set of destinations that are expected to be
	// collected from.
	collectionDestinations sets.Set[string]
}

// statusCollectorData is a struct that represents the data of a status
// collector in a combinedstatus resolution.
type statusCollectorData struct {
	*v1alpha1.StatusCollectorSpec

	// wecToData is a map of workstatus-hosting WEC name to the
	// evaluation of the workstatus against the statuscollector's clauses.
	// The map contains entries for workstatuses that pass the statuscollector's
	// filter.
	wecToData map[string]*workStatusData
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

func (c *combinedStatusResolution) getName() string {
	c.RLock()
	defer c.RUnlock()

	return c.name
}

// setCollectionDestinations sets the collection destinations of the
// combinedstatus resolution.
// The given set is expected not to be mutated during and after this call.
// The function returns:
//   - removedDestinations: a boolean indicating if one or more destinations
//     were removed.
//   - a set of destinations that were added.
func (c *combinedStatusResolution) setCollectionDestinations(destinationsSet sets.Set[string]) (bool, sets.Set[string]) {
	c.Lock()
	defer c.Unlock()

	removedDestinations := c.collectionDestinations.Difference(destinationsSet)
	newDestinations := destinationsSet.Difference(c.collectionDestinations)
	// if nothing changed, return
	if len(removedDestinations) == 0 && len(newDestinations) == 0 {
		return false, nil
	}

	c.collectionDestinations = destinationsSet
	// trim the statuscollector data that are not relevant anymore
	for _, data := range c.statusCollectorNameToData {
		if data.wecToData == nil {
			continue
		}

		for clusterName := range removedDestinations {
			delete(data.wecToData, clusterName)
		}
	}

	return len(removedDestinations) > 0, newDestinations
}

// setStatusCollectors sets ALL the statuscollectors relevant to the
// combinedstatus resolution.
// The given map is expected not to be mutated during this call, its values
// contain valid expressions, and are immutable.
// The function returns a tuple (removedSome, addedSome):
//
// - removedSome: true if one or more statuscollectors were removed.
//
// - addedSome: true if one or more statuscollectors were added.
func (c *combinedStatusResolution) setStatusCollectors(statusCollectorNameToSpec map[string]v1alpha1.StatusCollectorSpec) (bool, bool) {
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

		if !statusCollectorSpecsMatch(statusCollectorData.StatusCollectorSpec, &statusCollectorSpec) {
			c.statusCollectorNameToData[statusCollectorName].StatusCollectorSpec = &statusCollectorSpec
			addedSome = true
		}
	}

	// add new statuscollector data
	for statusCollectorName, statusCollectorSpec := range statusCollectorNameToSpec {
		if _, ok := c.statusCollectorNameToData[statusCollectorName]; !ok {
			c.statusCollectorNameToData[statusCollectorName] = &statusCollectorData{
				StatusCollectorSpec: &statusCollectorSpec,
				wecToData:           make(map[string]*workStatusData),
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
		wecToData:           make(map[string]*workStatusData),
	}

	return true
}

// removeStatusCollector removes the status collector data from the
// combinedstatus resolution. The function returns true if the status collector
// data is removed.
func (c *combinedStatusResolution) removeStatusCollector(statusCollectorName string) bool {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.statusCollectorNameToData[statusCollectorName]; !ok {
		return false // statusCollector is irrelevant to this combinedstatus resolution
	}

	delete(c.statusCollectorNameToData, statusCollectorName)
	return true
}

func (c *combinedStatusResolution) isEmpty() bool {
	c.RLock()
	defer c.RUnlock()

	if len(c.statusCollectorNameToData) == 0 {
		return true
	}

	for _, scData := range c.statusCollectorNameToData {
		if len(scData.wecToData) > 0 {
			return false
		}
	}

	return true
}

// generateCombinedStatus calculates the combinedstatus from the statuscollector
// data in the combinedstatus resolution.
func (c *combinedStatusResolution) generateCombinedStatus(bindingName string,
	workloadObjectIdentifier util.ObjectIdentifier) *v1alpha1.CombinedStatus {
	c.RLock()
	defer c.RUnlock()

	combinedStatus := &v1alpha1.CombinedStatus{
		ObjectMeta: v1.ObjectMeta{
			Name:      c.name,
			Namespace: workloadObjectIdentifier.ObjectName.Namespace,
		},
		Results: make([]v1alpha1.NamedStatusCombination, 0, len(c.statusCollectorNameToData)),
	}

	for scName, scData := range c.statusCollectorNameToData {
		// the data has either select or combinedFields (with groupBy)
		if len(scData.Select) > 0 {
			namedStatusCombination := v1alpha1.NamedStatusCombination{
				Name:        scName,
				ColumnNames: make([]string, 0, len(scData.Select)+1), // for now, manually add wecName to all rows
				Rows:        make([]v1alpha1.StatusCombinationRow, 0, len(scData.wecToData)),
			}

			// add column names
			namedStatusCombination.ColumnNames = append(namedStatusCombination.ColumnNames, "wecName")
			for _, selectNamedExp := range scData.Select {
				namedStatusCombination.ColumnNames = append(namedStatusCombination.ColumnNames, selectNamedExp.Name)
			}

			// add rows for each workstatus
			for wecName, wsData := range scData.wecToData {
				row := v1alpha1.StatusCombinationRow{
					Columns: make([]v1alpha1.Value, 0, len(scData.Select)+1),
				}

				wecNameStr := wecName
				row.Columns = append(row.Columns, v1alpha1.Value{
					Type:   v1alpha1.TypeString,
					String: &wecNameStr})

				for _, selectNamedExp := range scData.Select {
					eval := wsData.selectEval[selectNamedExp.Name]
					if eval == nil {
						row.Columns = append(row.Columns, v1alpha1.Value{
							Type: v1alpha1.TypeNull})
						continue
					}

					evalValue := eval.Value()
					var col v1alpha1.Value
					// temporary until type checking is implemented
					switch v := evalValue.(type) {
					case string:
						col = v1alpha1.Value{
							Type:   v1alpha1.TypeString,
							String: &v,
						}
					case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
						numStr := fmt.Sprintf("%d", v)
						col = v1alpha1.Value{
							Type:   v1alpha1.TypeNumber,
							Number: &numStr,
						}
					case float32, float64:
						numStr := fmt.Sprintf("%f", v)
						col = v1alpha1.Value{
							Type:   v1alpha1.TypeNumber,
							Number: &numStr,
						}
					case bool:
						col = v1alpha1.Value{
							Type: v1alpha1.TypeBool,
							Bool: &v,
						}
					default:
						evalJSON, err := json.Marshal(evalValue)
						if err != nil {
							runtime2.HandleError(fmt.Errorf("failed to marshal select evaluation: %w", err))
							col = v1alpha1.Value{
								Type: v1alpha1.TypeNull,
							}
						} else {
							col = v1alpha1.Value{
								Type:   v1alpha1.TypeObject,
								Object: json.RawMessage(evalJSON),
							}
						}
					}

					row.Columns = append(row.Columns, col)
				}

				namedStatusCombination.Rows = append(namedStatusCombination.Rows, row)
			}

			combinedStatus.Results = append(combinedStatus.Results, namedStatusCombination)
			continue
		}

		// aggregation, TODO
	}

	return addLabelsToCombinedStatus(combinedStatus, bindingName, workloadObjectIdentifier)
}

func (c *combinedStatusResolution) compareCombinedStatus(status *v1alpha1.CombinedStatus) bool {
	return false //TODO
}

// evaluateWorkStatus evaluates the workstatus per all statuscollectors in the
// combinedstatus resolution.
//
// The function returns true if any statuscollector
// data is updated. If an evaluation fails, the function logs an error.
// if workStatusContent is nil, the function removes the workstatus data if it
// exists.
// TODO: handle errors
func (c *combinedStatusResolution) evaluateWorkStatus(celEvaluator *celEvaluator,
	workStatusWECName string, workStatusContent map[string]interface{}) bool {
	c.Lock()
	defer c.Unlock()

	if !c.collectionDestinations.Has(workStatusWECName) {
		return false // workstatus is not relevant to this combinedstatus resolution
	}

	updated := false
	for statusCollectorName, scData := range c.statusCollectorNameToData {
		changed, err := evaluateWorkStatusAgainstStatusCollectorWriteLocked(celEvaluator, workStatusWECName,
			workStatusContent, scData)
		if err != nil {
			runtime2.HandleError(fmt.Errorf("failed to evaluate workstatus against statuscollector (%s): %w",
				statusCollectorName, err))
			continue
		}

		updated = updated || changed
	}

	return updated
}

// evaluateWorkStatusAgainstStatusCollectorWriteLocked evaluates the workstatus against
// the statuscollector clauses and caches the evaluations. If the workstatus
// does not match the filter, no other clause is evaluated.
// The function returns true if an evaluation is updated.
// If any evaluation fails, the function returns an error.
// The function assumes that the caller holds a lock over the combinedstatus
// resolution.
func evaluateWorkStatusAgainstStatusCollectorWriteLocked(celEvaluator *celEvaluator, workStatusWECName string,
	workStatusContent map[string]interface{}, scData *statusCollectorData) (bool, error) {
	wsData, exists := scData.wecToData[workStatusWECName]

	if workStatusContent == nil && exists { // workstatus is empty/deleted, remove the workstatus data if it exists
		delete(scData.wecToData, workStatusWECName)
		return true, nil
	}

	// evaluate filter to determine if the workstatus is relevant
	if scData.Filter != nil {
		eval, err := celEvaluator.Evaluate(*scData.Filter, workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate filter expression: %w", err)
		}

		if eval.Type().TypeName() != "bool" {
			return false, fmt.Errorf("filter expression must evaluate to a boolean")
		}

		if !eval.Value().(bool) { // workstatus is not relevant
			if exists { // remove the workstatus data if it exists
				delete(scData.wecToData, workStatusWECName)
				return true, nil
			}

			return false, nil
		}
	}

	updated := false

	if !exists {
		wsData = &workStatusData{}
		scData.wecToData[workStatusWECName] = wsData
		updated = true
	}

	// evaluate select
	selectEvals := make(map[string]ref.Val)
	for _, selectNamedExp := range scData.Select {
		eval, err := celEvaluator.Evaluate(selectNamedExp.Def, workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate select expression: %w", err)
		}

		updated = updated || eval.Equal(wsData.selectEval[selectNamedExp.Name]).Value() != true
		selectEvals[selectNamedExp.Name] = eval
	}

	// evaluate groupBy
	groupByEvals := make(map[string]ref.Val)
	for _, groupByNamedExp := range scData.GroupBy {
		eval, err := celEvaluator.Evaluate(groupByNamedExp.Def, workStatusContent)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate groupBy expression: %w", err)
		}

		updated = updated || eval.Equal(wsData.groupByEval[groupByNamedExp.Name]).Value() != true
		groupByEvals[groupByNamedExp.Name] = eval
	}

	// evaluate combinedFields
	combinedFieldEvals := make(map[string]ref.Val)
	for _, combinedFieldNamedAgg := range scData.CombinedFields {
		if combinedFieldNamedAgg.Type == v1alpha1.AggregatorTypeCount {
			// count does not require a subject - mark the evaluation with nil
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

		updated = updated || eval.Equal(wsData.combinedFieldsEval[combinedFieldNamedAgg.Name]).Value() != true
		combinedFieldEvals[combinedFieldNamedAgg.Name] = eval
	}

	// update the workstatus data
	wsData.selectEval = selectEvals
	wsData.groupByEval = groupByEvals
	wsData.combinedFieldsEval = combinedFieldEvals

	return updated, nil
}

func addLabelsToCombinedStatus(combinedStatus *v1alpha1.CombinedStatus,
	bindingName string, workloadObjectIdentifier util.ObjectIdentifier) *v1alpha1.CombinedStatus {
	if combinedStatus.Labels == nil {
		combinedStatus.Labels = make(map[string]string)
	}
	// The CombinedStatus object has the following labels:
	// - "status.kubestellar.io/api-group" holding the API Group (not verison) of the workload object;
	// - "status.kubestellar.io/resource" holding the resource (lowercase plural) of the workload object;
	// - "status.kubestellar.io/namespace" holding the namespace of the workload object;
	// - "status.kubestellar.io/name" holding the name of the workload object;
	// - "status.kubestellar.io/binding-policy" holding the name of the BindingPolicy object.
	combinedStatus.Labels["status.kubestellar.io/api-group"] = workloadObjectIdentifier.GVR().Group
	combinedStatus.Labels["status.kubestellar.io/resource"] = workloadObjectIdentifier.GVR().Resource
	combinedStatus.Labels["status.kubestellar.io/namespace"] = workloadObjectIdentifier.ObjectName.Namespace
	combinedStatus.Labels["status.kubestellar.io/name"] = workloadObjectIdentifier.ObjectName.Name
	combinedStatus.Labels["status.kubestellar.io/binding-policy"] = bindingName // identical to binding-policy name

	return combinedStatus
}
