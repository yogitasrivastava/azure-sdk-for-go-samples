// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicefabric/armservicefabric"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	clusterName       = "sample-servicefabric-cluster"
)

var (
	resourcesClientFactory     *armresources.ClientFactory
	servicefabricClientFactory *armservicefabric.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	clustersClient      *armservicefabric.ClustersClient
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	servicefabricClientFactory, err = armservicefabric.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	clustersClient = servicefabricClientFactory.NewClustersClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	cluster, err := createCluster(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service fabric cluster:", *cluster.ID)

	cluster, err = getCluster(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get service fabric cluster:", *cluster.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createCluster(ctx context.Context) (*armservicefabric.Cluster, error) {

	pollerResp, err := clustersClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		clusterName,
		armservicefabric.Cluster{
			Location: to.Ptr(location),
			Properties: &armservicefabric.ClusterProperties{
				ManagementEndpoint: to.Ptr("https://myCluster.eastus.cloudapp.azure.com:19080"),
				NodeTypes: []*armservicefabric.NodeTypeDescription{
					{
						Name:                         to.Ptr("nt1vm"),
						ClientConnectionEndpointPort: to.Ptr[int32](19000),
						HTTPGatewayEndpointPort:      to.Ptr[int32](19007),
						ApplicationPorts: &armservicefabric.EndpointRangeDescription{
							StartPort: to.Ptr[int32](20000),
							EndPort:   to.Ptr[int32](30000),
						},
						EphemeralPorts: &armservicefabric.EndpointRangeDescription{
							StartPort: to.Ptr[int32](49000),
							EndPort:   to.Ptr[int32](64000),
						},
						IsPrimary:       to.Ptr(true),
						VMInstanceCount: to.Ptr[int32](5),
						DurabilityLevel: to.Ptr(armservicefabric.DurabilityLevelBronze),
					},
				},
				FabricSettings: []*armservicefabric.SettingsSectionDescription{
					{
						Name: to.Ptr("UpgradeService"),
						Parameters: []*armservicefabric.SettingsParameterDescription{
							{
								Name:  to.Ptr("AppPollIntervalInSeconds"),
								Value: to.Ptr("60"),
							},
						},
					},
				},
				DiagnosticsStorageAccountConfig: &armservicefabric.DiagnosticsStorageAccountConfig{
					StorageAccountName:      to.Ptr("diag"),
					ProtectedAccountKeyName: to.Ptr("StorageAccountKey1"),
					BlobEndpoint:            to.Ptr("https://diag.blob.core.windows.net/"),
					QueueEndpoint:           to.Ptr("https://diag.queue.core.windows.net/"),
					TableEndpoint:           to.Ptr("https://diag.table.core.windows.net/"),
				},
				ReliabilityLevel: to.Ptr(armservicefabric.ReliabilityLevelSilver),
				UpgradeMode:      to.Ptr(armservicefabric.UpgradeModeAutomatic),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Cluster, nil
}

func getCluster(ctx context.Context) (*armservicefabric.Cluster, error) {

	resp, err := clustersClient.Get(ctx, resourceGroupName, clusterName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Cluster, nil
}

func createResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context) error {

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
