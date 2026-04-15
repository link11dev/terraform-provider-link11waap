# Example: Proxy Templates Data Source

data "link11waap_config" "main" {}

data "link11waap_proxy_templates" "all" {
  config_id = data.link11waap_config.main.id
}

output "name_of_first_proxy_tpl" {
  value = data.link11waap_proxy_templates.all.proxy_templates[0].name
}

output "proxy_template_count" {
  value = length(data.link11waap_proxy_templates.all.proxy_templates)
}
