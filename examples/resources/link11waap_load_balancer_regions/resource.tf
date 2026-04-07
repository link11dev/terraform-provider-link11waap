# Example: Load Balancer Regions Resource
#
# Manages the region configuration for a load balancer.
# Missing region keys are automatically filled with "automatic".
# Known region codes: ams, ash, ffm, hkg, lax, lon, sgp, stl

data "link11waap_config" "main" {}

# Look up existing load balancer regions
data "link11waap_load_balancer_regions" "current" {
  config_id = data.link11waap_config.main.id
}

# Update regions for the first load balancer
resource "link11waap_load_balancer_regions" "example" {
  count = length(data.link11waap_load_balancer_regions.current.lbs) > 0 ? 1 : 0

  config_id = data.link11waap_config.main.id
  lb_id     = data.link11waap_load_balancer_regions.current.lbs[0].id

  # Map of city codes to region values.
  # Any region codes not listed here default to "automatic".
  regions = {
    "ash" = "automatic"
    "stl" = "automatic"
    "lon" = "automatic"
  }
}
