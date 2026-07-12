# AGENTS.md — HW2: Infrastructure as Code Staging Environment

## 1. Project Context

This repository/folder is for DevOps Assignment 2, selected project number 14:

**Infrastructure as Code — M9 (Provisioning)**

Assignment requirement:

> Write a Terraform or CloudFormation script to provision a complete
> "Staging Environment" consisting of a VPC, Subnet, and EC2 instance
> automatically.

For this implementation, use **Terraform** and AWS.

The workspace also contains a completed `hw1` folder. HW1 implemented the
application using a serverless architecture based on AWS API Gateway and
AWS Lambda.

HW2 must reuse the application/backend logic from HW1 where practical, but
must deploy the backend as a long-running application on an **Amazon EC2
instance**.

The purpose is to demonstrate the architectural change:

HW1:
Client -> API Gateway -> Lambda

HW2:
Client -> EC2-hosted HTTP application

The EC2 instance must be provisioned automatically using Terraform.

Do not modify the `hw1` folder unless explicitly requested by the user.
Treat `hw1` as a reference implementation.

---

## 2. External Reference Repositories

Original application repository:

https://github.com/serverless-titans/NSP-4-S2-S25App

Prerequisite infrastructure/bootstrap repository:

https://github.com/rohitkpandeydev/infrastructure-bootstrap

The bootstrap repository provides prerequisite GitHub Actions / AWS IAM
roles or related deployment prerequisites.

Before implementing CI/CD, inspect the existing `hw1` implementation and
the bootstrap repository conventions. Reuse the existing AWS authentication
approach where possible.

Do not create long-lived AWS access keys if GitHub Actions OIDC and an
existing deployment role are available.

---

## 3. Primary Goal

Create a complete, reproducible AWS staging environment using Terraform.

At minimum, Terraform must provision:

1. A VPC.
2. A public subnet.
3. Internet connectivity required for the EC2 staging server.
4. An EC2 instance.
5. Security controls required to access the application.
6. Outputs required to locate and test the deployed application.

The application from HW1 must be adapted to run as a normal HTTP service
on EC2 rather than as an AWS Lambda handler behind API Gateway.

The final result should allow a user to:

1. Provision the staging environment using Terraform.
2. Deploy or bootstrap the application onto EC2.
3. Obtain the application URL from Terraform or deployment output.
4. Call the application over HTTP.
5. Destroy the staging environment using Terraform.

---

## 4. Required Architecture

Use the following baseline architecture:

Internet
|
Internet Gateway
|
Public Route Table
|
Public Subnet
|
Security Group
|
EC2 Instance
|
Application Process

Terraform should manage the infrastructure.

Required AWS resources should include, where appropriate:

- `aws_vpc`
- `aws_subnet`
- `aws_internet_gateway`
- `aws_route_table`
- `aws_route`
- `aws_route_table_association`
- `aws_security_group`
- `aws_instance`

Additional resources may be added only when they improve the implementation
without making the assignment unnecessarily complex.

Prefer a simple architecture that is easy to demonstrate and explain.

---

## 5. Terraform Requirements

Use Terraform rather than CloudFormation.

Create a clean Terraform layout under:

`hw2/terraform/`

Recommended files:

- `versions.tf`
- `providers.tf`
- `variables.tf`
- `main.tf`
- `network.tf`
- `security.tf`
- `ec2.tf`
- `outputs.tf`
- `terraform.tfvars.example`

A different split is acceptable if it is clearer.

Requirements:

- Pin or constrain the Terraform version.
- Constrain the AWS provider version.
- Do not hard-code AWS account IDs.
- Do not hard-code credentials.
- Make the AWS region configurable.
- Make the project/environment name configurable where useful.
- Use tags consistently.
- Use `staging` as the environment.
- Use Terraform outputs for important deployment information.
- Run `terraform fmt`.
- Run `terraform validate`.
- Add useful variable descriptions.
- Add useful output descriptions.

The implementation must be understandable by a student demonstrating
Infrastructure as Code.

Avoid unnecessary enterprise complexity.

---

## 6. Networking Requirements

Provision a dedicated VPC for HW2.

The VPC must contain at least one public subnet.

The public subnet must:

- belong to the HW2 VPC;
- have a route to an Internet Gateway;
- allow the EC2 instance to receive a public IPv4 address.

The route table must contain the required default Internet route.

Use configurable CIDR values where reasonable.

Suggested defaults:

- VPC CIDR: `10.20.0.0/16`
- Public subnet CIDR: `10.20.1.0/24`

Do not assume an existing default VPC.

---

## 7. EC2 Requirements

Provision one EC2 instance for the staging application.

Prefer:

- Amazon Linux 2023, unless the existing application has a strong reason
  to use another Linux distribution.
- A small instance type suitable for an academic staging environment.
- Architecture and AMI selection that work reliably in the configured AWS
  region.

Avoid hard-coding a region-specific AMI ID if Terraform can discover the
current suitable AMI using a data source.

The EC2 instance must be automatically configured sufficiently to run the
application.

Prefer `user_data` for initial bootstrapping when appropriate.

The instance should install the required runtime and start the application.

The application must survive normal process termination or instance reboot.
Use a suitable service manager such as `systemd` where practical.

---

## 8. Security Requirements

Follow least privilege and avoid unnecessary exposure.

The application port may be exposed to the Internet for assignment
demonstration purposes.

Prefer:

- HTTP application access on port 80, or
- a clearly documented configurable application port.

Do not expose every port.

Do not open SSH to `0.0.0.0/0`.

If SSH is not required for the assignment, do not require SSH.

Prefer automated deployment and diagnostics over manual SSH operations.

If SSH support is added, make the allowed source CIDR configurable and
secure by default.

Never commit:

- AWS access keys;
- AWS secret keys;
- GitHub tokens;
- private SSH keys;
- API secrets;
- `.tfstate` files containing infrastructure state;
- generated credentials.

---

## 9. Application Migration from HW1

Inspect `hw1` before changing or copying application code.

Determine:

1. The application language and runtime.
2. The Lambda entry point.
3. The request and response format.
4. Environment variables.
5. External APIs or services used.
6. Tests already available.
7. Existing GitHub Actions conventions.

Reuse the core business logic.

Do not duplicate business logic unnecessarily.

Separate Lambda-specific adapter code from application/domain logic where
required.

The EC2 version should expose equivalent functionality through a normal HTTP
server.

For example:

HW1:

API Gateway event
-> Lambda handler
-> application logic
-> Lambda/API Gateway response

HW2:

HTTP request
-> web framework route
-> same application logic
-> HTTP response

Preserve existing behavior as closely as practical.

Do not keep API Gateway or Lambda in the HW2 runtime architecture.

---

## 10. Application API

Preserve the important API behavior from HW1 where possible.

At minimum, add a health endpoint:

`GET /health`

Expected successful response:

```json
{
  "status": "ok",
  "environment": "staging"
}
```
