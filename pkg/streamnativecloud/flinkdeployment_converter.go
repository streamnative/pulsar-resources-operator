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

package streamnativecloud

import (
	computeapi "github.com/streamnative/cloud-api-server/pkg/apis/compute/v1alpha1"
	resourcev1alpha1 "github.com/streamnative/pulsar-resources-operator/api/v1alpha1"
)

// convertVvpDeploymentTemplateSpec converts VvpDeploymentTemplateSpec
func convertVvpDeploymentTemplateSpec(spec resourcev1alpha1.VvpDeploymentTemplateSpec) computeapi.VvpDeploymentTemplateSpec {
	return computeapi.VvpDeploymentTemplateSpec{
		UserMetadata: convertUserMetadata(spec.UserMetadata),
		Spec:         convertVvpDeploymentDetails(spec.Spec),
	}
}

// convertUserMetadata converts UserMetadata
func convertUserMetadata(metadata resourcev1alpha1.UserMetadata) *computeapi.UserMetadata {
	return &computeapi.UserMetadata{
		Name:        metadata.Name,
		Namespace:   metadata.Namespace,
		Labels:      metadata.Labels,
		Annotations: metadata.Annotations,
		DisplayName: metadata.DisplayName,
	}
}

// convertVvpDeploymentDetails converts VvpDeploymentDetails
func convertVvpDeploymentDetails(details resourcev1alpha1.VvpDeploymentDetails) computeapi.VvpDeploymentDetails {
	return computeapi.VvpDeploymentDetails{
		DeploymentTargetName:         details.DeploymentTargetName,
		JobFailureExpirationTime:     details.JobFailureExpirationTime,
		MaxJobCreationAttempts:       details.MaxJobCreationAttempts,
		MaxSavepointCreationAttempts: details.MaxSavepointCreationAttempts,
		RestoreStrategy:              convertRestoreStrategy(details.RestoreStrategy),
		SessionClusterName:           details.SessionClusterName,
		State:                        details.State,
		Template:                     convertVvpDeploymentDetailsTemplate(details.Template),
	}
}

// convertRestoreStrategy converts VvpRestoreStrategy
func convertRestoreStrategy(strategy *resourcev1alpha1.VvpRestoreStrategy) *computeapi.VvpRestoreStrategy {
	if strategy == nil {
		return nil
	}
	return &computeapi.VvpRestoreStrategy{
		AllowNonRestoredState: strategy.AllowNonRestoredState,
		Kind:                  strategy.Kind,
	}
}

// convertVvpDeploymentDetailsTemplate converts VvpDeploymentDetailsTemplate
func convertVvpDeploymentDetailsTemplate(template resourcev1alpha1.VvpDeploymentDetailsTemplate) computeapi.VvpDeploymentDetailsTemplate {
	return computeapi.VvpDeploymentDetailsTemplate{
		Metadata: convertVvpDeploymentDetailsTemplateMetadata(template.Metadata),
		Spec:     convertVvpDeploymentDetailsTemplateSpec(template.Spec),
	}
}

// convertVvpDeploymentDetailsTemplateMetadata converts VvpDeploymentDetailsTemplateMetadata
func convertVvpDeploymentDetailsTemplateMetadata(metadata resourcev1alpha1.VvpDeploymentDetailsTemplateMetadata) *computeapi.VvpDeploymentDetailsTemplateMetadata {
	return &computeapi.VvpDeploymentDetailsTemplateMetadata{
		Annotations: metadata.Annotations,
	}
}

// convertVvpDeploymentDetailsTemplateSpec converts VvpDeploymentDetailsTemplateSpec
func convertVvpDeploymentDetailsTemplateSpec(spec resourcev1alpha1.VvpDeploymentDetailsTemplateSpec) computeapi.VvpDeploymentDetailsTemplateSpec {
	return computeapi.VvpDeploymentDetailsTemplateSpec{
		Artifact:                      convertArtifact(spec.Artifact),
		FlinkConfiguration:            spec.FlinkConfiguration,
		Kubernetes:                    convertKubernetes(spec.Kubernetes),
		LatestCheckpointFetchInterval: spec.LatestCheckpointFetchInterval,
		Logging:                       convertLogging(spec.Logging),
		NumberOfTaskManagers:          spec.NumberOfTaskManagers,
		Parallelism:                   spec.Parallelism,
		Resources:                     convertResources(spec.Resources),
	}
}

// convertKubernetes converts VvpDeploymentDetailsTemplateSpecKubernetesSpec
func convertKubernetes(k *resourcev1alpha1.VvpDeploymentDetailsTemplateSpecKubernetesSpec) *computeapi.VvpDeploymentDetailsTemplateSpecKubernetesSpec {
	if k == nil {
		return nil
	}
	return &computeapi.VvpDeploymentDetailsTemplateSpecKubernetesSpec{
		Labels: k.Labels,
	}
}

// convertLogging converts Logging
func convertLogging(logging *resourcev1alpha1.Logging) *computeapi.Logging {
	if logging == nil {
		return nil
	}
	return &computeapi.Logging{
		Log4j2ConfigurationTemplate: logging.Log4j2ConfigurationTemplate,
		Log4jLoggers:                logging.Log4jLoggers,
		LoggingProfile:              logging.LoggingProfile,
	}
}

// convertResources converts VvpDeploymentKubernetesResources
func convertResources(resources *resourcev1alpha1.VvpDeploymentKubernetesResources) *computeapi.VvpDeploymentKubernetesResources {
	if resources == nil {
		return nil
	}
	return &computeapi.VvpDeploymentKubernetesResources{
		Jobmanager:  convertResourceSpec(resources.Jobmanager),
		Taskmanager: convertResourceSpec(resources.Taskmanager),
	}
}

// convertResourceSpec converts ResourceSpec
func convertResourceSpec(spec *resourcev1alpha1.ResourceSpec) *computeapi.ResourceSpec {
	if spec == nil {
		return nil
	}
	return &computeapi.ResourceSpec{
		Cpu:    spec.CPU,
		Memory: spec.Memory,
	}
}

// convertArtifact converts Artifact
func convertArtifact(artifact *resourcev1alpha1.Artifact) *computeapi.Artifact {
	if artifact == nil {
		return nil
	}
	return &computeapi.Artifact{
		Kind:                   artifact.Kind,
		JarUri:                 artifact.JarURI,
		PythonArtifactUri:      artifact.PythonArtifactURI,
		SqlScript:              artifact.SQLScript,
		AdditionalDependencies: artifact.AdditionalDependencies,
		EntryClass:             artifact.EntryClass,
		MainArgs:               artifact.MainArgs,
		FlinkVersion:           artifact.FlinkVersion,
		FlinkImageRegistry:     artifact.FlinkImageRegistry,
		FlinkImageRepository:   artifact.FlinkImageRepository,
		FlinkImageTag:          artifact.FlinkImageTag,
		Uri:                    artifact.URI,
	}
}
