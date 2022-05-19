// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID       string
	location             = "westus"
	resourceGroupName    = "sample-resource-group"
	activityLogAlertName = "sample-activity-log-alert"
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

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	activityLogAlert, err := createActivityLogAlert(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("activity log alert:", *activityLogAlert.ID)

	activityLogAlert, err = getActivityLogAlert(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get activity log alert:", *activityLogAlert.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createActivityLogAlert(ctx context.Context, cred azcore.TokenCredential) (*armmonitor.ActivityLogAlertResource, error) {
	activityLogAlert, err := armmonitor.NewActivityLogAlertsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := activityLogAlert.CreateOrUpdate(
		ctx,
		resourceGroupName,
		activityLogAlertName,
		armmonitor.ActivityLogAlertResource{
			Location: to.Ptr("global"),
			Properties: &armmonitor.AlertRuleProperties{
				Scopes: []*string{
					to.Ptr("subscriptions/" + subscriptionID),
				},
				Enabled: to.Ptr(true),
				Condition: &armmonitor.AlertRuleAllOfCondition{
					AllOf: []*armmonitor.AlertRuleAnyOfOrLeafCondition{
						{
							Field:  to.Ptr("category"),
							Equals: to.Ptr("Administrative"),
						},
						{
							Field:  to.Ptr("level"),
							Equals: to.Ptr("Error"),
						},
					},
				},
				Actions: &armmonitor.ActionList{
					ActionGroups: []*armmonitor.ActionGroupAutoGenerated{},
				},
				Description: to.Ptr("Sample activity log alert description"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.ActivityLogAlertResource, nil
}

func getActivityLogAlert(ctx context.Context, cred azcore.TokenCredential) (*armmonitor.ActivityLogAlertResource, error) {
	activityLogAlert, err := armmonitor.NewActivityLogAlertsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := activityLogAlert.Get(ctx, resourceGroupName, activityLogAlertName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ActivityLogAlertResource, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

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
