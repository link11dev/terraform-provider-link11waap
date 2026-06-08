# Example: Content Filter Rules Data Source

data "link11waap_config" "main" {}

data "link11waap_content_filter_rules" "all" {
  config_id = data.link11waap_config.main.id
}

output "content_filter_rules" {
  value = data.link11waap_content_filter_rules.all.content_filter_rules
}

output "content_filter_rule_names" {
  value = [for rule in data.link11waap_content_filter_rules.all.content_filter_rules : rule.name]
}

output "high_risk_rules" {
  value = [for rule in data.link11waap_content_filter_rules.all.content_filter_rules : rule if rule.risk >= 4]
}
