# Example: Publish Resource
#
# The publish resource triggers a configuration publish whenever
# tracked resources change. Use depends_on and triggers to control
# when publishing occurs.

data "link11waap_config" "main" {}

resource "link11waap_certificate" "web" {
  config_id   = data.link11waap_config.main.id
  cert_body   = file("${path.module}/certs/cert.pem")
  private_key = file("${path.module}/certs/key.pem")
  side        = "server"
}

resource "link11waap_server_group" "web" {
  config_id                = data.link11waap_config.main.id
  name                     = "Web Application"
  server_names             = ["www.example.com"]
  security_policy          = "__default__"
  proxy_template           = "__default__"
  challenge_cookie_domain  = "$host"
  ssl_certificate          = link11waap_certificate.web.id
  mobile_application_group = "__default__"
  client_certificate_mode  = "off"
}

# Publish changes whenever the server group changes
resource "link11waap_publish" "main" {
  config_id = data.link11waap_config.main.id

  # Triggers cause a re-publish when tracked values change
  triggers = {
    server_group = sha1(jsonencode(link11waap_server_group.web))
  }

  # Optional: target buckets for publishing
  buckets = [{
    name = "prod"
    url  = "gs://rbz-myexample-config/prod/"
  }]

  depends_on = [
    link11waap_server_group.web,
  ]
}
