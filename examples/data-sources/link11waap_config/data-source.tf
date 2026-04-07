# Example: Config Data Source

# Get the first available configuration
data "link11waap_config" "main" {}

output "config_id" {
  value = data.link11waap_config.main.id
}

output "config_version" {
  value = data.link11waap_config.main.version
}

# Get configuration by description
data "link11waap_config" "production" {
  description = "production"
}

output "production_config_id" {
  value = data.link11waap_config.production.id
}
