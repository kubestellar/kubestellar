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

package placement

func NewHashMap[Key, Val any](domain HashDomain[Key]) func(MapChangeReceiver[Key, Val]) MutableMap[Key, Val] {
	return func(observer MapChangeReceiver[Key, Val]) MutableMap[Key, Val] {
		return &hashMap[Key, Val]{
			domain:   domain,
			observer: observer,
		}
	}
}

// hashMap is the dumbest possible implementation.
// TODO: better later, or find something good to import
type hashMap[Key, Val any] struct {
	domain   HashDomain[Key]
	observer MapChangeReceiver[Key, Val]
	buckets  [numHashBuckets][]Pair[Key, Val]
}

const numHashBuckets = 8

func (hm *hashMap[Key, Val]) IsEmpty() bool {
	for _, bucket := range hm.buckets {
		if len(bucket) > 0 {
			return false
		}
	}
	return true
}

func (hm *hashMap[Key, Val]) LenIsCheap() bool { return true }

func (hm *hashMap[Key, Val]) Len() int {
	var ans int
	for _, bucket := range hm.buckets {
		ans += len(bucket)
	}
	return ans
}

func (hm *hashMap[Key, Val]) Get(key Key) (Val, bool) {
	bucketNum, indexInBucket, have := hm.seek(key)
	if have {
		return hm.buckets[bucketNum][indexInBucket].Second, true
	}
	var zero Val
	return zero, false
}

func (hm *hashMap[Key, Val]) Put(key Key, val Val) {
	bucketNum, indexInBucket, have := hm.seek(key)
	if have {
		oldVal := hm.buckets[bucketNum][indexInBucket].Second
		hm.buckets[bucketNum][indexInBucket].Second = val
		if hm.observer != nil {
			hm.observer.Update(key, oldVal, val)
		}
	} else {
		hm.buckets[bucketNum] = append(hm.buckets[bucketNum], NewPair(key, val))
		if hm.observer != nil {
			hm.observer.Create(key, val)
		}
	}
}

func (hm *hashMap[Key, Val]) Delete(key Key) {
	bucketNum, indexInBucket, have := hm.seek(key)
	if !have {
		return
	}
	bucket := hm.buckets[bucketNum]
	oldVal := bucket[indexInBucket].Second
	hm.buckets[bucketNum] = append(bucket[:indexInBucket], bucket[indexInBucket+1:]...)
	if hm.observer != nil {
		hm.observer.DeleteWithFinal(key, oldVal)
	}
}

func (hm *hashMap[Key, Val]) seek(key Key) (int, int, bool) {
	bucketNum := int(hm.domain.Hash(key) % numHashBuckets)
	bucket := hm.buckets[bucketNum]
	for indexInBucket, pair := range bucket {
		if hm.domain.Equal(key, pair.First) {
			return bucketNum, indexInBucket, true
		}
	}
	return bucketNum, len(bucket), false
}

func (hm *hashMap[Key, Val]) Visit(visitor func(Pair[Key, Val]) error) error {
	for _, bucket := range hm.buckets {
		for _, pair := range bucket {
			if err := visitor(pair); err != nil {
				return err
			}
		}
	}
	return nil
}
