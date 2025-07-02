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
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/go-logr/logr"
	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	machtypes "k8s.io/apimachinery/pkg/types"
	runtime2 "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/abstract"
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

	// Name of the CombinedStatus object
	Name string
	// UID of the workload object
	workloadObjectUID machtypes.UID
	// StatusCollectorNameToData is a map of status collector names to
	// their corresponding data. This map has an entry for every relevant
	// StatusCollector name; the value is `nil` while the StatusCollector
	// does not exist.
	StatusCollectorNameToData map[string]*statusCollectorData
	// CollectionDestinations is a set of destinations that are expected to be
	// collected from.
	CollectionDestinations sets.Set[string]
}

var _ logr.Marshaler = &combinedStatusResolution{}

func (csr *combinedStatusResolution) MarshalLog() any {
	if csr == nil {
		return nil
	}
	return map[string]any{
		"Name":                      csr.Name,
		"WorkloadObjectUID":         string(csr.workloadObjectUID),
		"StatusCollectorNameToData": csr.StatusCollectorNameToData,
		"CollectionDestinations":    csr.CollectionDestinations.UnsortedList(),
	}
}

// statusCollectorData is a struct that represents the data of a
// statuscollector in a combinedstatus resolution.
// Making it the resolution of the tuple:
// (binding, object, statuscollector).
type statusCollectorData struct {
	// Never nil
	collectorSpec *v1alpha1.StatusCollectorSpec

	// wecToData is a map of workstatus-hosting WEC name to the
	// evaluation of the workstatus against the statuscollector's clauses.
	// The map contains entries for workstatuses that pass the statuscollector's
	// filter.
	wecToData map[string]*workStatusData
}

// workStatusData is a struct that represents the evaluation of a workstatus
// against a statuscollector's clauses.
// Making it the resolution of the tuple:
// (binding, object, statuscollector, workstatus).
type workStatusData struct {
	// groupByEval is a map of groupBy expression names to their evaluated values.
	groupByEval rowFragment
	// combinedFieldsEval is a map of combinedFields expression names to their
	// evaluated values. CombinedField types:
	// - COUNT: the number of workstatuses per groupBy value. If groupBy is
	//   empty, the count is the number of workstatuses.
	// - SUM: the sum of the values of the workstatuses groupBy values.
	// - AVG: the average of the values of the workstatuses groupBy values.
	// - MIN: the minimum value of the workstatuses groupBy values.
	// - MAX: the maximum value of the workstatuses groupBy values.
	// If a combinedField's eval is nil, its TYPE is COUNT.
	combinedFieldsEval rowFragment

	selectEval rowFragment

	evalErrors []v1alpha1.ErrorInColumn
}

// rowFragment is a map from column name to value
type rowFragment = map[string]ref.Val

