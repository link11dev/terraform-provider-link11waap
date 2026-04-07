# Example: Mobile Application Groups Data Source

data "link11waap_config" "main" {}

data "link11waap_mobile_application_groups" "all" {
  config_id = data.link11waap_config.main.id
}

output "mobile_application_groups" {
  value = data.link11waap_mobile_application_groups.all.mobile_application_groups
}

output "mobile_group_names" {
  value = [for mag in data.link11waap_mobile_application_groups.all.mobile_application_groups : mag.name]
}
