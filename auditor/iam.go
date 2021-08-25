package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// AuditIAMRoles lists all interesting IAM roles in an AWS account.
func AuditIAMRoles(config aws.Config, accountID string, accountName string) {
	iamc := iam.NewFromConfig(config)

	// Standard logging prefix
	log_prefix := accountName + " (" + accountID + ") - IAM ROLE -"

	// accountRoles is a list of all roles in the account
	var accountRoles []types.Role

	listRolesInput := &iam.ListRolesInput{}
	listRolesResult, err := iamc.ListRoles(context.TODO(), listRolesInput)
	if err != nil {
		log.Fatalln(log_prefix, "Error retrieving roles:", err)
		return
	}

	for _, role := range listRolesResult.Roles {
		accountRoles = append(accountRoles, role)
	}

	// While there are more results ...
	for listRolesResult.IsTruncated {
		listRolesInput = &iam.ListRolesInput{
			// Setup the next ListRoles call with the current value of Marker
			Marker: listRolesResult.Marker,
		}
		nextListRolesResult, err := iamc.ListRoles(context.TODO(), listRolesInput)
		if err != nil {
			log.Fatalln(log_prefix, "Error retrieving more roles:", err)
			return
		}

		for _, role := range nextListRolesResult.Roles {
			accountRoles = append(accountRoles, role)
		}

		// Update the value of IsTruncated
		listRolesResult.IsTruncated = nextListRolesResult.IsTruncated
	}

	log.Println(log_prefix, len(listRolesResult.Roles), "roles")

	var auditRole bool
	for _, role := range accountRoles {
		// Assume this is an an interesting role
		auditRole = true

		// Exclude roles created by Control Tower, SSO, and other services (non-interesting roles)
		for _, rp := range AWSIAMExcludedRolePatterns {
			if strings.Contains(*role.RoleName, rp) {
				auditRole = false
			}
		}

		if auditRole {
			log.Println(log_prefix, *role.RoleName, "created on", *role.CreateDate)
		}
	}
}

// AuditIAMUsers lists all IAM users in an AWS account.
func AuditIAMUsers(config aws.Config, accountID string, accountName string) {
	iamc := iam.NewFromConfig(config)

	// Standard logging prefix
	log_prefix := accountName + " (" + accountID + ") - IAM USER -"

	// accountUsers is a list of all users in the acccount
	var accountUsers []types.User

	listUsersInput := &iam.ListUsersInput{}
	listUsersResult, err := iamc.ListUsers(context.TODO(), listUsersInput)
	if err != nil {
		log.Fatalln(log_prefix, "Error retrieving users:", err)
		return
	}

	for _, user := range listUsersResult.Users {
		accountUsers = append(accountUsers, user)
	}

	// While there are more results ...
	for listUsersResult.IsTruncated {
		listUsersInput = &iam.ListUsersInput{
			// Setup the next ListUsers call with the current value of Marker
			Marker: listUsersResult.Marker,
		}
		nextListUsersResult, err := iamc.ListUsers(context.TODO(), listUsersInput)
		if err != nil {
			log.Fatalln(log_prefix, "Error retrieving more users:", err)
			return
		}

		for _, user := range nextListUsersResult.Users {
			accountUsers = append(accountUsers, user)
		}

		// Update the value of IsTruncated
		listUsersResult.IsTruncated = nextListUsersResult.IsTruncated
	}

	if len(accountUsers) == 0 {
		log.Println(log_prefix, "No users found")
	} else {
		for _, user := range accountUsers {
			log.Println(log_prefix, *user.UserName, "created on", *user.CreateDate)
		}
	}
}
