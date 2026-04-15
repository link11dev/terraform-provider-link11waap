# Example: Proxy Template Resource

data "link11waap_config" "main" {}

# First, upload a certificate for the server group
resource "link11waap_certificate" "web" {
  config_id   = data.link11waap_config.main.id
  cert_body   = file("${path.module}/certs/cert.pem")
  private_key = file("${path.module}/certs/key.pem")
  side        = "server"
}

# Create simple proxy template without advanced configuration
resource "link11waap_proxy_template" "test_proxt_tpl" {
  config_id   = data.link11waap_config.main.id
  name        = "Example Proxy Template"
  description = "Proxy template for example server group"
}

# Create simple server group using the custom proxy template
resource "link11waap_server_group" "sg_with_proxy_tpl" {
  config_id               = data.link11waap_config.main.id
  name                    = "test-sg-with-proxy-tpl"
  description             = "Simple server group with custom proxy template"
  server_names            = ["www.example.com"]
  security_policy         = "__default__"
  proxy_template          = link11waap_proxy_template.test_proxt_tpl.id
  challenge_cookie_domain = "$host"
  client_certificate_mode = "off"
  ssl_certificate         = link11waap_certificate.web.id
  mobile_application_group = "__default__"
}


# Create advanced proxy template with custom configuration. Be careful with
# the advanced configuration, if you set it wrong, it can broke your server
resource "link11waap_proxy_template" "test_proxt_tpl_advanced" {
  config_id   = data.link11waap_config.main.id
  name        = "Advanced Proxy Template"
  description = "Proxy template with advanced configuration"
  ssl_protocols = ["TLSv1.2", "TLSv1.3"]
  proxy_connect_timeout = 600
  acao_header = true
  xff_header_name = "X-Forwarded-For"
  xrealip_header_name = "X-Real-IP"
  proxy_read_timeout = 600
  upstream_host = "$http_host"
  keepalive_timeout = 600
  limit_req_rate = 1000
  limit_req_burst = 200
  mask_headers = "server*|Server*|Powered-*|powered-*"
  send_timeout = 10

  advanced_configuration = [
    {
      configuration =  trimspace(<<-EOT
              -----BEGIN SERVER-----
                add_header Cache-Control "max-age=0, no-cache, no-store";
                add_header Pragma no-cache;
              }
              -----END SERVER-----
              -----BEGIN LOCATION-----
              -----END LOCATION-----
          EOT
        )
        description   = "Some example which can broke your server if you set it wrong, be careful with this"
        name          = "Just-example-advanced-config"
        protocol      = ["http"]
    }
  ]
}
