terraform {
  backend "s3" {
    bucket       = "bits-hw-nsp4-terraform-state"
    key          = "NSP-4-S2-S25App-HW2/terraform.tfstate"
    region                   = "ap-south-2"
    skip_region_validation   = true
    encrypt                  = true
  }
}
