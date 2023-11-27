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

package providerclient

const (
	INCLUSTER = "incluster"
	EXTERNAL  = "external"
)

// SpaceInfo is a minimal space information.
type SpaceInfo struct {
	// Name is the name of the space.
	Name string

	// When accessing the space from within the cluster hosting the space
	// This relevant for space that is hosted in Pods and accessed from ther Pods in the same cluster
	Config map[string]string
}

// Options are the possible options for provider create/delete operations.
type Options struct {
	// Parent is a name of parent space in space hierarchy.
	Parent string
}

// Watcher watches for changes to spaces and provides events to a channel
// for the Manager to react to.
type Watcher interface {
	// Stop stops watching. Will close the channel returned by ResultChan(). Releases
	// any resources used by the watch.
	Stop()

	// ResultChan returns a chan which will receive all the events. If an error occurs
	// or Stop() is called, the implementation will close this channel and
	// release any resources used by the watch.
	ResultChan() <-chan WatchEvent
}

// EventType defines the possible types of events.
type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
	Bookmark EventType = "BOOKMARK"
	Error    EventType = "ERROR"
)

// WatchEvent is an event that is sent when a space is added, modified, or deleted.
type WatchEvent struct {
	// Type is the type of event that occurred.
	//
	// - ADDED or MODIFIED
	//	 	The space was added or updated: a new RESTConfig is available, or needs to be refreshed.
	// - DELETED
	// 		The space was deleted: the space is removed.
	// - ERROR
	// 		An error occurred while watching the space: the space is removed.
	// - BOOKMARK
	// 		A periodic event is sent that contains no new data: ignored.
	Type EventType

	// SpaceInfo is a minimal space information.
	SpaceInfo SpaceInfo

	// Name is the name of the space related to the event.
	Name string
}

// Provider defines methods to retrieve, list, and watch fleet of spaces.
// The provider is responsible for discovering and managing the lifecycle of each
// space.
type ProviderClient interface {
	Create(name string, opts Options) error
	Delete(name string, opts Options) error

	// List returns a list of spaces.
	// This method is used to discover the initial set of spaces
	// and to refresh the list of spaces periodically.
	ListSpaces() ([]SpaceInfo, error)

	// List returns a list of space names.
	// This method is used to discover the initial set of spaces
	// and to refresh the list of spaces periodically.
	ListSpacesNames() ([]string, error)

	// Get returns a space info.
	Get(name string) (SpaceInfo, error)

	// Watch returns a Watcher that watches for changes to a list of spaces
	// and react to potential changes.
	Watch() (Watcher, error)
}
