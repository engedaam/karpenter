/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package instancetype

import (
	"context"

	"github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"
	awscache "github.com/aws/karpenter-provider-aws/pkg/cache"
	"github.com/aws/karpenter-provider-aws/pkg/providers/subnet"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/karpenter/pkg/operator/controller"
)

type Controller struct {
	kubeClient     client.Client
	subnetProvider subnet.Provider
}

func NewController(kubeClient client.Client, subnetProvider subnet.Provider) *Controller {
	return &Controller{
		kubeClient:     kubeClient,
		subnetProvider: subnetProvider,
	}
}

func (c *Controller) Reconcile(ctx context.Context, _ reconcile.Request) (reconcile.Result, error) {
	nodeClassList := &v1beta1.EC2NodeClassList{}
	if err := c.kubeClient.List(ctx, nodeClassList); err != nil {
		return reconcile.Result{}, err
	}

	for i := range nodeClassList.Items {
		if err := c.subnetProvider.Update(ctx, &nodeClassList.Items[i]); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{RequeueAfter: awscache.DefaultTTL}, nil
}

func (c *Controller) Name() string {
	return "instancetype"
}

func (c *Controller) Builder(_ context.Context, m manager.Manager) controller.Builder {
	return controller.NewSingletonManagedBy(m)
}
