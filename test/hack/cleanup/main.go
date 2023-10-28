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

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"

	"github.com/aws/karpenter/test/hack/cleanup/metrics"
	"github.com/aws/karpenter/test/hack/cleanup/resourcetypes"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

var expirationTTL = time.Hour * 12

const soakExpirationTTL = time.Hour * 168 // 7 Days

func main() {
	var clusterName string
	if len(os.Args) == 2 {
		clusterName = os.Args[1]
	}
	ctx := context.Background()
	cfg := lo.Must(config.LoadDefaultConfig(ctx))

	log := lo.Must(zap.NewProduction()).Sugar()

	if cfg.Region == "us-east-1" {
		expirationTTL = soakExpirationTTL + (24 * time.Hour)
	}
	if clusterName != "" {
		sweepedResources := cleanUpCluster(ctx, clusterName, cfg, log)
		if sweepedResources {
			os.Exit(0)
		}
	}
	sweepResources(ctx, clusterName, cfg, log)
}

func sweepResources(ctx context.Context, name string, cfg aws.Config, logger *zap.SugaredLogger) {
	expirationTime := time.Now().Add(-expirationTTL)

	ec2Client := ec2.NewFromConfig(cfg)
	cloudFormationClient := cloudformation.NewFromConfig(cfg)
	iamClient := iam.NewFromConfig(cfg)
	metricsClient := metrics.Client(metrics.NewTimeStream(cfg))

	logger.With("expiration-time", expirationTime.String()).Infof("resolved expiration time for all resourceTypes")
	// These resources are intentionally ordered so that instances that are using ENIs
	// will be cleaned before ENIs are attempted to be cleaned up. Likewise, instances and ENIs
	// are cleaned up before security groups are cleaned up to ensure that everything is detached and doesn't
	// prevent deletion
	resourceTypes := []resourcetypes.Type{
		resourcetypes.NewInstance(ec2Client),
		resourcetypes.NewVPCEndpoint(ec2Client),
		resourcetypes.NewENI(ec2Client),
		resourcetypes.NewSecurityGroup(ec2Client),
		resourcetypes.NewLaunchTemplate(ec2Client),
		resourcetypes.NewOIDC(iamClient),
		resourcetypes.NewInstanceProfile(iamClient),
		resourcetypes.NewStack(cloudFormationClient),
	}

	for i := range resourceTypes {
		resourceLogger := logger.With("type", resourceTypes[i].String())
		var ids []string
		var err error
		if name == "" {
			ids, err = resourceTypes[i].GetExpired(ctx, expirationTime)
		} else {
			ids, err = resourceTypes[i].Get(ctx, name)
		}
		if err != nil {
			resourceLogger.Errorf("%v", err)
		}
		resourceLogger.With("ids", ids, "count", len(ids)).Infof("discovered resourceTypes")
		if len(ids) > 0 {
			cleaned, err := resourceTypes[i].Cleanup(ctx, ids)
			if err != nil {
				resourceLogger.Errorf("%v", err)
			}
			if err = metricsClient.FireMetric(ctx, fmt.Sprintf("%sDeleted", resourceTypes[i].String()), float64(len(cleaned)), cfg.Region); err != nil {
				resourceLogger.Errorf("%v", err)
			}
			resourceLogger.With("ids", cleaned, "count", len(cleaned)).Infof("deleted resourceTypes")
		}
	}
}

func cleanUpCluster(ctx context.Context, name string, cfg aws.Config, logger *zap.SugaredLogger) bool {
	soakExpirationTime := time.Now().Add(-soakExpirationTTL)
	eksClient := eks.NewFromConfig(cfg)

	if strings.HasPrefix(name, "soak-periodic-") {
		cluster, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: aws.String(name)})
		if err != nil {
			logger.Errorf("%v", err)
			return false
		}

		if !lo.FromPtr(cluster.Cluster.CreatedAt).Before(soakExpirationTime) {
			logger.Infof("soak testing cluster (%s) does not need to be cleaned up until %s", name, soakExpirationTime)
			return true
		}
	}

	deleteCluster := exec.Command("eksctl", "delete", "cluster", "--name", name, "--timeout", "60m", "--wait")
	if out, err := deleteCluster.Output(); err != nil {
		logger.Errorf("%v", err)
	} else {
		fmt.Println(string(out))
	}

	return false
}
