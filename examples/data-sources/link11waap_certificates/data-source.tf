# Example: Certificates Data Source

data "link11waap_config" "main" {}

# List all certificates
data "link11waap_certificates" "all" {
  config_id = data.link11waap_config.main.id
}

output "certificate_count" {
  value = length(data.link11waap_certificates.all.certificates)
}

output "certificates" {
  value = [for cert in data.link11waap_certificates.all.certificates : {
    id      = cert.id
    name    = cert.name
    subject = cert.subject
    expires = cert.expires
  }]
}

# Get a specific certificate by ID
data "link11waap_certificates" "by_id" {
  config_id = data.link11waap_config.main.id
  id        = "placeholder"
}

output "placeholder_cert" {
  value = data.link11waap_certificates.by_id.id
}

# Get a specific certificate by name
data "link11waap_certificates" "by_name" {
  config_id = data.link11waap_config.main.id
  name      = "my-cert-name"
}

output "cert_by_name" {
  value = data.link11waap_certificates.by_name.name
}
