package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// AuditIAM performs read-only audit actions around IAM resources in an AWS account.
func AuditIAM(config aws.Config, accountID string, accountName string) {
	iamc := iam.NewFromConfig(config)

	// Standard logging prefix
	log_prefix := accountName + " (" + accountID + ") - IAM -"

	listUsersInput := &iam.ListUsersInput{}
	listUsersResult, err := iamc.ListUsers(context.TODO(), listUsersInput)
	if err != nil {
		log.Fatalln(log_prefix, "Error retrieving users:", err)
		return
	}

	listRolesInput := &iam.ListRolesInput{}
	listRolesResult, err := iamc.ListRoles(context.TODO(), listRolesInput)
	if err != nil {
		log.Fatalln(log_prefix, "Error retrieving roles:", err)
		return
	}

	if len(listUsersResult.Users) == 0 {
		log.Println(log_prefix, "No users found")
	} else {
		for _, user := range listUsersResult.Users {
			log.Println(log_prefix, *user.UserName, "created on", *user.CreateDate)
		}
	}

	log.Println(log_prefix, len(listRolesResult.Roles), "roles")
	for _, role := range listRolesResult.Roles {
		// Exclude roles created by Control Tower, SSO, and other services
		if !(strings.Contains(*role.RoleName, "aws-controltower-") ||
			strings.Contains(*role.RoleName, "AWSControlTower") ||
			strings.Contains(*role.RoleName, "AWSReservedSSO_") ||
			strings.Contains(*role.RoleName, "AWSServiceRoleFor")) {
			log.Println(log_prefix, *role.RoleName, "created on", *role.CreateDate)
		}
	}
}
