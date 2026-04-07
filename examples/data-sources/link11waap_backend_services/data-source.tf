# Example: Backend Services Data Source

data "link11waap_config" "main" {}

# List all backend services
data "link11waap_backend_services" "all" {
  config_id = data.link11waap_config.main.id
}

output "backend_services" {
  value = data.link11waap_backend_services.all.backend_services
}

# Get a specific backend service by ID
data "link11waap_backend_services" "by_id" {
  config_id = data.link11waap_config.main.id
  id        = "my-backend-service-id"
}

output "backend_service_name" {
  value = data.link11waap_backend_services.by_id.backend_services[0].name
}
