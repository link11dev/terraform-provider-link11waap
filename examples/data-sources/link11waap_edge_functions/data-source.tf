# Example: Edge Functions Data Source

data "link11waap_config" "main" {}

data "link11waap_edge_functions" "all" {
  config_id = data.link11waap_config.main.id
}

output "edge_functions" {
  value = data.link11waap_edge_functions.all.edge_functions
}

output "edge_function_names" {
  value = [for ef in data.link11waap_edge_functions.all.edge_functions : ef.name]
}
