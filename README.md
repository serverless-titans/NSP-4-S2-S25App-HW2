# NSP-4-S2-S25App-HW2

This repository contains the Infrastructure-as-Code (IaC) and Application code for Homework 2.

## 🏗️ Architecture & Features

- **VPC & Networking**: Custom VPC, Public Subnet, Internet Gateway, and Custom Route Tables deployed in `ap-south-2` (Hyderabad).
- **Compute**: An EC2 instance running Amazon Linux 2023.
- **Application**: A Golang backend application deployed and started automatically on the EC2 instance using Terraform `user_data` scripts.
- **Monitoring & Security**: 
  - VPC Flow Logs pushing to CloudWatch Log Groups.
  - Custom IAM Roles and Instance Profiles.
  - Security Groups with explicitly opened HTTP (port 80) access.
- **State Management**: Terraform state is stored securely in a remote Amazon S3 backend with encryption and locking capabilities enabled.
- **CI/CD Pipeline**: 
  - **PR Validation (`pr-validation.yml`)**: Runs static code analysis (`super-linter`), security scanning (`checkov`), formatting/linting validation (`terraform validate`, `fmt`, `tflint`), and a speculative execution (`terraform plan`).
  - **Deploy (`deploy.yml`)**: Deploys the infrastructure automatically (`terraform apply`) when code is merged into the `main` branch.

## 📝 Prerequisites

- AWS Account credentials (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) configured as secrets in GitHub.
- An OIDC Deployment Role setup (reused from HW1 `infrastructure-bootstrap`).
- Hugging Face API Token (`HUGGINGFACE_API_TOKEN`) configured in GitHub Actions secrets (optional, enables AI capabilities in the Go application).

## 🚀 How to Test the Application

Once the GitHub Actions CI/CD pipeline merges and deploys the infrastructure to AWS, it will output the public IP address of your newly provisioned EC2 instance.

**To find your application URL:**
1. Navigate to the **Actions** tab in this GitHub repository.
2. Click the latest successful run of the **Deploy Application** workflow on the `main` branch.
3. Open the **Terraform Output** step.
4. Copy the IP address printed next to `application_url`.

### 1. Health Check Endpoint
To verify the Golang server is running and reachable, send a `GET` request to the `/health` endpoint:

```bash
curl -X GET http://<YOUR_EC2_PUBLIC_IP>/health
```
**Expected Response:**
```json
{
  "environment": "staging",
  "status": "ok"
}
```

### 2. Core Application Endpoint (Prompt Generation)
To test the core functionality of the LLM/ZenQuotes backend, send a `POST` request to the root endpoint with a JSON body containing a prompt:

```bash
curl -X POST http://<YOUR_EC2_PUBLIC_IP>/ \
     -H "Content-Type: application/json" \
     -d '{"prompt": "Tell me a joke"}'
```
**Expected Response:**
```json
{
  "application": "NSP-4-S2-S25App - v1.1 - EC2 Staging",
  "prompt": "Tell me a joke",
  "response": "NSP-4-S2-S25App processed your prompt: \"Tell me a joke\". Public API context: \"The only true wisdom is in knowing you know nothing.\" - Socrates",
  "source": "zenquotes"
}
```
*(Note: If Hugging Face API keys are properly configured, it will reply using an LLM model instead of the fallback ZenQuotes API).*

## 🧹 Cost Optimization (Cleanup)
To completely destroy all resources created by this project and stop incurring AWS charges, run:
```bash
export AWS_REGION=ap-south-2
terraform destroy -auto-approve
```
