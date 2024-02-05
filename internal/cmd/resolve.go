// Copyright 2024 timoreimann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ciliumutils "github.com/timoreimann/kubectl-cilium/internal/utils/cilium"
	nodeutils "github.com/timoreimann/kubectl-cilium/internal/utils/kubernetes"
)

func resolveCiliumPodName(ctx context.Context, target string) (string, error) {
	var resource, objName string
	targetParts := strings.SplitN(target, "/", 2)
	switch len(targetParts) {
	case 1:
		objName = targetParts[0]
	case 2:
		resource = targetParts[0]
		objName = targetParts[1]
	}

	checkNode := true
	checkPod := true
	switch resource {
	case "node", "nodes":
		checkPod = false
	case "pod", "pods":
		checkNode = false
	default:
		if resource != "" {
			return "", fmt.Errorf("unsupported resource %q", resource)
		}
	}

	var nodeName string
	namespace := specifiedNamespace

	if checkPod {
		var err error
		nodeName, err = nodeutils.GetNodeNameForPod(ctx, kubeClient, namespace, objName)
		if err != nil {
			if !errors.IsNotFound(err) {
				return "", fmt.Errorf("failed to get node name for pod %s/%s: %s", namespace, objName, err)
			}
		}
	}

	if nodeName == "" && checkNode {
		node, err := kubeClient.CoreV1().Nodes().Get(ctx, objName, metav1.GetOptions{})
		switch {
		case err == nil:
			nodeName = node.Name
		case !errors.IsNotFound(err):
			return "", fmt.Errorf("failed to get node %s: %s", objName, err)
		}
	}

	if nodeName == "" {
		return "", fmt.Errorf("failed to find node name for target %s (namespace %s)", target, namespace)
	}

	// Try to get the name of the Cilium pod running in the targeted node.
	pn, err := ciliumutils.DiscoverCiliumPodInNode(ctx, kubeClient, ciliumNamespace, nodeName)
	if err != nil {
		return "", fmt.Errorf("failed to discover Cilium pod in namespace %s for node %s: %s", ciliumNamespace, nodeName, err)
	}

	return pn, nil
}
