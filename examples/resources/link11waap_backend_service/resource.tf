# Example: Backend Service Resource

data "link11waap_config" "main" {}

# Single-host backend service
resource "link11waap_backend_service" "example" {
  config_id = data.link11waap_config.main.id
  name      = "Example Backend Service"

  # Optional: description
  # description = "Backend service for main website"

  # Whether to use HTTP/1.1 for upstream connections
  http11 = true

  # Transport protocol. Valid values: default, http, https, port_bridge
  transport_mode = "default"

  # Load balancing stickiness. Valid values: none, autocookie, customcookie, iphash, least_conn
  sticky = "none"

  # Whether to use least-connections load balancing
  least_conn = false

  # Optional: custom cookie name when sticky is "customcookie"
  # sticky_cookie_name = "my_cookie"

  # Optional: mTLS certificate for backend connections
  # mtls_certificate        = link11waap_certificate.mtls.id
  # mtls_trusted_certificate = link11waap_certificate.ca.id

  back_hosts {
    host          = "origin.example.com"
    http_ports    = [80]
    https_ports   = [443]
    weight        = 1
    max_fails     = 3
    fail_timeout  = 10
    down          = false
    monitor_state = ""
    backup        = false
  }
}

# Multi-host backend service with failover
resource "link11waap_backend_service" "multi_host" {
  config_id = data.link11waap_config.main.id
  name      = "Multi-Host Backend"

  http11         = true
  transport_mode = "default"
  sticky         = "none"
  least_conn     = false

  back_hosts {
    # host must be resolvable by Link11's infrastructure, e.g. via public DNS or private peering
    host          = "primary.example.com"
    http_ports    = [80]
    https_ports   = [443]
    weight        = 1
    max_fails     = 3
    fail_timeout  = 10
    down          = false
    monitor_state = ""
    backup        = false
  }

  back_hosts {
    # host must be resolvable by Link11's infrastructure, e.g. via public DNS or private peering
    host          = "secondary.example.com"
    http_ports    = [80]
    https_ports   = [443]
    weight        = 1
    max_fails     = 3
    fail_timeout  = 10
    down          = false
    monitor_state = ""
    backup        = true
  }
}
