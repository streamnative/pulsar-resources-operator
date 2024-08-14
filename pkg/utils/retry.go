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

type ReconcileRetryer struct {
	source     *EventSource
	mu         sync.Mutex
	tasks      map[string]*retryTask
	maxRetries int
	log        logr.Logger
	events     chan event.GenericEvent
	stop       chan struct{}
}

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

func (r *ReconcileRetryer) Close() {
	close(r.stop)
	close(r.events)
	r.source.Close()
}

func (r *ReconcileRetryer) Source() <-chan event.GenericEvent {
	return r.events
}

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
	uid := string(obj.GetUID())
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exist := r.tasks[uid]
	return exist
}

func (r *ReconcileRetryer) Remove(obj client.Object) {
	uid := string(obj.GetUID())
	r.mu.Lock()
	defer r.mu.Unlock()
	if task, exist := r.tasks[uid]; exist {
		delete(r.tasks, uid)
		r.source.Remove(task.key)
	}
}