func (c *combinedStatusResolution) getName() string {
	c.RLock()
	defer c.RUnlock()

	return c.Name
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

	removedDestinations := c.CollectionDestinations.Difference(destinationsSet)
	newDestinations := destinationsSet.Difference(c.CollectionDestinations)
	// if nothing changed, return
	if len(removedDestinations) == 0 && len(newDestinations) == 0 {
		return false, nil
	}

	c.CollectionDestinations = destinationsSet
	// trim the statuscollector data that are not relevant anymore
	for _, data := range c.StatusCollectorNameToData {
		if data == nil || len(data.wecToData) == 0 {
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
// The given map is expected not to be mutated during this call,
// and the specs in it are immutable.
// The given map has an entry for every relevant StatusCollector name,
// but a `nil` spec pointer if the collector does not exist now.
// The function returns a tuple (removedSome, addedSome):
//
// - removedSome: true if one or more statuscollectors were removed.
//
// - addedSome: true if one or more statuscollectors were added.
func (c *combinedStatusResolution) setStatusCollectors(statusCollectorNameToSpec map[string]*v1alpha1.StatusCollectorSpec) (bool, bool) {
	c.Lock()
	defer c.Unlock()

	removedSome, addedSome := false, false

	// remove entries for collectors that are not relevant anymore and update the
	// statuscollector data that are. If one of the latter is updated, mark it as added
	for statusCollectorName, statusCollectorData := range c.StatusCollectorNameToData {
		statusCollectorSpec, ok := statusCollectorNameToSpec[statusCollectorName]
		if !ok {
			delete(c.StatusCollectorNameToData, statusCollectorName)
			removedSome = true
			continue
		}

		if statusCollectorSpec != nil && (statusCollectorData == nil || !statusCollectorSpecsMatch(statusCollectorData.collectorSpec, statusCollectorSpec)) {
			c.StatusCollectorNameToData[statusCollectorName].collectorSpec = statusCollectorSpec
			addedSome = true
		}
	}

	// add new data for newly relevant StatusCollectors
	for statusCollectorName, statusCollectorSpec := range statusCollectorNameToSpec {
		if _, ok := c.StatusCollectorNameToData[statusCollectorName]; !ok {
			if statusCollectorSpec == nil {
				c.StatusCollectorNameToData[statusCollectorName] = nil
			} else {
				c.StatusCollectorNameToData[statusCollectorName] = &statusCollectorData{
					collectorSpec: statusCollectorSpec,
					wecToData:     make(map[string]*workStatusData),
				}
			}

			addedSome = true
		}
	}

	return removedSome, addedSome
}

// updateStatusCollector updates the status collector data in the
// combinedstatus resolution. If the status collector is not relevant to the
// latter, the function returns false. The function returns true if the status
// collector data is updated. The given spec pointer is not nil and
// the spec is assumed to be valid and immutable.
func (c *combinedStatusResolution) updateStatusCollector(statusCollectorName string,
	statusCollectorSpec *v1alpha1.StatusCollectorSpec) bool {
	c.Lock()
	defer c.Unlock()

	scData, ok := c.StatusCollectorNameToData[statusCollectorName]
	if !ok {
		return false // statusCollector is irrelevant to this combinedstatus resolution
	}

	if scData != nil && statusCollectorSpecsMatch(scData.collectorSpec, statusCollectorSpec) {
		return false // statusCollector data is already up-to-date
	}

	// status collector clauses need to be updated, therefore update fields
	// and invalidate all cached workstatus evaluations by resetting the map
	c.StatusCollectorNameToData[statusCollectorName] = &statusCollectorData{
		collectorSpec: statusCollectorSpec,
		wecToData:     make(map[string]*workStatusData),
	}

	return true
}

// noteStatusCollectorAbsence updates the resolution, if it does not already do so,
// to recognize that the given StatusCollector does not exist.
// Returns true if there was a change.
func (c *combinedStatusResolution) noteStatusCollectorAbsence(statusCollectorName string) bool {
	c.Lock()
	defer c.Unlock()

	if oval := c.StatusCollectorNameToData[statusCollectorName]; oval == nil {
		return false // statusCollector is irrelevant to this combinedstatus resolution
	}

	c.StatusCollectorNameToData[statusCollectorName] = nil
	return true
}

// generateCombinedStatus calculates the combinedstatus from the statuscollector
// data in the combinedstatus resolution.
func (c *combinedStatusResolution) generateCombinedStatus(bindingName string,
	workloadObjectIdentifier util.ObjectIdentifier) *v1alpha1.CombinedStatus {
	c.RLock()
	defer c.RUnlock()

	combinedStatus := &v1alpha1.CombinedStatus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: workloadObjectIdentifier.ObjectName.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: workloadObjectIdentifier.GVK.GroupVersion().String(),
				Kind:       workloadObjectIdentifier.GVK.Kind,
				Name:       workloadObjectIdentifier.ObjectName.Name,
				UID:        c.workloadObjectUID,
			}},
		},
		Results: make([]v1alpha1.NamedStatusCombination, 0, len(c.StatusCollectorNameToData)),
	}

	if combinedStatus.Namespace == metav1.NamespaceNone {
		combinedStatus.Namespace = util.ClusterScopedObjectsCombinedStatusNamespace
	}

	for _, scName := range sortedStringSlice(abstract.PrimitiveMapKeySlice(c.StatusCollectorNameToData)) {
		scData := c.StatusCollectorNameToData[scName]
		if scData == nil {
			continue
		}
		// the data, if not nil, has either select or combinedFields (with groupBy)
		if len(scData.collectorSpec.Select) > 0 {
			combinedStatus.Results = append(combinedStatus.Results, *handleSelectReadLocked(scName, scData))
			continue
		}
		combinedStatus.Results = append(combinedStatus.Results, *handleAggregationReadLocked(scName, scData))
	}

	return addLabelsToCombinedStatus(combinedStatus, bindingName, workloadObjectIdentifier)
}

func (c *combinedStatusResolution) compareCombinedStatus(status *v1alpha1.CombinedStatus,
	bindingName string, sourceObjectIdentifier util.ObjectIdentifier) *v1alpha1.CombinedStatus {
	c.RLock()
	defer c.RUnlock()

	localCombinedStatus := c.generateCombinedStatus(bindingName, sourceObjectIdentifier)

	// check labels
	if !validateCombinedStatusLabels(status, bindingName, sourceObjectIdentifier) {
		return localCombinedStatus
	}

	if len(localCombinedStatus.Results) != len(status.Results) {
		return localCombinedStatus
	}
	// make map of name -> content and check if all mapped. This is sufficient since size matches,
	// and the names are unique due to being that of cluster-wide resource name (status collector).
	localResultsMap := abstract.SliceToPrimitiveMap(localCombinedStatus.Results,
		func(namedStatusCombination v1alpha1.NamedStatusCombination) string {
			return namedStatusCombination.Name
		},
		func(namedStatusCombination v1alpha1.NamedStatusCombination) v1alpha1.NamedStatusCombination {
			return namedStatusCombination
		})

	for _, statusCombination := range status.Results {
		if _, ok := localResultsMap[statusCombination.Name]; !ok {
			return localCombinedStatus
		}

		localStatusCombination := localResultsMap[statusCombination.Name]
		if !statusCombinationEqual(&statusCombination, &localStatusCombination) {
			return localCombinedStatus
		}
	}

	return nil
}

