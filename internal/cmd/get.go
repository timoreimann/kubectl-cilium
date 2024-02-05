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

	"github.com/spf13/cobra"
)

var includeNamespace bool

func init() {
	getCmd.Flags().BoolVarP(&includeNamespace, "include-namespace", "i", false, "include Cilium namespace in output")
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Use:                   "get [pod/]<pod>|[node/]<node>",
	Short:                 "Get a Cilium agent pod name",
	Long: `Get the name of the Cilium agent managing the given node or pod.

Whether a pod or node is referenced by the given name is auto-discovered. A particular type can be enforced by prefixing the resource name with a slash delimiter. Both singular and plural resource name variations are supported.

If no namespace is specified or defined in the kube context, "default" is used.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.TODO()
		return get(ctx, args[0])
	},
}

func get(ctx context.Context, target string) error {
	name, err := resolveCiliumPodName(ctx, target)
	if err != nil {
		return err
	}

	output := name
	if includeNamespace {
		output = fmt.Sprintf("%s/%s", ciliumNamespace, output)
	}

	fmt.Println(output)
	return nil
}
