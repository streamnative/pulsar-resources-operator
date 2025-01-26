// Copyright (c) 2020 StreamNative, Inc.. All Rights Reserved.

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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

const (
	initialDelay = 5 * time.Second
	maxDelay     = 10 * time.Minute
)

type retryTask struct {
	key             string
	currentRetries  int
	currentDelay    time.Duration
	resourceVersion string
}

// ReconcileRetryer is a helper to trigger reconcile with retry
type ReconcileRetryer struct {
	source     *EventSource
	mu         sync.Mutex
	tasks      map[string]*retryTask
	maxRetries int
	log        logr.Logger
	events     chan event.GenericEvent
	stop       chan struct{}
}

// NewReconcileRetryer creates a new ReconcileRetryer
func NewReconcileRetryer(maxRetries int, source *EventSource) *ReconcileRetryer {
	r := &ReconcileRetryer{
		source:     source,
		tasks:      make(map[string]*retryTask),
		maxRetries: maxRetries,
		log:        ctrl.Log.WithName("ReconcileRetryer"),
		events:     make(chan event.GenericEvent, 20),
		stop:       make(chan struct{}),
	}

	go func() {
		for {
			select {
			case <-r.stop:
				return
			case genericEvent := <-r.source.Events:
				obj := genericEvent.Object
				key := string(obj.GetUID())
				r.mu.Lock()
				task, exist := r.tasks[key]
				r.mu.Unlock()
				if !exist {
					continue
				}

				r.events <- genericEvent
				if task.currentRetries >= r.maxRetries {
					r.log.Info("Max retries reached",
						"Name", obj.GetName(), "Namespace", obj.GetNamespace())
					continue
				}

				r.mu.Lock()
				task.currentRetries++
				task.currentDelay *= 2
				if task.currentDelay > maxDelay {
					task.currentDelay = maxDelay
				}
				r.source.CreateIfAbsent(task.currentDelay, genericEvent.Object, task.key)
				r.mu.Unlock()
			}
		}
	}()

	return r
}

// Close closes the ReconcileRetryer
func (r *ReconcileRetryer) Close() {
	close(r.stop)
	close(r.events)
	r.source.Close()
}

// Source returns the event source
func (r *ReconcileRetryer) Source() <-chan event.GenericEvent {
	return r.events
}

// CreateIfAbsent creates a new task if not exist
func (r *ReconcileRetryer) CreateIfAbsent(obj client.Object) {
	uid := string(obj.GetUID())
	initTask := &retryTask{
		key:             uid + "-" + obj.GetResourceVersion(),
		currentRetries:  1,
		currentDelay:    initialDelay,
		resourceVersion: obj.GetResourceVersion(),
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if task, exist := r.tasks[uid]; !exist {
		r.tasks[uid] = initTask
		r.source.CreateIfAbsent(initTask.currentDelay, obj, initTask.key)
	} else if task.resourceVersion != obj.GetResourceVersion() {
		r.log.Info("Resource version changed, reset the task",
			"Name", obj.GetName(), "Namespace", obj.GetNamespace())
		r.tasks[uid] = initTask
		r.source.CreateIfAbsent(initTask.currentDelay, obj, initTask.key)
	}
}

func (r *ReconcileRetryer) Contains(obj client.Object) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exist := r.tasks[string(obj.GetUID())]
	return exist
}

// Remove removes the task
func (r *ReconcileRetryer) Remove(obj client.Object) {
	uid := string(obj.GetUID())
	r.mu.Lock()
	defer r.mu.Unlock()
	if task, exist := r.tasks[uid]; exist {
		delete(r.tasks, uid)
		r.source.Remove(task.key)
	}
}
