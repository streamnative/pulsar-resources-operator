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
