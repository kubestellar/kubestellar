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

package metrics

import (
	"sync"
	"time"

	k8smetrics "k8s.io/component-base/metrics"
)

// Sampler is something that can be prodded to act,
// and every prod will eventually be followed by an act.
// Quick successive prods may lead to only one act.
type Sampler interface {
	Prod()
	Register(RegisterFn) error
}

func NewSampler(read func() float64, opts *k8smetrics.KubeOpts) Sampler {
	return &sampler{
		Read:  read,
		Gauge: k8smetrics.NewGauge((*k8smetrics.GaugeOpts)(opts)),
	}
}

func NewListLenSampler[Elt any](getList func() []Elt, opts *k8smetrics.KubeOpts) Sampler {
	return NewSampler(func() float64 { return float64(len(getList())) }, opts)
}

type sampler struct {
	Read      func() float64
	Gauge     *k8smetrics.Gauge
	mutex     sync.Mutex
	EvalTimer *time.Timer
}

const samplerDelay = 5 * time.Second

func (sr *sampler) Register(reg RegisterFn) error {
	return reg(sr.Gauge)
}

func (sr *sampler) Prod() {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	if sr.EvalTimer != nil {
		sr.EvalTimer.Reset(samplerDelay)
	} else {
		sr.EvalTimer = time.AfterFunc(samplerDelay, sr.sample)
	}
}

func (sr *sampler) sample() {
	val := sr.Read()
	sr.Gauge.Set(val)
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	sr.EvalTimer = nil
}
