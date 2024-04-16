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

// The following types are used to represent the abstract syntax tree
// that results from parsing a JSONPath query.
// This follows the grammar in RFC 9535 with some omissions:
// the function extensions are omitted, as are comparisons other than `==` and `!=`.

// Query represents a `jsonpath-query`
type Query []Segment

type Segment struct {
	Recurse   bool // Indicates whether this is a descendant-segment
	Selectors []Selector
}

type Selector struct {
	Name       *string
	IsWildcard bool
	Slice      *SliceSelector
	Index      *int
	Filter     *LogicalOrExpr
}

type SliceSelector struct {
	Start, End, Step *int
}

type LogicalOrExpr struct {
	Terms []LogicalAndExpr
}

type LogicalAndExpr struct {
	Factors []BasicExpr
}

// BasicExpr is a union of three cases, two of which can have negation
// (the grammar does not allow negation for comparison-expr).
type BasicExpr struct {
	Negate        bool
	Parenthesized *LogicalOrExpr
	Test          *FilterQuery
	Compare       *ComparisonExpr
}

type FilterQuery struct {
	Absolute *Query
	Relative *[]Segment
}

type ComparisonExpr struct {
	Left, Right Comparable
	Op          ComparisonOp
}

type ComparisonOp string

const (
	CompareEQ ComparisonOp = "=="
	CompareNE ComparisonOp = "!="
)

// Comparable is a union of three cases.
// This re-uses the datatypes from Query but the values should be singular.
type Comparable struct {
	Literal  *any       // *Literal may be a string, float64, bool, or nil
	Relative *[]Segment // these are singular ones
	Absolute *Query     // but the Segments are singular ones
}
