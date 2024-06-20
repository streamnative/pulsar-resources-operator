// Copyright 2022 StreamNative
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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
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

func ExecInPod(config *rest.Config, namespace, podName, containerName, command string) (string, string, error) {
	k8sCli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "", err
	}
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	const tty = false
	req := k8sCli.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).SubResource("exec").Param("container", containerName)
	req.VersionedParams(
		&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     tty,
		},
		scheme.ParameterCodec,
	)

	var stdout, stderr bytes.Buffer
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return "", strings.TrimSpace(stderr.String()), err
	}
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), nil
}

func TransferFileFromPodToPod(fromNamespace, fromPodName, fromContainerName, fromFilePath, toNamespace, toPodName, toContainerName, toFilePath string) error {
	// Create a temporary directory to store the file locally
	tempDir, err := ioutil.TempDir("", "transfer-file")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Temporary file path
	localFilePath := filepath.Join(tempDir, "transfer-file")

	// Step 1: Copy the file from the source pod to the local temporary file
	err = CopyFileFromPod(fromNamespace, fromPodName, fromContainerName, fromFilePath, localFilePath)
	if err != nil {
		return fmt.Errorf("failed to copy file from pod: %v", err)
	}

	// Step 2: Copy the file from the local temporary file to the destination pod
	err = CopyFileToPod(localFilePath, toNamespace, toPodName, toContainerName, toFilePath)
	if err != nil {
		return fmt.Errorf("failed to copy file to pod: %v", err)
	}

	return nil
}

// CopyFileFromPod copies a file from a pod to a local path
func CopyFileFromPod(namespace, podName, containerName, srcPath, dstPath string) error {
	cmd := []string{"kubectl", "cp", fmt.Sprintf("%s/%s:%s", namespace, podName, srcPath), dstPath, "-c", containerName}
	out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command %s: %v\nOutput: %s", cmd, err, string(out))
	}
	return nil
}

// CopyFileToPod copies a local file to a pod
func CopyFileToPod(srcPath, namespace, podName, containerName, dstPath string) error {
	cmd := []string{"kubectl", "cp", srcPath, fmt.Sprintf("%s/%s:%s", namespace, podName, dstPath), "-c", containerName}
	out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute command %s: %v\nOutput: %s", cmd, err, string(out))
	}
	return nil
}
