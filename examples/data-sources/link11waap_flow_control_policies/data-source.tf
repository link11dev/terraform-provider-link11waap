# Example: Flow Control Policies Data Source

data "link11waap_config" "main" {}

data "link11waap_flow_control_policies" "all" {
  config_id = data.link11waap_config.main.id
}

output "flow_control_policies" {
  value = data.link11waap_flow_control_policies.all.flow_control_policies
}

output "flow_control_policy_names" {
  value = [for fc in data.link11waap_flow_control_policies.all.flow_control_policies : fc.name]
}
