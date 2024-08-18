// Copyright 2024 StreamNative
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

package reconciler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Object is a abstract interface
type Object interface {
	metav1.Object
	runtime.Object
}

// Interface implements the functions to observe and reconcile resource object
type Interface interface {
	Observe(ctx context.Context) error
	Reconcile(ctx context.Context) error
}

// Dummy is a dummy reconciler that does nothing.
type Dummy struct {
}

// Observe is a fake implements of Observe
func (d *Dummy) Observe(context.Context) error {
	return nil
}

// Reconcile is a fake implements of Reconcile
func (d *Dummy) Reconcile(context.Context) error {
	return nil
}

var _ Interface = &Dummy{}
