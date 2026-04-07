# Example: Server Groups Data Source

data "link11waap_config" "main" {}

data "link11waap_server_groups" "all" {
  config_id = data.link11waap_config.main.id
}

output "server_group_count" {
  value = length(data.link11waap_server_groups.all.server_groups)
}

output "server_group_names" {
  value = [for sg in data.link11waap_server_groups.all.server_groups : sg.name]
}
