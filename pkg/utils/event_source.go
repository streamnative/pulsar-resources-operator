// Copyright 2025 StreamNative
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"sync"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// Event is a custom event that can be used to trigger reconcile
type Event struct {
	client.Object
}

// EventSource is a custom event source that can be used to trigger reconcile
type EventSource struct {
	Log      logr.Logger
	Events   chan event.GenericEvent
	eventMap map[string]*time.Timer
	mu       sync.Mutex
}

// NewEventSource creates a new EventSource
func NewEventSource(log logr.Logger) *EventSource {
	return &EventSource{
		Log:      log,
		Events:   make(chan event.GenericEvent, 20),
		eventMap: make(map[string]*time.Timer),
	}
}

// CreateIfAbsent triggers reconcile after delay, idempotent operation for the same key
func (s *EventSource) CreateIfAbsent(delay time.Duration, obj client.Object, key string) {
	if delay <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.eventMap[key]; ok {
		return
	}
	s.Log.Info("Will trigger reconcile after delay", "Key", key,
		"Delay", delay, "Name", obj.GetName(), "Namespace", obj.GetNamespace())
	// add a little jitter
	delay += time.Second * 2
	s.eventMap[key] = time.AfterFunc(delay, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.Log.Info("Trigger reconcile",
			"Key", key, "Name", obj.GetName(), "Namespace", obj.GetNamespace())
		s.Events <- event.GenericEvent{Object: obj}
		delete(s.eventMap, key)
	})
}

// Update updates the delay of the reconcile event
func (s *EventSource) Update(key string, delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Log.Info("Update reconcile event", "Key", key)
	if timer, ok := s.eventMap[key]; ok {
		timer.Reset(delay)
	} else {
		s.Log.Info("No reconcile event found", "Key", key)
	}
}

// Contains checks if the key exists in the event source
func (s *EventSource) Contains(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.eventMap[key]
	return ok
}

// Remove removes the reconcile event
func (s *EventSource) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Log.Info("Remove reconcile event", "Key", key)
	if timer, ok := s.eventMap[key]; ok {
		timer.Stop()
		delete(s.eventMap, key)
	}
}

// Close closes the EventSource
func (s *EventSource) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, timer := range s.eventMap {
		timer.Stop()
		delete(s.eventMap, key)
	}
	close(s.Events)
}
