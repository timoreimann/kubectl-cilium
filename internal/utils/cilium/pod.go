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

package cilium

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/timoreimann/kubectl-cilium/internal/constants"
)

func DiscoverCiliumPodInNode(ctx context.Context, kubeClient kubernetes.Interface, ciliumNamespace, nodeName string) (string, error) {
	p, err := kubeClient.CoreV1().Pods(ciliumNamespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName==%s", nodeName),
		LabelSelector: constants.CiliumLabelSelector,
		Limit:         1,
	})
	if err != nil {
		return "", fmt.Errorf("failed to discover Cilium pod in node %q: %v", nodeName, err)
	}
	if len(p.Items) == 0 {
		return "", fmt.Errorf("no Cilium pod is running on node %q", nodeName)
	}
	return p.Items[0].Name, nil
}
