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
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/samber/lo"
)

func main() {
	ctx := context.Background()
	cfg := lo.Must(config.LoadDefaultConfig(ctx))
	eksClient := eks.NewFromConfig(cfg)

	var outputList []string
	createNewCluster := true

	var nextToken *string
	for {
		clusters := lo.Must(eksClient.ListClusters(ctx, &eks.ListClustersInput{NextToken: nextToken, MaxResults: aws.Int32(50)}))

		for _, cluster := range clusters.Clusters {
			clusterDetails := lo.Must(eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: aws.String(cluster)}))
			if clusterDetails.Cluster.CreatedAt.YearDay() == time.Now().YearDay() {
				createNewCluster = false
			}

			if strings.HasPrefix(cluster, "soak-periodic-") {
				outputList = append(outputList, cluster)
			}
		}

		if createNewCluster {
			outputList = append(outputList, "createNewCluster")
		}

		nextToken = clusters.NextToken
		if nextToken == nil {
			break
		}
	}

	fmt.Println(outputList)
}
