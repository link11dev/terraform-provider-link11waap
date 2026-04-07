# Example: Certificate Resource

data "link11waap_config" "main" {}

# Upload a certificate from PEM files
resource "link11waap_certificate" "uploaded" {
  config_id   = data.link11waap_config.main.id
  cert_body   = file("${path.module}/certs/cert.pem")
  private_key = file("${path.module}/certs/key.pem")

  # Certificate side. Valid values: clientCA, server, serverToBackendMTLS, backendCA
  side = "server"

  # Optional: Let's Encrypt auto-renewal settings (defaults to false)
  # le_auto_renew   = false
  # le_auto_replace = false
}

# Let's Encrypt certificate with automatic renewal
resource "link11waap_certificate" "letsencrypt" {
  config_id = data.link11waap_config.main.id
  domains   = ["www.example.com", "example.com"]

  le_auto_renew   = true
  le_auto_replace = true
}

# Client CA certificate for mTLS
resource "link11waap_certificate" "client_ca" {
  config_id = data.link11waap_config.main.id
  cert_body = file("${path.module}/certs/client-ca.pem")
  side      = "clientCA"
}

# Computed attributes available after creation:
# - name, subject, issuer, san, expires, uploaded, revoked, links

output "uploaded_cert_id" {
  value = link11waap_certificate.uploaded.id
}

output "letsencrypt_expires" {
  value = link11waap_certificate.letsencrypt.expires
}