// evaluateWorkStatus evaluates the workstatus per all statuscollectors in the
// combinedstatus resolution.
//
// The function returns true if any statuscollector
// data is updated. If an evaluation fails, the function logs an error.
// if workStatusContent is nil, the function removes the workstatus data if it
// exists.
// TODO: handle errors
func (c *combinedStatusResolution) evaluateWorkStatus(ctx context.Context, celEvaluator *celEvaluator,
	workStatusWECName string, content map[string]interface{}) bool {
	c.Lock()
	defer c.Unlock()
	logger := klog.FromContext(ctx)

	if !c.CollectionDestinations.Has(workStatusWECName) {
		logger.V(5).Info("WEC is not status-collected", "wecName", workStatusWECName)
		return false // workstatus is not relevant to this combinedstatus resolution
	}

	updated := false
	for _, scData := range c.StatusCollectorNameToData {
		if scData == nil {
			continue
		}
		changed := evaluateWorkStatusAgainstStatusCollectorWriteLocked(celEvaluator, workStatusWECName,
			content, scData)
		updated = updated || changed
	}
	logger.V(5).Info("Evaluated collectors", "wecName", workStatusWECName, "numCollectors", len(c.StatusCollectorNameToData), "updated", updated)

	return updated
}

// queryingContentRequirements determines which objects are required for
// querying within the CEL expressions of the status collectors.
// The function returns a tuple of booleans:
//   - `sourceObjectKey` required
//   - `returnedKey` required
//   - `inventoryKey` required
//   - `propagationMetaKey` required
func (c *combinedStatusResolution) queryingContentRequirements() (bool, bool, bool, bool) {
	c.RLock()
	defer c.RUnlock()

	sourceObjectKeyRequired, returnedKeyRequired,
		inventoryKeyRequired, propagationMetaKeyRequired := false, false, false, false

	pred := func(expr *string) (bool, bool, bool, bool) {
		return objectIsQueried(expr, sourceObjectKey), objectIsQueried(expr, returnedKey),
			objectIsQueried(expr, inventoryKey), objectIsQueried(expr, propagationMetaKey)
	}

	mergeBooleans := func(s, r, i, p bool) {
		// outside variables are referenced in closures https://go.dev/tour/moretypes/25
		sourceObjectKeyRequired = sourceObjectKeyRequired || s
		returnedKeyRequired = returnedKeyRequired || r
		inventoryKeyRequired = inventoryKeyRequired || i
		propagationMetaKeyRequired = propagationMetaKeyRequired || p
	}

	for _, scData := range c.StatusCollectorNameToData {
		if scData == nil {
			continue
		}
		if scData.collectorSpec.Filter != nil {
			mergeBooleans(pred((*string)(scData.collectorSpec.Filter)))
		}

		for _, selectNamedExp := range scData.collectorSpec.Select {
			mergeBooleans(pred((*string)(&selectNamedExp.Def)))
		}

		for _, combinedFieldNamedAgg := range scData.collectorSpec.CombinedFields {
			if combinedFieldNamedAgg.Subject != nil {
				mergeBooleans(pred((*string)(combinedFieldNamedAgg.Subject)))
			}
		}

		for _, groupByNamedExp := range scData.collectorSpec.GroupBy {
			mergeBooleans(pred((*string)(&groupByNamedExp.Def)))
		}
	}

	return sourceObjectKeyRequired, returnedKeyRequired, inventoryKeyRequired, propagationMetaKeyRequired
}

