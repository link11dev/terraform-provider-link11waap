# Example: Load Balancers Data Source

data "link11waap_config" "main" {}

data "link11waap_load_balancers" "all" {
  config_id = data.link11waap_config.main.id
}

output "load_balancer_count" {
  value = length(data.link11waap_load_balancers.all.load_balancers)
}

output "load_balancers" {
  value = [for lb in data.link11waap_load_balancers.all.load_balancers : {
    name                = lb.name
    provider            = lb.provider
    region              = lb.region
    dns_name            = lb.dns_name
    certificates        = lb.certificates
    default_certificate = lb.default_certificate
  }]
}
