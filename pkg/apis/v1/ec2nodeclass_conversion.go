/*
Copyright The Kubernetes Authors.

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

package v1

import (
	"context"
	"fmt"

	"github.com/aws/karpenter-provider-aws/pkg/apis/v1beta1"
	"github.com/samber/lo"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	corev1beta1 "sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

func (in *EC2NodeClass) ConvertTo(ctx context.Context, to apis.Convertible) error {
	fmt.Println("ConvertToV1NodeClass")
	sink := to.(*v1beta1.EC2NodeClass)
	sink.Name = in.Name
	sink.UID = in.UID

	sink.Spec.AMIFamily = in.Spec.AMIFamily
	sink.Spec.InstanceProfile = in.Spec.InstanceProfile
	sink.Spec.SecurityGroupSelectorTerms = lo.Map(in.Spec.SecurityGroupSelectorTerms, func(item SecurityGroupSelectorTerm, _ int) v1beta1.SecurityGroupSelectorTerm {
		return v1beta1.SecurityGroupSelectorTerm{
			Tags: item.Tags,
			ID:   item.ID,
			Name: item.Name,
		}
	})
	sink.Spec.SubnetSelectorTerms = lo.Map(in.Spec.SubnetSelectorTerms, func(item SubnetSelectorTerm, _ int) v1beta1.SubnetSelectorTerm {
		return v1beta1.SubnetSelectorTerm{
			Tags: item.Tags,
			ID:   item.ID,
		}
	})
	sink.Spec.AMISelectorTerms = lo.Map(in.Spec.AMISelectorTerms, func(item AMISelectorTerm, _ int) v1beta1.AMISelectorTerm {
		return v1beta1.AMISelectorTerm{
			Tags:  item.Tags,
			ID:    item.ID,
			Name:  item.Name,
			Owner: item.Owner,
		}
	})
	sink.Spec.BlockDeviceMappings = lo.Map(in.Spec.BlockDeviceMappings, func(item *BlockDeviceMapping, _ int) *v1beta1.BlockDeviceMapping {
		return &v1beta1.BlockDeviceMapping{
			DeviceName: item.DeviceName,
			EBS:        (*v1beta1.BlockDevice)(item.EBS),
			RootVolume: item.RootVolume,
		}
	})
	sink.Spec.Role = in.Spec.Role
	sink.Spec.InstanceStorePolicy = (*v1beta1.InstanceStorePolicy)(in.Spec.InstanceStorePolicy)
	sink.Spec.UserData = in.Spec.UserData
	sink.Spec.Tags = in.Spec.Tags
	sink.Spec.MetadataOptions = (*v1beta1.MetadataOptions)(in.Spec.MetadataOptions)
	sink.Spec.DetailedMonitoring = in.Spec.DetailedMonitoring
	sink.Spec.AssociatePublicIPAddress = in.Spec.AssociatePublicIPAddress

	return nil
}

func (in *EC2NodeClass) ConvertFrom(ctx context.Context, from apis.Convertible) error {
	fmt.Println("ConvertFromV1NodeClass")
	sink := from.(*v1beta1.EC2NodeClass)
	in.Name = sink.Name
	in.UID = sink.UID

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	mgr, err := ctrl.NewManager(config, manager.Options{})
	if err != nil {
		return err
	}

	nodePoolList := &corev1beta1.NodePoolList{}
	err = mgr.GetClient().List(ctx, nodePoolList, client.MatchingFields{
		"spec.template.spec.nodeClassRef.name": sink.Name,
	})
	if err != nil || len(nodePoolList.Items) == 1 {
		return err
	}
	in.Spec.Kubelet = (*KubeletConfiguration)(nodePoolList.Items[0].Spec.Template.Spec.Kubelet)

	in.Spec.AMIFamily = sink.Spec.AMIFamily
	in.Spec.InstanceProfile = sink.Spec.InstanceProfile
	in.Spec.SecurityGroupSelectorTerms = lo.Map(sink.Spec.SecurityGroupSelectorTerms, func(item v1beta1.SecurityGroupSelectorTerm, _ int) SecurityGroupSelectorTerm {
		return SecurityGroupSelectorTerm{
			Tags: item.Tags,
			ID:   item.ID,
			Name: item.Name,
		}
	})
	in.Spec.SubnetSelectorTerms = lo.Map(sink.Spec.SubnetSelectorTerms, func(item v1beta1.SubnetSelectorTerm, _ int) SubnetSelectorTerm {
		return SubnetSelectorTerm{
			Tags: item.Tags,
			ID:   item.ID,
		}
	})
	in.Spec.AMISelectorTerms = lo.Map(sink.Spec.AMISelectorTerms, func(item v1beta1.AMISelectorTerm, _ int) AMISelectorTerm {
		return AMISelectorTerm{
			Tags:  item.Tags,
			ID:    item.ID,
			Name:  item.Name,
			Owner: item.Owner,
		}
	})
	in.Spec.BlockDeviceMappings = lo.Map(sink.Spec.BlockDeviceMappings, func(item *v1beta1.BlockDeviceMapping, _ int) *BlockDeviceMapping {
		return &BlockDeviceMapping{
			DeviceName: item.DeviceName,
			EBS:        (*BlockDevice)(item.EBS),
			RootVolume: item.RootVolume,
		}
	})
	in.Spec.Role = sink.Spec.Role
	in.Spec.InstanceStorePolicy = (*InstanceStorePolicy)(sink.Spec.InstanceStorePolicy)
	in.Spec.UserData = sink.Spec.UserData
	in.Spec.Tags = sink.Spec.Tags
	in.Spec.MetadataOptions = (*MetadataOptions)(sink.Spec.MetadataOptions)
	in.Spec.DetailedMonitoring = sink.Spec.DetailedMonitoring
	in.Spec.AssociatePublicIPAddress = sink.Spec.AssociatePublicIPAddress

	return nil
}
