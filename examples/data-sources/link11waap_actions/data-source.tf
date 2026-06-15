# Example: Actions Data Source

data "link11waap_config" "main" {}

# List all actions in the configuration.
data "link11waap_actions" "all" {
  config_id = data.link11waap_config.main.id
}

# Look up a single action by name.
data "link11waap_actions" "by_name" {
  config_id = data.link11waap_config.main.id
  name      = "Block Forbidden"
}

output "actions" {
  value = data.link11waap_actions.all.actions
}

output "action_names" {
  value = [for a in data.link11waap_actions.all.actions : a.name]
}
