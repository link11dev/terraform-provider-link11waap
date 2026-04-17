# Example: Global Filter Data Source (fetch single filter by name)

data "link11waap_config" "main" {}

data "link11waap_global_filter" "filter_by_name" {
  config_id = data.link11waap_config.main.id
  name      = "API Discovery"
}

output "filter_id" {
  value = data.link11waap_global_filter.filter_by_name.id
}

output "filter_active" {
  value = data.link11waap_global_filter.filter_by_name.active
}

output "filter_api_discovery" {
  value = data.link11waap_global_filter.filter_by_name
}
