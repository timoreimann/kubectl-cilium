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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	ciliumutils "github.com/timoreimann/kubectl-cilium/internal/utils/cilium"
	"github.com/timoreimann/kubectl-cilium/internal/version"
)

func init() {
	configFlags = genericclioptions.NewConfigFlags(true)
	configFlags.AddFlags(rootCmd.PersistentFlags())
	rootCmd.SetVersionTemplate("kubectl-cilium " + version.Version)
}

var (
	configFlags        *genericclioptions.ConfigFlags
	specifiedNamespace string
	ciliumNamespace    string
	kubeClient         kubernetes.Interface
	kubeConfig         *rest.Config
	streams            genericclioptions.IOStreams
)

var rootCmd = &cobra.Command{
	Version:      version.Version,
	Args:         cobra.NoArgs,
	Use:          "kubectl-cilium",
	SilenceUsage: true,
	Short:        "A kubectl plugin for interacting with Cilium",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		clientConfig := configFlags.ToRawKubeConfigLoader()

		var err error
		specifiedNamespace, _, err = clientConfig.Namespace()
		if err != nil {
			return fmt.Errorf("failed to get namespace from clientconfig: %s", err)
		}

		kubeConfig, err = clientConfig.ClientConfig()
		if err != nil {
			return fmt.Errorf("failed to get client config: %s", err)
		}
		kubeClient, err = kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			return fmt.Errorf("failed to create kube config: %s", err)
		}

		ciliumNamespace, err = ciliumutils.DiscoverCiliumNamespace(context.Background(), kubeClient)
		if err != nil {
			return fmt.Errorf("failed to discover Cilium namespace: %s", err)
		}

		streams = genericclioptions.IOStreams{
			In:     os.Stdin,
			ErrOut: os.Stderr,
			Out:    os.Stdout,
		}

		return nil
	},
}

func Execute() {
	pflag.CommandLine = pflag.NewFlagSet("kubectl-cilium", pflag.ExitOnError)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
