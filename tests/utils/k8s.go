// Copyright (c) 2022 StreamNative, Inc.. All Rights Reserved.
package utils

import (
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// GetKubeConfig return the kubeconfig from kubeConfigPath
// if the path is empty, it will get config from $HOME/.kube/config as default
func GetKubeConfig(kubeConfigPath string) (*rest.Config, error) {
	if len(kubeConfigPath) != 0 {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}
	return clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
}
