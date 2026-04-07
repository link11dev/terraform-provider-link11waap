# Example: Security Policies Data Source

data "link11waap_config" "main" {}

data "link11waap_security_policies" "all" {
  config_id = data.link11waap_config.main.id
}

output "security_policies" {
  value = data.link11waap_security_policies.all.security_policies
}

output "security_policy_names" {
  value = [for sp in data.link11waap_security_policies.all.security_policies : sp.name]
}
