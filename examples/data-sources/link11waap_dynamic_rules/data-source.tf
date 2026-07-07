# Example: Dynamic Rules Data Source

data "link11waap_config" "main" {}

data "link11waap_dynamic_rules" "all" {
  config_id = data.link11waap_config.main.id
}

output "dynamic_rules" {
  value = data.link11waap_dynamic_rules.all.dynamic_rules
}

output "dynamic_rule_names" {
  value = [for dr in data.link11waap_dynamic_rules.all.dynamic_rules : dr.name]
}
