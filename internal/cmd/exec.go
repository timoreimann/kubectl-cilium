// Copyright 2020 bmcustodio
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

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/timoreimann/kubectl-cilium/internal/constants"
	ciliumutils "github.com/timoreimann/kubectl-cilium/internal/utils/cilium"
	nodeutils "github.com/timoreimann/kubectl-cilium/internal/utils/kubernetes"
)

func init() {
	rootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Args:                  cobra.MinimumNArgs(1),
	DisableFlagsInUseLine: true,
	Use:                   "exec [pod/]<pod>|[node/]<node> [<command> [args...]]",
	Short:                 "Execute a command in a particular Cilium agent",
	Long: `Execute a command in a Cilium agent managing the given node or pod.

Whether a pod or node is referenced by the given name is auto-discovered. A particular type can be enforced by prefixing the resource name with a slash delimiter. Both singular and plural resource name variations are supported.

If no namespace is specified or defined in the kube context, "default" is used.

The default exec command is "/bin/bash".`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var command []string
		switch len(args) {
		case 1:
			command = []string{constants.DefaultCommand}
		default:
			command = args[1:]
		}

		ctx := context.TODO()
		return exec(ctx, args[0], command...)
	},
}

func exec(ctx context.Context, target string, command ...string) error {
	// Start by attempting to discover the namespace in which Cilium is installed.
	ciliumNamespace, err := ciliumutils.DiscoverCiliumNamespace(ctx, kubeClient)
	if err != nil {
		return fmt.Errorf("failed to discover Cilium namespace: %s", err)
	}

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
			return fmt.Errorf("unsupported resource %q", resource)
		}
	}

	var nodeName string
	namespace := specifiedNamespace

	if checkPod {
		nodeName, err = nodeutils.GetNodeNameForPod(ctx, kubeClient, namespace, objName)
		if err != nil {
			if !errors.IsNotFound(err) {
				return fmt.Errorf("failed to get node name for pod %s/%s: %s", namespace, objName, err)
			}
		}
	}

	if nodeName == "" && checkNode {
		node, err := kubeClient.CoreV1().Nodes().Get(ctx, objName, metav1.GetOptions{})
		switch {
		case err == nil:
			nodeName = node.Name
		case !errors.IsNotFound(err):
			return fmt.Errorf("failed to get node %s: %s", objName, err)
		}
	}

	if nodeName == "" {
		return fmt.Errorf("failed to find node name for target %s (namespace %s)", target, namespace)
	}

	// Try to get the name of the Cilium pod running in the targeted node.
	pn, err := ciliumutils.DiscoverCiliumPodInNode(ctx, kubeClient, ciliumNamespace, nodeName)
	if err != nil {
		return fmt.Errorf("failed to discover Cilium pod in namespace %s for node %s: %s", ciliumNamespace, nodeName, err)
	}
	return nodeutils.Exec(ctx, kubeClient, kubeConfig, streams, ciliumNamespace, pn, constants.CiliumAgentContainerName, true, true, command...)
}
