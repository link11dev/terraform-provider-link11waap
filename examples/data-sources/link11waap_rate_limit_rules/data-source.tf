# Example: Rate Limit Rules Data Source

data "link11waap_config" "main" {}

data "link11waap_rate_limit_rules" "all" {
  config_id = data.link11waap_config.main.id
}

output "rate_limit_rules" {
  value = data.link11waap_rate_limit_rules.all.rate_limit_rules
}

output "rate_limit_rule_names" {
  value = [for rl in data.link11waap_rate_limit_rules.all.rate_limit_rules : rl.name]
}