// evaluateWorkStatusAgainstStatusCollectorWriteLocked evaluates the workstatus against
// the statuscollector clauses and caches the evaluations. If the workstatus
// does not match the filter, no other clause is evaluated.
// The function returns true if an evaluation is updated.
// If any evaluation fails, the function returns an error.
// The function assumes that the caller holds a lock over the combinedstatus
// resolution.
func evaluateWorkStatusAgainstStatusCollectorWriteLocked(celEvaluator *celEvaluator, workStatusWECName string,
	content map[string]interface{}, scData *statusCollectorData) bool {
	wsData, exists := scData.wecToData[workStatusWECName]

	if content == nil { // workstatus is empty/deleted, remove the workstatus data if it exists
		delete(scData.wecToData, workStatusWECName)
		return exists
	}
	var evalErrors []v1alpha1.ErrorInColumn

	// evaluate filter to determine if the workstatus is relevant
	if scData.collectorSpec.Filter != nil {
		eval, err := celEvaluator.Evaluate(*scData.collectorSpec.Filter, content)
		if err != nil {
			evalErrors = append(evalErrors,
				v1alpha1.ErrorInColumn{ColumnName: v1alpha1.FilterColumnName, Error: err.Error()})
		} else if tn := eval.Type().TypeName(); tn != "bool" && tn != "null" {
			evalErrors = append(evalErrors,
				v1alpha1.ErrorInColumn{ColumnName: v1alpha1.FilterColumnName,
					Error: fmt.Sprintf("filter expression has type %s but expected bool or null", tn)})
		} else {
			if tn == "bool" && !eval.Value().(bool) { // workstatus is not relevant
				if exists { // remove the workstatus data if it exists
					delete(scData.wecToData, workStatusWECName)
					return true
				}
				return false
			}
		}
	}

	updated := false

	if !exists {
		wsData = &workStatusData{
			groupByEval:        make(map[string]ref.Val),
			combinedFieldsEval: make(map[string]ref.Val),
			selectEval:         make(map[string]ref.Val),
		}
		scData.wecToData[workStatusWECName] = wsData
		updated = true
	}

	// evaluate select
	selectEvals := make(map[string]ref.Val)
	for _, selectNamedExp := range scData.collectorSpec.Select {
		eval, err := celEvaluator.Evaluate(selectNamedExp.Def, content)
		if err != nil {
			evalErrors = append(evalErrors, v1alpha1.ErrorInColumn{ColumnName: selectNamedExp.Name, Error: err.Error()})
			eval = celtypes.DefaultTypeAdapter.NativeToValue(err.Error())
		}

		prevVal, exists := wsData.selectEval[selectNamedExp.Name]
		if eval != nil {
			updated = updated || !exists || eval.Equal(prevVal).Value() != true
		} else {
			updated = updated || exists
		}

		selectEvals[selectNamedExp.Name] = eval
	}

	// evaluate groupBy
	groupByEvals := make(map[string]ref.Val)
	for _, groupByNamedExp := range scData.collectorSpec.GroupBy {
		eval, err := celEvaluator.Evaluate(groupByNamedExp.Def, content)
		if err != nil {
			evalErrors = append(evalErrors, v1alpha1.ErrorInColumn{ColumnName: groupByNamedExp.Name, Error: err.Error()})
			eval = celtypes.DefaultTypeAdapter.NativeToValue(err.Error())
		}

		prevVal, exists := wsData.selectEval[groupByNamedExp.Name]
		if eval != nil {
			updated = updated || !exists || eval.Equal(prevVal).Value() != true
		} else {
			updated = updated || exists
		}

		groupByEvals[groupByNamedExp.Name] = eval
	}

	// evaluate combinedFields
	combinedFieldEvals := make(map[string]ref.Val)
	for _, combinedFieldNamedAgg := range scData.collectorSpec.CombinedFields {
		if combinedFieldNamedAgg.Type == v1alpha1.AggregatorTypeCount {
			// count does not require a subject - mark the evaluation with nil
			currentEval, exists := wsData.combinedFieldsEval[combinedFieldNamedAgg.Name]
			updated = updated || !exists || currentEval != nil
			combinedFieldEvals[combinedFieldNamedAgg.Name] = nil
			continue
		}

		// evaluate subject which should not be nil since the statuscollector is valid
		eval, err := celEvaluator.Evaluate(*combinedFieldNamedAgg.Subject, content)
		if err != nil {
			eval = celtypes.DefaultTypeAdapter.NativeToValue(err.Error())
			evalErrors = append(evalErrors, v1alpha1.ErrorInColumn{Error: err.Error(),
				ColumnName: combinedFieldNamedAgg.Name,
			})
		}

		prevVal, exists := wsData.selectEval[combinedFieldNamedAgg.Name]
		if eval != nil {
			updated = updated || !exists || eval.Equal(prevVal).Value() != true
		} else {
			updated = updated || exists
		}

		combinedFieldEvals[combinedFieldNamedAgg.Name] = eval
	}

	// update the workstatus data
	wsData.selectEval = selectEvals
	wsData.groupByEval = groupByEvals
	wsData.combinedFieldsEval = combinedFieldEvals
	wsData.evalErrors = evalErrors
	return updated
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

func validateCombinedStatusLabels(combinedStatus *v1alpha1.CombinedStatus,
	bindingName string, workloadObjectIdentifier util.ObjectIdentifier) bool {
	if len(combinedStatus.Labels) != 5 {
		return false
	}

	if combinedStatus.Labels["status.kubestellar.io/api-group"] != workloadObjectIdentifier.GVR().Group {
		return false
	}

	if combinedStatus.Labels["status.kubestellar.io/resource"] != workloadObjectIdentifier.GVR().Resource {
		return false
	}

	if combinedStatus.Labels["status.kubestellar.io/namespace"] != workloadObjectIdentifier.ObjectName.Namespace {
		return false
	}

	if combinedStatus.Labels["status.kubestellar.io/name"] != workloadObjectIdentifier.ObjectName.Name {
		return false
	}

	if combinedStatus.Labels["status.kubestellar.io/binding-policy"] != bindingName {
		return false
	}

	return true
}

// handleSelectReadLocked handles the select expressions of a statuscollector
// data. This means that the function evaluates the select expressions against
// the possibly filtered workstatuses and returns the result in a
// NamedStatusCombination.
func handleSelectReadLocked(scName string, scData *statusCollectorData) *v1alpha1.NamedStatusCombination {
	namedStatusCombination := v1alpha1.NamedStatusCombination{
		Name:        scName,
		ColumnNames: make([]string, 0, len(scData.collectorSpec.Select)),
		Rows:        make([]v1alpha1.StatusCombinationRow, 0, len(scData.wecToData)),
	}

	// add column names
	namedStatusCombination.ColumnNames = append(namedStatusCombination.ColumnNames,
		abstract.SliceMap(scData.collectorSpec.Select, func(selectNamedExp v1alpha1.NamedExpression) string {
			return selectNamedExp.Name
		})...)

	// add rows for each workstatus
	for _, wsData := range scData.wecToData {
		row := v1alpha1.StatusCombinationRow{
			Columns: make([]v1alpha1.Value, 0, len(scData.collectorSpec.Select)),
		}

		for _, selectNamedExp := range scData.collectorSpec.Select {
			row.Columns = append(row.Columns, refValToValue(wsData.selectEval[selectNamedExp.Name]))
		}

		namedStatusCombination.Rows = append(namedStatusCombination.Rows, row)
	}

	return &namedStatusCombination
}

func refValToValue(val ref.Val) v1alpha1.Value {
	if val == nil {
		return v1alpha1.Value{Type: v1alpha1.TypeNull}
	}

	valValue := val.Value()
	// temporary until type checking is implemented
	switch v := valValue.(type) {
	case string:
		return v1alpha1.Value{
			Type:   v1alpha1.TypeString,
			String: &v,
		}
	case int:
		numStr := strconv.FormatInt(int64(v), 10)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case int8:
		numStr := strconv.FormatInt(int64(v), 10)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case int16:
		numStr := strconv.FormatInt(int64(v), 10)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case int32:
		numStr := strconv.FormatInt(int64(v), 10)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case int64:
		numStr := strconv.FormatInt(v, 10)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case uint:
		numStr := strconv.FormatUint(uint64(v), 10)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case uint8:
		numStr := strconv.FormatUint(uint64(v), 10)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case uint16:
		numStr := strconv.FormatUint(uint64(v), 10)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case uint32:
		numStr := strconv.FormatUint(uint64(v), 10)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case uint64:
		numStr := strconv.FormatUint(v, 10)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case float32:
		numStr := strconv.FormatFloat(float64(v), 'g', -1, 64)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case float64:
		numStr := strconv.FormatFloat(v, 'g', -1, 64)
		return v1alpha1.Value{
			Type:   v1alpha1.TypeNumber,
			Number: &numStr,
		}
	case bool:
		return v1alpha1.Value{
			Type: v1alpha1.TypeBool,
			Bool: &v,
		}
	default:
		evalJSON, err := json.Marshal(valValue)
		if err != nil {
			runtime2.HandleError(fmt.Errorf("failed to marshal select evaluation: %w", err))
			return v1alpha1.Value{
				Type: v1alpha1.TypeNull,
			}
		} else {
			return v1alpha1.Value{
				Type:   v1alpha1.TypeObject,
				Object: &extv1.JSON{Raw: evalJSON},
			}
		}
	}
}

// handleAggregationReadLocked handles the aggregation of the statuscollector
// data. This means that the function evaluates the groupBy expressions and the
// combinedFields expressions against the possibly filtered workstatuses and
// returns the result in a NamedStatusCombination.
//
// If there is no groupBy, the function treats all workstatuses as a single
// group.
func handleAggregationReadLocked(scName string, scData *statusCollectorData) *v1alpha1.NamedStatusCombination {
	// The aggregation requires grouping workstatuses by tuples of groupBy values,
	// where for N groupBy expressions, a group key would be a tuple of N values.
	// To achieve this grouping, we maintain two maps:
	// 1. groupBy Evaluation Value -> unique ID
	// 2. N-values tuple -> []*workStatusData
	// 		Where the N-values tuple is the string concatenation of the assigned numbers
	// 		seperated by commas, ordered by the groupBy expressions in scData.GroupBy slice.
	generator := 0
	ValueToNumber := map[any]int{}
	idToAggregationGroup := map[string]*aggregationGroup{}

	if len(scData.collectorSpec.GroupBy) == 0 && len(scData.wecToData) == 0 {
		// len(scData.GroupBy) == 0 means there is exactly one group to aggregate over.
		// len(scData.wecToData) == 0 means that the loop below will not put the group in the map.
		idToAggregationGroup[""] = &aggregationGroup{GroupBy: map[string]ref.Val{},
			Rows: map[v1alpha1.Destination]rowFragment{}}
	}
	rowErrors := []v1alpha1.RowEvaluationError{}
	coveredColumns := sets.New[string]()
	for wecName, wsData := range scData.wecToData {
		if len(wsData.evalErrors) > 0 {
			for _, errIC := range wsData.evalErrors {
				if coveredColumns.Has(errIC.ColumnName) {
					continue
				}
				coveredColumns.Insert(errIC.ColumnName)
				rowErrors = append(rowErrors, v1alpha1.RowEvaluationError{
					WEC:        v1alpha1.Destination{ClusterId: wecName},
					ColumnName: errIC.ColumnName,
					Error:      errIC.Error})
			}
			continue // exclude WECs that have evaluation errors
		}
		valuesTuple := make([]string, 0, len(scData.collectorSpec.GroupBy))
		for _, groupByNamedExp := range scData.collectorSpec.GroupBy {
			groupByValue := wsData.groupByEval[groupByNamedExp.Name]

			// ensure unique number mapping for the value
			uid, exists := ValueToNumber[groupByValue.Value()]
			if !exists {
				ValueToNumber[groupByValue.Value()] = generator
				uid = generator
				generator++
			}

			valuesTuple = append(valuesTuple, strconv.Itoa(uid))
		}

		key := strings.Join(valuesTuple, ",")
		ag := idToAggregationGroup[key]
		if ag == nil {
			ag = &aggregationGroup{GroupBy: wsData.groupByEval, Rows: map[v1alpha1.Destination]rowFragment{}}
			idToAggregationGroup[key] = ag
		}
		ag.Rows[v1alpha1.Destination{ClusterId: wecName}] = wsData.combinedFieldsEval
	}

	// calculate the combinedFields for each group in one table
	return calculateCombinedResult(idToAggregationGroup, scName, scData, rowErrors)
}

type aggregationGroup struct {
	// GroupBy holds the unique tuple of "GROUP BY" values that this group is associated with
	GroupBy map[string]ref.Val

	// Rows holds the rows to aggregate, as a map from Destination to rowFragment
	Rows map[v1alpha1.Destination]rowFragment
}

// calculateCombinedResult calculates the combinedFields for each group in the
// idToAggregationGroup map and returns the result in a
// NamedStatusCombination.
func calculateCombinedResult(idToAggregationGroup map[string]*aggregationGroup,
	statusCollectorName string, statusCollectorData *statusCollectorData,
	rowErrors []v1alpha1.RowEvaluationError) *v1alpha1.NamedStatusCombination {
	// create the named status combination
	namedStatusCombination := v1alpha1.NamedStatusCombination{
		Name:        statusCollectorName,
		ColumnNames: make([]string, 0, len(statusCollectorData.collectorSpec.CombinedFields)),
		Rows:        []v1alpha1.StatusCombinationRow{},
		RowErrors:   rowErrors,
	}

	// add column names: one per groupBy expression and one per combinedField
	namedStatusCombination.ColumnNames = append(namedStatusCombination.ColumnNames,
		abstract.SliceMap(statusCollectorData.collectorSpec.GroupBy,
			func(groupByNamedExp v1alpha1.NamedExpression) string {
				return groupByNamedExp.Name
			})...)
	namedStatusCombination.ColumnNames = append(namedStatusCombination.ColumnNames,
		abstract.SliceMap(statusCollectorData.collectorSpec.CombinedFields,
			func(combinedFieldNamedAgg v1alpha1.NamedAggregator) string {
				return combinedFieldNamedAgg.Name
			})...)

	// handle combinedFields (named aggregators) per group
	for _, ag := range idToAggregationGroup {
		row := v1alpha1.StatusCombinationRow{
			Columns: make([]v1alpha1.Value, 0,
				len(statusCollectorData.collectorSpec.GroupBy)+len(statusCollectorData.collectorSpec.CombinedFields)),
		}

		// fill groupBy values using one of the workstatuses in the group
		if len(statusCollectorData.collectorSpec.GroupBy) > 0 {
			for _, groupByNamedExp := range statusCollectorData.collectorSpec.GroupBy {
				groupByValue := ag.GroupBy[groupByNamedExp.Name]
				row.Columns = append(row.Columns, refValToValue(groupByValue))
			}
		}

		// add combinedFields
		for _, combinedFieldNamedAgg := range statusCollectorData.collectorSpec.CombinedFields {
			aggregation, aggErr := calculateCombinedFieldAggregation(combinedFieldNamedAgg, ag.Rows)
			row.Columns = append(row.Columns, aggregation)
			if aggErr != "" {
				namedStatusCombination.AggregationErrors = append(namedStatusCombination.AggregationErrors,
					v1alpha1.ErrorInColumn{ColumnName: combinedFieldNamedAgg.Name, Error: aggErr})
			}
		}

		namedStatusCombination.Rows = append(namedStatusCombination.Rows, row)
	}

	return &namedStatusCombination
}

func calculateCombinedFieldAggregation(combinedFieldNamedAgg v1alpha1.NamedAggregator,
	rows map[v1alpha1.Destination]rowFragment) (v1alpha1.Value, string) {
	var numStr, errStr string

	switch combinedFieldNamedAgg.Type {
	case v1alpha1.AggregatorTypeCount:
		numStr = strconv.Itoa(len(rows))
	case v1alpha1.AggregatorTypeSum:
		sum := 0.0
		for dest, row := range rows {
			subject, err1 := getCombinedFieldSubject(combinedFieldNamedAgg, row)
			if err1 != "" {
				if errStr == "" {
					errStr = fmt.Sprintf("for WEC %s, %s", dest.ClusterId, err1)
				}
			}
			if subject == nil {
				continue
			}
			sum += *subject
		}
		numStr = strconv.FormatFloat(sum, 'g', -1, 64)
	case v1alpha1.AggregatorTypeAvg:
		var avg float64 = math.NaN()
		if count := len(rows); count > 0 {
			sum := 0.0
			for dest, row := range rows {
				subject, err1 := getCombinedFieldSubject(combinedFieldNamedAgg, row)
				if err1 != "" {
					if errStr == "" {
						errStr = fmt.Sprintf("for WEC %s, %s", dest.ClusterId, err1)
					}
				}
				if subject == nil {
					continue
				}
				sum += *subject
			}
			avg = sum / float64(count)
		} else {
			errStr = "no values to average"
		}
		numStr = strconv.FormatFloat(avg, 'g', -1, 64)
	case v1alpha1.AggregatorTypeMin:
		min := math.Inf(1)
		for dest, row := range rows {
			subject, err1 := getCombinedFieldSubject(combinedFieldNamedAgg, row)
			if err1 != "" {
				if errStr == "" {
					errStr = fmt.Sprintf("for WEC %s, %s", dest.ClusterId, err1)
				}
			}
			if subject == nil {
				continue
			}
			if *subject < min {
				min = *subject
			}
		}
		numStr = strconv.FormatFloat(min, 'g', -1, 64)
	case v1alpha1.AggregatorTypeMax:
		max := math.Inf(-1)
		for dest, row := range rows {
			subject, err1 := getCombinedFieldSubject(combinedFieldNamedAgg, row)
			if err1 != "" {
				if errStr == "" {
					errStr = fmt.Sprintf("for WEC %s, %s", dest.ClusterId, err1)
				}
			}
			if subject == nil {
				continue
			}
			if *subject > max {
				max = *subject
			}
		}
		numStr = strconv.FormatFloat(max, 'g', -1, 64)
	default:
		return v1alpha1.Value{
			Type: v1alpha1.TypeNull,
		}, fmt.Sprintf("unsupported aggregation type %s", combinedFieldNamedAgg.Type)
	}

	return v1alpha1.Value{
		Type:   v1alpha1.TypeNumber,
		Number: &numStr,
	}, errStr
}

// getCombinedFieldSubject returns the subject of the combinedField evaluation.
// If the subject does not conform to the expected type, the function logs an
// error and returns nil. TODO: handle errors
// The given `row` map has domain = name of a NamedAggregator,
// domain = evaluated value (before aggregation, of course) for a given WEC.
func getCombinedFieldSubject(combinedFieldNamedAgg v1alpha1.NamedAggregator, row map[string]ref.Val) (*float64, string) {
	// NamedAggregator pairs a name with a way to aggregate over some objects.
	//
	// - For `type=="COUNT"`, `subject` is omitted and the aggregate is the count
	// of those objects that are not `null`.
	//
	// - For the other types, `subject` is required and SHOULD
	// evaluate to a numeric value; exceptions are handled in the following way.
	// For a string value: if it parses as an int64 or float64 then that is used.
	// Otherwise this is an error condition: a value of 0 is used, and the error
	// is reported in the BindingPolicyStatus.Errors (not necessarily repeated for each WEC).
	eval := row[combinedFieldNamedAgg.Name]
	if eval == nil {
		return nil, ""
	}

	evalValue := eval.Value()
	switch v := evalValue.(type) {
	case int:
		f := float64(v)
		return &f, ""
	case int8:
		f := float64(v)
		return &f, ""
	case int16:
		f := float64(v)
		return &f, ""
	case int32:
		f := float64(v)
		return &f, ""
	case int64:
		f := float64(v)
		return &f, ""
	case uint:
		f := float64(v)
		return &f, ""
	case uint8:
		f := float64(v)
		return &f, ""
	case uint16:
		f := float64(v)
		return &f, ""
	case uint32:
		f := float64(v)
		return &f, ""
	case uint64:
		f := float64(v)
		return &f, ""
	case float32:
		f := float64(v)
		return &f, ""
	case float64:
		f := v
		return &f, ""
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, "failed to parse combinedField subject as a float: " + err.Error()
		}
		return &f, ""
	default:
		return nil, fmt.Sprintf("combinedField subject has unexpected type %T", evalValue)
	}
}

