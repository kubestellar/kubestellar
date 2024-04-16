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

package jsonpath

import (
	k8sreflect "k8s.io/apimachinery/third_party/forked/golang/reflect"
)

// This file implements JSONPath querying almost as described in RFC 9535.
// This file does not implement the function extensions.
// The only comparisons implemented here are equality and inequality.
// The order in which result nodes are produced is not consistent with the RFC.

// JSONValue is something that can be produced by encoding/json.Unmarshal into a pointer
// to a nil `any`.
// That is: `bool`, `float64`, `string`, `nil`, `[]any`, or `map[string]any` --- where those
// nested `any` have the same restriction.
type JSONValue = any

// Node is a JSON document node.
// The other methods MUST NOT be invoked after Remove is invoked.
type Node interface {
	Get() JSONValue
	Set(JSONValue)
	Remove()
}

// RootNode is the Node implementation to use for the document's root node.
// Set `*Value` to the document.
// Removing this node amounts to setting `Value` to `nil`.
type RootNode struct {
	Value *JSONValue
}

var _ Node = &RootNode{}

func (vn *RootNode) Get() JSONValue {
	return *vn.Value
}

func (vn *RootNode) Set(val JSONValue) {
	*vn.Value = val
}

func (vn *RootNode) Remove() {
	vn.Value = nil
}

// IndexNode is a member of a JSON array
type IndexNode struct {
	Array *[]any
	Index int
}

var _ Node = IndexNode{}

func (in IndexNode) Get() JSONValue {
	return (*in.Array)[in.Index]
}

func (in IndexNode) Set(val JSONValue) {
	(*in.Array)[in.Index] = val
}

func (in IndexNode) Remove() {
	*in.Array = append((*in.Array)[:in.Index], (*in.Array)[in.Index+1:]...)
}

// FieldNode is a member of a JSON object
type FieldNode struct {
	Object map[string]any
	Key    string
}

var _ Node = FieldNode{}

func (fn FieldNode) Get() JSONValue {
	return fn.Object[fn.Key]
}
func (fn FieldNode) Set(val JSONValue) {
	fn.Object[fn.Key] = val
}

func (fn FieldNode) Remove() {
	delete(fn.Object, fn.Key)
}

// QueryValue applies `query` to `node`, invoking `yield` on each
// of the nodes that the query produces, in a context where the document
// root is `root`.
// For a JSON array (Go `[]anay`), the order is not what RFC 9535 says.
func QueryValue(query Query, root, node Node, yield func(Node)) {
	if len(query) == 0 {
		yield(node)
		return
	}
	qHead := query[0]
	qTail := query[1:]
	ApplySegment(qHead, root, node, func(output Node) {
		QueryValue(qTail, root, output, yield)
	})
}

// ApplySegment invokes `yield` on every node that is selected by
// the given Segment.
func ApplySegment(seg Segment, root, node Node, yield func(Node)) {
	val := node.Get()
	switch typed := val.(type) {
	case []any:
		for idx := len(typed) - 1; idx >= 0; idx-- {
			child := IndexNode{&typed, idx}
			if seg.SelectsIndex(idx, root, child) {
				yield(child) // might delete this child
			}
		}
		if !seg.Recurse {
			return
		}
		for idx := range typed {
			child := IndexNode{&typed, idx}
			ApplySegment(seg, root, child, yield)
		}
	case map[string]any:
		for key := range typed {
			child := FieldNode{typed, key}
			if seg.SelectsField(key, root, child) {
				yield(child)
			}
		}
		if !seg.Recurse {
			return
		}
		for key := range typed {
			child := FieldNode{typed, key}
			ApplySegment(seg, root, child, yield)
		}
	default:
	}
}

func (seg Segment) SelectsIndex(idx int, root, child Node) bool {
	for _, sel := range seg.Selectors {
		if sel.SelectsIndex(idx, root, child) {
			return true
		}
	}
	return false
}

func (sel Selector) SelectsIndex(idx int, root, child Node) bool {
	if sel.IsWildcard {
		return true
	}
	if sel.Name != nil {
		return false
	}
	if sel.Slice != nil {
		var start, step int = 0, 1
		if sel.Slice.Start != nil {
			start = *sel.Slice.Start
		}
		if idx < start {
			return false
		}
		if sel.Slice.End != nil {
			if idx >= *sel.Slice.End {
				return false
			}
		}
		if sel.Slice.Step != nil {
			step = *sel.Slice.Step
			if step < 1 {
				return false
			}
		}
		return idx == step*((idx-start)/step)
	}
	if sel.Index != nil {
		return idx == *sel.Index
	}
	return evalOr(root, child, *sel.Filter)
}

func (seg Segment) SelectsField(name string, root, child Node) bool {
	for _, sel := range seg.Selectors {
		if sel.SelectsField(name, root, child) {
			return true
		}
	}
	return false
}

func (sel Selector) SelectsField(name string, root, child Node) bool {
	if sel.IsWildcard {
		return true
	}
	if sel.Name != nil {
		return name == *sel.Name
	}
	if sel.Slice != nil {
		return false
	}
	if sel.Index != nil {
		return false
	}
	return evalOr(root, child, *sel.Filter)
}

func evalOr(root, node Node, expr LogicalOrExpr) bool {
	for _, and := range expr.Terms {
		if evalAnd(root, node, and) {
			return true
		}
	}
	return false
}

func evalAnd(root, node Node, expr LogicalAndExpr) bool {
	for _, term := range expr.Factors {
		if !evalBasic(root, node, term) {
			return false
		}
	}
	return true
}

func evalBasic(root, node Node, expr BasicExpr) bool {
	if expr.Parenthesized != nil {
		return evalOr(root, node, *expr.Parenthesized) != expr.Negate
	}
	if expr.Test != nil {
		var foundSome bool
		if expr.Test.Absolute != nil {
			QueryValue(*expr.Test.Absolute, root, root, func(Node) { foundSome = true })
		} else if expr.Test.Relative != nil {
			QueryValue(*expr.Test.Absolute, root, node, func(Node) { foundSome = true })
		}
		return foundSome != expr.Negate
	}
	if expr.Compare != nil {
		left, haveLeft := evalComparable(root, node, expr.Compare.Left)
		right, haveRight := evalComparable(root, node, expr.Compare.Right)
		var isEqual bool
		if !(haveLeft && haveRight) {
			isEqual = !(haveLeft || haveRight)
		} else {
			isEqual = jsonEqualities.DeepEqual(left, right)
		}
		return isEqual == (expr.Compare.Op == CompareEQ)
	}
	return false
}

var jsonEqualities = k8sreflect.Equalities{}

func evalComparable(root, node Node, cbl Comparable) (JSONValue, bool) {
	if cbl.Literal != nil {
		return *cbl.Literal, true
	}
	var ans JSONValue
	var found bool
	if cbl.Absolute != nil {
		QueryValue(*cbl.Absolute, root, root, func(node Node) { ans = node.Get(); found = true })
	}
	if cbl.Relative != nil {
		QueryValue(*cbl.Relative, root, node, func(node Node) { ans = node.Get(); found = true })
	}
	return ans, found
}
