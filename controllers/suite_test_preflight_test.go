// Copyright 2026 StreamNative
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

package controllers

import "testing"

func TestValidateKubebuilderAssetsEnv(t *testing.T) {
	t.Setenv("KUBEBUILDER_ASSETS", "/tmp/envtest-assets")
	if err := validateKubebuilderAssetsEnv(); err != nil {
		t.Fatalf("expected non-empty KUBEBUILDER_ASSETS to pass validation, got %v", err)
	}

	t.Setenv("KUBEBUILDER_ASSETS", "")
	if err := validateKubebuilderAssetsEnv(); err == nil {
		t.Fatal("expected empty KUBEBUILDER_ASSETS to fail validation")
	}
}
