# Example: Load Balancer Regions Data Source

data "link11waap_config" "main" {}

data "link11waap_load_balancer_regions" "all" {
  config_id = data.link11waap_config.main.id
}

output "city_codes" {
  value = data.link11waap_load_balancer_regions.all.city_codes
}

output "load_balancer_regions" {
  value = [for lb in data.link11waap_load_balancer_regions.all.lbs : {
    id      = lb.id
    name    = lb.name
    regions = lb.regions
  }]
}
