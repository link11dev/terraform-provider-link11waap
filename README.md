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
| `link11waap_server_group` | Manages a Server Group (site/domain) in Link11 WAAP |
| `link11waap_acl_profile` | Manages an ACL Profile in Link11 WAAP |
| `link11waap_certificate` | Manages an SSL certificate in Link11 WAAP |
| `link11waap_load_balancer_certificate` | Attaches a certificate to a load balancer in Link11 WAAP |
| `link11waap_load_balancer_regions` | Manages load balancer region configuration in Link11 WAAP |
| `link11waap_security_policy` | Manages a Security Policy in Link11 WAAP |
| `link11waap_backend_service` | Manages a Backend Service in Link11 WAAP |
| `link11waap_proxy_template` | Manages a Proxy Template in Link11 WAAP |
| `link11waap_global_filter` | Manages a Global Filter in Link11 WAAP |
| `link11waap_planet_trusted_nets` | Manages the trusted networks list of the planet in Link11 WAAP |
| `link11waap_publish` | Publishes configuration changes in Link11 WAAP |
| `link11waap_user` | Manages a User account in Link11 WAAP |
| `link11waap_mobile_application_group` | Manages a Mobile Application Group in Link11 WAAP |
| `link11waap_rate_limit_rule` | Manages a Rate Limit Rule in Link11 WAAP |
| `link11waap_flow_control_policy` | Manages a Flow Control Policy in Link11 WAAP |
| `link11waap_edge_function` | Manages an Edge Function in Link11 WAAP |
| `link11waap_content_filter_rule` | Manages a Content Filter Rule in Link11 WAAP |
| `link11waap_content_filter_profile` | Manages a Content Filter Profile in Link11 WAAP |
| `link11waap_action` | Manages an Action in Link11 WAAP |
| `link11waap_dynamic_rule` | Manages a Dynamic Rule in Link11 WAAP |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `link11waap_config` | Retrieves configuration information from Link11 WAAP |
| `link11waap_acl_profiles` | Lists all ACL profiles in a configuration |
| `link11waap_server_groups` | Lists all server groups in a configuration |
| `link11waap_certificates` | Lists all certificates in a configuration |
| `link11waap_load_balancers` | Lists all load balancers in a configuration |
| `link11waap_load_balancer_regions` | Retrieves load balancer region configuration |
| `link11waap_security_policies` | Lists all security policies in a configuration |
| `link11waap_proxy_templates` | Lists all proxy templates in a configuration |
| `link11waap_global_filters` | Lists all global filters in a configuration |
| `link11waap_global_filter` | Fetches a single global filter by name from a configuration |
| `link11waap_backend_services` | Lists all backend services in a configuration |
| `link11waap_planet_trusted_nets` | Reads the trusted networks list from the planet |
| `link11waap_users` | Lists all users across all organizations |
| `link11waap_mobile_application_groups` | Lists all mobile application groups in a configuration |
| `link11waap_rate_limit_rules` | Lists all rate limit rules in a configuration |
| `link11waap_flow_control_policies` | Lists all flow control policies in a configuration |
| `link11waap_edge_functions` | Lists all edge functions in a configuration |
| `link11waap_content_filter_rules` | Lists all content filter rules in a configuration |
| `link11waap_content_filter_profiles` | Lists content filter profiles in a configuration |
| `link11waap_actions` | Lists actions in a configuration |
| `link11waap_dynamic_rules` | Lists all dynamic rules in a configuration |

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
