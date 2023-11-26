/*
Copyright 2022 The KubeStellar Authors.

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

/*
Portions of this code are based on or inspired by the KCP Author's work
Copyright Copyright 2022 The KCP Authors
Original code: https://github.com/kcp-dev/logicalcluster/blob/v3.0.2/path.go
*/

package logicalcluster

import (
	"path"
)

var (
	// Wildcard is the path indicating a requests that spans many logical clusters.
	Wildcard = Path{value: "*"}

	// None represents an unset path.
	None = Path{}

	// TODO is a value created by automated refactoring tools that should be replaced by a real path.
	TODO = None
)

const (
	separator = ":"
)

// Path represents a colon separated list of words describing a path in a logical cluster hierarchy,
// like a file path in a file-system.
//
// For instance, in the following hierarchy:
//
// root/                    (62208dab)
// ├── accounting           (c8a942c5)
// │   └── us-west          (33bab531)
// │       └── invoices     (f5865fce)
// └── management           (e7e08986)
//
//	└── us-west-invoices (f5865fce)
//
// the following would all be valid paths:
//
//   - root:accounting:us-west:invoices
//   - 62208dab:accounting:us-west:invoices
//   - c8a942c5:us-west:invoices
//   - 33bab531:invoices
//   - f5865fce
//   - root:management:us-west-invoices
//   - 62208dab:management:us-west-invoices
//   - e7e08986:us-west-invoices
type Path struct {
	value string
}

// NewPath returns a new Path.
func NewPath(value string) Path {
	return Path{value}
}

// RequestPath returns a URL path segment used to access API for the stored path.
func (p Path) RequestPath() string {
	return path.Join("/clusters", p.value)
}

// String returns string representation of the stored value.
// Satisfies the Stringer interface.
func (p Path) String() string {
	return p.value
}

// Join returns a new path by adding the given path segment
// into already existing path and separating it with a colon.
func (p Path) Join(name string) Path {
	if p.value == "" {
		return Path{name}
	}
	return Path{p.value + separator + name}
}
