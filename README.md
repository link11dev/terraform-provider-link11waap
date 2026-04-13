# Link11 WAAP Terraform Provider

Terraform provider for managing [Link11 WAAP](https://www.link11.com/) (Web Application and API Protection) resources.

![Link11 Logo](img/link11_logo.jpeg)

## Using the provider

Getting Started with Terraform at WAAP Link11: [waap.docs.link11.com](https://waap.docs.link11.com/using-the-product/how-do-i.../use-terraform-with-link11-waap).

Documentation is available at: [docs/providers/link11waap](https://registry.terraform.io/providers/link11dev/link11waap/latest/docs).

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for building from source)

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    link11waap = {
      source  = "link11dev/link11waap"
      version = "~> 0.1"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/link11dev/terraform-provider-link11waap.git
cd terraform-provider-link11waap
make install
```

## Authentication

The provider requires a domain and API key to authenticate with the Link11 WAAP API.

### Using Environment Variables (Recommended)

```bash
export TF_VAR_link11_domain="customer.app.reblaze.io"
export TF_VAR_link11_api_key="your-api-key"
```

```hcl
provider "link11waap" {}
```

### Using Provider Configuration

```hcl
provider "link11waap" {
  domain  = "customer.app.reblaze.io"
  api_key = "your-api-key"
}
```

### Using Variables

```hcl
variable "link11_domain" {
  description = "Link11 WAAP domain"
  type        = string
}

variable "link11_api_key" {
  description = "Link11 WAAP API key"
  type        = string
  sensitive   = true
}

provider "link11waap" {
  domain  = var.link11_domain
  api_key = var.link11_api_key
}
```

## Resources

| Resource | Description |
|----------|-------------|
| `link11waap_server_group` | Manages server groups (sites/applications) |
| `link11waap_acl_profile` | Manages ACL profiles for access control |
| `link11waap_certificate` | Manages SSL/TLS certificates (uploaded or Let's Encrypt) |
| `link11waap_load_balancer_certificate` | Attaches certificates to load balancers |
| `link11waap_load_balancer_regions` | Configures load balancer region settings |
| `link11waap_security_policy` | Manages security policies for server groups |
| `link11waap_backend_service` | Manages backend services for server groups |
| `link11waap_publish` | Triggers configuration publishing to edge nodes |
| `link11waap_user` | Manages user accounts |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `link11waap_config` | Reads current configuration metadata |
| `link11waap_acl_profiles` | Lists all ACL profiles |
| `link11waap_server_groups` | Lists all server groups |
| `link11waap_certificates` | Lists all certificates |
| `link11waap_load_balancers` | Lists all load balancers |
| `link11waap_load_balancer_regions` | Reads load balancer region configuration |
| `link11waap_security_policies` | Lists all security policies |
| `link11waap_backend_services` | Lists all backend services |
| `link11waap_users` | Lists all user accounts |

## Usage Examples

Example configurations can be found in the [examples](./examples) directory of
the repository, demonstrating how to use the provider to manage various Link11
WAAP resources.

Resource defenitions are avaliable in [docs](./docs/index.md) directory.

## Importing Resources

Existing resources can be imported into Terraform state:

```bash
# Import a server group
terraform import link11waap_server_group.example <config_id>/<server_group_id>

# Import a certificate
terraform import link11waap_certificate.example <config_id>/<certificate_id>
```

## Development

### Building

```bash
make build
```

### Installing Locally

```bash
make install
```

### Running Tests

```bash
# Unit tests
make test
```

### Linting

```bash
make lint
```

### Generating Documentation

```bash
make docs
```
