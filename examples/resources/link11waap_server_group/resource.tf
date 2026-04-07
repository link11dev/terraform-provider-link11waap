# Example: Server Group Resource

data "link11waap_config" "main" {}

# First, upload a certificate for the server group
resource "link11waap_certificate" "web" {
  config_id   = data.link11waap_config.main.id
  cert_body   = file("${path.module}/certs/cert.pem")
  private_key = file("${path.module}/certs/key.pem")
  side        = "server"
}

# Basic server group using default policies
resource "link11waap_server_group" "example" {
  config_id                = data.link11waap_config.main.id
  name                     = "Example Website"
  description              = "Main website server group"
  server_names             = ["www.example.com", "example.com"]
  security_policy          = "__default__"
  # In current version, proxy templates resource is not yet supported
  # And here we use the default proxy template provided by Link11 for completeness
  proxy_template           = "__default__"
  challenge_cookie_domain  = "$host"
  ssl_certificate          = link11waap_certificate.web.id
  mobile_application_group = "__default__"

  # Optional: client certificate mode for mTLS. Valid values: on, off, optional
  client_certificate_mode = "off"

  # Optional: client CA certificate for mTLS
  # client_certificate = link11waap_certificate.client_ca.id
}

# Server group with a custom security policy and mobile application group
resource "link11waap_server_group" "advanced" {
  config_id                = data.link11waap_config.main.id
  name                     = "Advanced Website"
  description              = "Server group with custom security policy"
  server_names             = ["secure.example.com"]
  security_policy          = link11waap_security_policy.web.id
  # In current version, proxy templates resource is not yet supported
  # And here we use the default proxy template provided by Link11 for completeness
  proxy_template           = "__default__"
  challenge_cookie_domain  = "$host"
  ssl_certificate          = link11waap_certificate.web.id
  mobile_application_group = link11waap_mobile_application_group.example.id
  client_certificate_mode  = "off"
}
