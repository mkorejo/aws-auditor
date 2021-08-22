package main

import "errors"

var (
	// Accounts contains the IDs for the management and Audit accounts
	Accounts = map[string]string{
		"Management": "665735848255",
		"Audit":      "984217156667",
	}

	// Regions to audit
	Regions = [17]string{
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
		"ca-central-1",
		"eu-central-1",
		"eu-north-1",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-northeast-3",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-south-1",
		"sa-east-1",
	}

	// Cross-account role with ReadOnly access
	AWSAuditorRole = "aws-controltower-ReadOnlyExecutionRole"

	// AWS Config
	// AWSConfigAggregatorName
	AWSConfigAggregatorName = "aws-controltower-GuardrailsComplianceAggregator"
	// AWSConfigRole must exist in every account with "ReadOnlyAccess" and "AWS_ConfigRole" policies attached.
	AWSConfigRole = "aws-controltower-ConfigRecorderRole"
	// AWSConfigSnapshotDeliveryFrequency
	// AWSConfigS3BucketName is the archival bucket in Log archive
	AWSConfigS3BucketName = "aws-controltower-logs-543705552769-us-east-1"
	// AWSConfigS3KeyPrefix
	// AWSConfigS3KMSKeyARN
	// AWSConfigSNSTopicAccount is the account which contains the SNS topic for delivery
	AWSConfigSNSTopicAccount = "984217156667"
	// AWSConfigSNSTopicName is the name of the SNS topic for delivery
	AWSConfigSNSTopicName = "aws-controltower-AllConfigNotifications"

	// AWS IAM
	// AWSIAMExcludedRolePatterns
	AWSIAMExcludedRolePatterns = []string{
		"aws-controltower-",
		"AWSControlTower",
		"AWSReservedSSO_",
		"AWSServiceRoleFor",
	}

	// Handler configuration
	// DefaultHTTPGetAddress is the site to GET
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"
	// ErrNoIP occurs when no IP is found in response
	ErrNoIP = errors.New("No IP in HTTP response")
	// ErrNon200Response occurs when non-200 status code is received
	ErrNon200Response = errors.New("Non-200 response found")
)
