# Example: Global Filters Data Source

data "link11waap_config" "main" {}

data "link11waap_global_filters" "all" {
  config_id = data.link11waap_config.main.id
}

output "global_filters" {
  value = data.link11waap_global_filters.all.global_filters
}

output "global_filter_names" {
  value = [for gf in data.link11waap_global_filters.all.global_filters : gf.name]
}
