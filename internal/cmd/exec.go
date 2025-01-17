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

	"github.com/timoreimann/kubectl-cilium/internal/constants"
	nodeutils "github.com/timoreimann/kubectl-cilium/internal/utils/kubernetes"
)

func init() {
	rootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Args:                  cobra.MinimumNArgs(1),
	DisableFlagsInUseLine: true,
	Use:                   "exec [pod/]<pod>|[node/]<node> [<command> [args...]]",
	Short:                 "Execute a command in a Cilium agent",
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
	name, err := resolveCiliumPodName(ctx, target)
	if err != nil {
		return fmt.Errorf("failed to resolve Cilium pod name: %s", err)
	}

	err = nodeutils.Exec(ctx, kubeClient, kubeConfig, streams, ciliumNamespace, name, constants.CiliumAgentContainerName, true, true, command...)
	if err != nil {
		return fmt.Errorf("failed to exec command %q: %s", strings.Join(command, " "), err)
	}

	return nil
}
