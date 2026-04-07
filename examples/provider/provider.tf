terraform {
  required_providers {
    link11waap = {
      source  = "link11/link11waap"
      version = "~> 0.1"
    }
  }
}

provider "link11waap" {
  domain  = var.link11_domain
  api_key = var.link11_api_key
}

variable "link11_domain" {
  description = "Link11 WAAP domain (e.g., customer.app.reblaze.io)"
  type        = string
}

variable "link11_api_key" {
  description = "Link11 WAAP API key"
  type        = string
  sensitive   = true
}
