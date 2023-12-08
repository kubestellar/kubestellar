/*
Copyright 2023 The KubeStellar Authors.

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

package msc

import (
	"errors"
	"time"
)

// MapMultiSpaceInformerGen is a MultiSpaceInformerGen that is given the stubs.
// It is useful for constructing tests.
type MapMultiSpaceInformerGen[ClientInterface, FactoryOption, InformerFactory any] struct {
	stubseses  map[ /*provider NS*/ string]map[ /*space name*/ string]ClientInterface
	newFactory func(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory
}

var _ MultiSpaceInformerGen[int, string, float64] = &MapMultiSpaceInformerGen[int, string, float64]{}

// NewMapMSC constructs a new one holding no stubs.
// Follow this with SetStubs and optionally RemoveStubs.
func NewMapMSC[ClientInterface, FactoryOption, InformerFactory any](
	newFactory func(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory,
) *MapMultiSpaceInformerGen[ClientInterface, FactoryOption, InformerFactory] {
	return &MapMultiSpaceInformerGen[ClientInterface, FactoryOption, InformerFactory]{
		stubseses:  map[ /*provider NS*/ string]map[ /*space name*/ string]ClientInterface{},
		newFactory: newFactory}
}

func (mms *MapMultiSpaceInformerGen[ClientInterface, FactoryOption, InformerFactory]) SetStubs(name, providerNS string, stubs ClientInterface) {
	byName, ok := mms.stubseses[providerNS]
	if !ok {
		byName = map[ /*space name*/ string]ClientInterface{}
		mms.stubseses[providerNS] = byName
	}
	byName[name] = stubs
}

func (mms *MapMultiSpaceInformerGen[ClientInterface, FactoryOption, InformerFactory]) RemoveStubs(name, providerNS string, stubs ClientInterface) {
	byName, ok := mms.stubseses[providerNS]
	if !ok {
		return
	}
	delete(byName, name)
}

func (mms *MapMultiSpaceInformerGen[ClientInterface, FactoryOption, InformerFactory]) NewForSpace(name, providerNS string) (ClientInterface, error) {
	byName, ok := mms.stubseses[providerNS]
	var zero ClientInterface
	if !ok {
		return zero, errors.New("provider NS not known")
	}
	stubs, ok := byName[name]
	if !ok {
		return zero, errors.New("space not known")
	}
	return stubs, nil
}

func (mms *MapMultiSpaceInformerGen[ClientInterface, FactoryOption, InformerFactory]) NewInformerFactoryWithOptions(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory {
	return mms.newFactory(client, defaultResync, options...)
}
