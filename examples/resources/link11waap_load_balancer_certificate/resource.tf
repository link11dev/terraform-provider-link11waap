# Example: Load Balancer Certificate Resource

data "link11waap_config" "main" {}

# Look up available load balancers
data "link11waap_load_balancers" "all" {
  config_id = data.link11waap_config.main.id
}

# Upload a certificate
resource "link11waap_certificate" "web" {
  config_id   = data.link11waap_config.main.id
  cert_body   = file("${path.module}/certs/cert.pem")
  private_key = file("${path.module}/certs/key.pem")
  side        = "server"
}

# Attach the certificate to the first available load balancer
resource "link11waap_load_balancer_certificate" "web" {
  config_id          = data.link11waap_config.main.id
  load_balancer_name = data.link11waap_load_balancers.all.load_balancers[0].name
  certificate_id     = link11waap_certificate.web.id

  # Cloud provider. Valid values: aws, gcp, link11
  provider_type = data.link11waap_load_balancers.all.load_balancers[0].provider

  # Cloud region of the load balancer
  region = data.link11waap_load_balancers.all.load_balancers[0].region

  # Listener identifier (ARN for AWS, name for others)
  listener = data.link11waap_load_balancers.all.load_balancers[0].listener_name

  # Listener port number
  listener_port = data.link11waap_load_balancers.all.load_balancers[0].listener_port

  # Whether this is the default certificate for the load balancer (default: false)
  is_default = false

  # Use ELB v2 (Application Load Balancer) for AWS (default: true)
  elbv2 = false
}
