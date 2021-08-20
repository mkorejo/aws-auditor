package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
)

func AuditConfig(config aws.Config, accountID string, accountName string) {
	configc := configservice.NewFromConfig(config)

	// Standard logging prefix
	log_prefix := accountName + " (" + accountID + ") - CONFIG - " + config.Region + " -"

	// Query for all configuration recorders in the current account and region
	describeConfigRecordersInput := &configservice.DescribeConfigurationRecordersInput{}
	describeConfigRecordersResult, err := configc.DescribeConfigurationRecorders(context.TODO(), describeConfigRecordersInput)
	if err != nil {
		log.Fatalln(log_prefix, "Error describing configuration recorders:", err)
		return
	}

	// AWS Config supports a single ConfigurationRecorder called "default"
	if len(describeConfigRecordersResult.ConfigurationRecorders) == 0 {
		log.Println(log_prefix, "No configuration recorders found")
	} else if (len(describeConfigRecordersResult.ConfigurationRecorders) == 1) &&
		(*describeConfigRecordersResult.ConfigurationRecorders[0].Name == "default") {
		// AWS Config is enabled! Let's check some other settings.

		// Confirm recording status
		describeConfigRecordersStatusInput := &configservice.DescribeConfigurationRecorderStatusInput{}
		describeConfigRecordersStatusResult, err := configc.DescribeConfigurationRecorderStatus(context.TODO(), describeConfigRecordersStatusInput)
		if err != nil {
			log.Fatalln(log_prefix, "Error describing configuration recorders status:", err)
			return
		}
		if !describeConfigRecordersStatusResult.ConfigurationRecordersStatus[0].Recording {
			log.Println(log_prefix, "Recording is off")
		}

		// Confirm recorder configuration (all supported resources including global resources)
		if !(describeConfigRecordersResult.ConfigurationRecorders[0].RecordingGroup.AllSupported &&
			describeConfigRecordersResult.ConfigurationRecorders[0].RecordingGroup.IncludeGlobalResourceTypes) {
			log.Println(log_prefix, "All supported resources is not enabled, or global resources not included")
		}

		// Confirm the IAM role being used by AWS Config
		if describeConfigRecordersResult.ConfigurationRecorders[0].RoleARN != &AWSConfigRole {
			log.Println(log_prefix, "Using an incorrect role")
		}

		// Confirm that AWS Config results are archived to the correct S3 bucket
		describeDeliveryChannelsInput := &configservice.DescribeDeliveryChannelsInput{}
		describeDeliveryChannelsResult, err := configc.DescribeDeliveryChannels(context.TODO(), describeDeliveryChannelsInput)
		if err != nil {
			log.Fatalln(log_prefix, "Error describing delivery channels:", err)
			return
		}
		if describeDeliveryChannelsResult.DeliveryChannels[0].S3BucketName != &AWSConfigS3BucketName {
			log.Println(log_prefix, "Using an incorrect delivery configuration for S3")
		}

		// Confirm Aggregator if Audit account
		if accountID == Accounts["Audit"] {
			describeConfigAggregatorsInput := &configservice.DescribeConfigurationAggregatorsInput{}
			describeConfigAggregatorsResult, err := configc.DescribeConfigurationAggregators(context.TODO(), describeConfigAggregatorsInput)
			if err != nil {
				log.Fatalln(log_prefix, "Error describing configuration aggregators:", err)
				return
			}
			if len(describeConfigAggregatorsResult.ConfigurationAggregators) > 1 {
				log.Println(log_prefix, "More than one aggregator is defined")
			} else if (len(describeConfigAggregatorsResult.ConfigurationAggregators) == 1) &&
				!(describeConfigAggregatorsResult.ConfigurationAggregators[0].ConfigurationAggregatorName == &AWSConfigAggregatorName) {
				log.Println(log_prefix, "Aggregator has an incorrect name")
			}
		}
	}
}
