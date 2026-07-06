# Example: Load Balancer Regions Resource
#
# Manages the region configuration for a load balancer.
# Region codes left out of the map are set to "automatic" on the server,
# but only the keys listed below are tracked in Terraform state.
# Known region codes: ams, ash, ffm, hkg, lax, lon, nyc, sgp, stl
# (any other key is rejected during `terraform plan`/`validate`)

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
