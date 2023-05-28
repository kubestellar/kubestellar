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

// ClientTracker keeps track of a set of clients
// and reports when RemoveClient removed the last of them.
// NOT safe for concurrent use.
type ClientTracker[Provider any] struct {
	provider Provider
	clients  map[Client[Provider]]Empty
}

func NewClientTracker[Provider any]() *ClientTracker[Provider] {
	return &ClientTracker[Provider]{
		clients: map[Client[Provider]]Empty{},
	}
}

func (ct *ClientTracker[Provider]) IsEmpty() bool {
	return len(ct.clients) == 0
}

func (ct *ClientTracker[Provider]) AddClient(client Client[Provider]) {
	if _, found := ct.clients[client]; found {
		return
	}
	ct.clients[client] = Empty{}
	client.SetProvider(ct.provider)
}

func (ct *ClientTracker[Provider]) RemoveClient(client Client[Provider]) bool {
	if _, found := ct.clients[client]; !found {
		return false
	}
	delete(ct.clients, client)
	return len(ct.clients) == 0
}

func (ct *ClientTracker[Provider]) SetProvider(provider Provider) {
	ct.provider = provider
	for client := range ct.clients {
		client.SetProvider(provider)
	}
}
