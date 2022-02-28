// Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.

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
func (d *Dummy) Observe(ctx context.Context) error {
	return nil
}

// Reconcile is a fake implements of Reconcile
func (d *Dummy) Reconcile(ctx context.Context) error {
	return nil
}

var _ Interface = &Dummy{}
