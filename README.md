# sam-aws-auditor
This project demonstrates how to use AWS Lambda to audit accounts in your organization using cross-account roles. The roles in this case are created automatically in each account (except for the Master account) by AWS Control Tower.

Our Lambda function is written in Go, and packaged and deployed using AWS Serverless Application Model (SAM).