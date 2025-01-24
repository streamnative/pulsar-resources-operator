package controllers

import (
	"bytes"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Finalizers for resources
	FlinkDeploymentFinalizer = "resource.compute.streamnative.io/flinkdeployment-finalizer"
	WorkspaceFinalizer       = "resource.compute.streamnative.io/workspace-finalizer"
)

// StatusHasChanged compares two slices of conditions to determine if there's a meaningful change
func StatusHasChanged(oldConditions, newConditions []metav1.Condition) bool {
	if len(oldConditions) != len(newConditions) {
		return true
	}

	// Create maps for easier comparison
	oldMap := make(map[string]metav1.Condition)
	for _, condition := range oldConditions {
		oldMap[condition.Type] = condition
	}

	// Compare each new condition with the old one
	for _, newCond := range newConditions {
		oldCond, exists := oldMap[newCond.Type]
		if !exists {
			return true
		}

		// Compare relevant fields (ignoring LastTransitionTime)
		if oldCond.Status != newCond.Status ||
			oldCond.Reason != newCond.Reason ||
			oldCond.Message != newCond.Message ||
			oldCond.ObservedGeneration != newCond.ObservedGeneration {
			return true
		}
	}

	return false
}

// FlinkDeploymentStatusHasChanged compares two FlinkDeployment statuses to determine if there's a meaningful change
func FlinkDeploymentStatusHasChanged(old, new *resourcev1alpha1.ComputeFlinkDeploymentStatus) bool {
	// Compare conditions
	if StatusHasChanged(old.Conditions, new.Conditions) {
		return true
	}

	// Compare DeploymentStatus
	if old.DeploymentStatus == nil && new.DeploymentStatus == nil {
		return false
	}
	if (old.DeploymentStatus == nil && new.DeploymentStatus != nil) ||
		(old.DeploymentStatus != nil && new.DeploymentStatus == nil) {
		return true
	}

	// Compare the raw bytes of DeploymentStatus
	return !bytes.Equal(old.DeploymentStatus.Raw, new.DeploymentStatus.Raw)
}

// ContainsString checks if a string is present in a slice
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// RemoveString removes a string from a slice
func RemoveString(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