func statusCombinationEqual(a, b *v1alpha1.NamedStatusCombination) bool {
	if a.Name != b.Name {
		return false
	}

	if len(a.ColumnNames) != len(b.ColumnNames) {
		return false
	}

	for i := range a.ColumnNames {
		if a.ColumnNames[i] != b.ColumnNames[i] {
			return false
		}
	}

	if len(a.Rows) != len(b.Rows) {
		return false
	}

	// rows may noy be named, check rows one by one across all columns
	for _, aRow := range a.Rows {
		// check if identical to at least one b row
		found := false
		for _, bRow := range b.Rows {
			if statusCombinationRowEqual(&aRow, &bRow) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func statusCombinationRowEqual(a, b *v1alpha1.StatusCombinationRow) bool {
	if len(a.Columns) != len(b.Columns) {
		return false
	}

	for i := range a.Columns {
		if !valueEqual(&a.Columns[i], &b.Columns[i]) {
			return false
		}
	}

	return true
}

func valueEqual(a, b *v1alpha1.Value) bool {
	if a.Type != b.Type {
		return false
	}

	if !validateValue(a) || !validateValue(b) {
		return false
	}

	switch a.Type {
	case v1alpha1.TypeString:
		return *a.String == *b.String
	case v1alpha1.TypeNumber:
		aNum, err := parseNumber(*a.Number)
		if err != nil {
			return false
		}

		bNum, err := parseNumber(*b.Number)
		if err != nil {
			return false
		}

		return numericEqual(aNum, bNum)
	case v1alpha1.TypeBool:
		return *a.Bool == *b.Bool
	case v1alpha1.TypeObject:
		var v1, v2 interface{}
		json.Unmarshal([]byte(a.Object.Raw), &v1)
		json.Unmarshal([]byte(b.Object.Raw), &v2)
		return reflect.DeepEqual(v1, v2)
	case v1alpha1.TypeNull:
		return true
	default:
		return false
	}
}

func validateValue(value *v1alpha1.Value) bool {
	switch value.Type {
	case v1alpha1.TypeString:
		return value.String != nil
	case v1alpha1.TypeNumber:
		return value.Number != nil
	case v1alpha1.TypeBool:
		return value.Bool != nil
	case v1alpha1.TypeObject:
		return value.Object != nil
	case v1alpha1.TypeNull:
		return true
	default:
		return false
	}
}

func parseNumber(s string) (any, error) {
	if intValue, err := strconv.ParseInt(s, 10, 64); err == nil {
		return intValue, nil
	}

	if uintValue, err := strconv.ParseUint(s, 10, 64); err == nil {
		return uintValue, nil
	}

	if floatValue, err := strconv.ParseFloat(s, 64); err == nil {
		return floatValue, nil
	}

	return nil, fmt.Errorf("failed to parse number")
}

func numericEqual(a, b interface{}) bool {
	switch aValue := a.(type) {
	case int64:
		if bValue, ok := b.(int64); ok {
			return aValue == bValue
		}
	case uint64:
		if bValue, ok := b.(uint64); ok {
			return aValue == bValue
		}
	case float64:
		if bValue, ok := b.(float64); ok {
			return aValue == bValue
		}
	}
	return false
}

func sortedStringSlice(s []string) []string {
	sort.Sort(sort.StringSlice(s))
	return s
}

// objectIsQueried checks whether `obj` appears in `query` as a standalone
// occurrence.
// That means, it must not be surrounded by alphanumeric characters.
// query is assumed to be non-nil.
func objectIsQueried(query *string, obj string) bool {
	idx := 0

	for {
		idx = strings.Index((*query)[idx:], obj) // slices share the same storage, therefore efficient
		if idx == -1 {
			return false
		}

		if isWholeWord(query, idx, len(obj)) {
			return true
		}

		idx += len(obj)
	}
}

// isWholeWord checks whether the word at `idx` in `s` is a whole word.
// s is assumed to be non-nil.
func isWholeWord(s *string, idx, length int) bool {
	// check preceding rune
	if idx > 0 {
		r := rune((*s)[idx-1])

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}

	// check following rune
	if idx+length < len(*s) {
		r := rune((*s)[idx+length])

		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}

	return true
}
