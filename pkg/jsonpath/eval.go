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

// This file implements JSONPath querying for the small subset
// of JSONPath that this package currently supports.

// The algorithms and data structures in here are designed for serialized usage,
// not concurrent usage.

// JSONValue is something that can be produced by encoding/json.Unmarshal into a pointer
// to a nil `any`.
// That is: `bool`, `float64`, `string`, `nil`, `[]any`, or `map[string]any` --- where those
// nested `any` have the same restriction.
type JSONValue = any

// Node is a JSON document node.
type Node interface {
	// Get returns (the current contents of the node, `true`) if the node is in the document,
	// (`nil`, `false`) otherwise.
	Get() (JSONValue, bool)

	// Remove deletes the node from the JSON document.
	Remove()
}

// RootNode is the Node implementation to use for the document's root node.
// Set `*Value` to the document.
// Removing this node amounts to setting `Value` to `nil`.
type RootNode struct {
	Value *JSONValue
}

var _ Node = &RootNode{}

func (vn *RootNode) Get() (JSONValue, bool) {
	if vn.Value == nil {
		return nil, false
	}
	return *vn.Value, true
}

func (vn *RootNode) Remove() {
	vn.Value = nil
}

// FieldNode is a member of a JSON object
type FieldNode struct {
	Object map[string]any
	Key    string
}

var _ Node = FieldNode{}

func (fn FieldNode) Get() (JSONValue, bool) {
	val, have := fn.Object[fn.Key]
	return val, have
}

func (fn FieldNode) Remove() {
	delete(fn.Object, fn.Key)
}

// ArrayItemNode is an element of a JSON array
type ArrayItemNode struct {
	Slice []any
	Index int
}

var _ Node = ArrayItemNode{}

func (an ArrayItemNode) Get() (JSONValue, bool) {
	if an.Index < 0 || an.Index >= len(an.Slice) {
		return nil, false
	}
	return an.Slice[an.Index], true
}

func (an ArrayItemNode) Remove() {
	if an.Index >= 0 && an.Index < len(an.Slice) {
		an.Slice[an.Index] = nil
	}
}

// QueryValue applies `query` to `node`, invoking `yield` on each
// of the nodes that the query produces, in a context where the document
// root is `root`.
func QueryValue(query Query, node Node, yield func(Node)) {
	queryValueHelper(query, node, yield)
}

func queryValueHelper(query Query, node Node, yield func(Node)) {
	if len(query) == 0 {
		yield(node)
		return
	}

	segment := query[0]
	val, ok := node.Get()
	if !ok {
		return
	}

	if segment.Kind == SegmentKindField {
		objM, ok := val.(map[string]any)
		if !ok {
			return
		}
		nextNode := FieldNode{Object: objM, Key: segment.Name}
		queryValueHelper(query[1:], nextNode, yield)
	} else if segment.Kind == SegmentKindWildcard {
		sliceVal, ok := val.([]any)
		if !ok {
			return
		}
		for i := range sliceVal {
			nextNode := ArrayItemNode{Slice: sliceVal, Index: i}
			queryValueHelper(query[1:], nextNode, yield)
		}
	}
}
