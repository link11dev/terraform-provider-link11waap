# Example: Planet Trusted Nets Data Source
#
# Reads the singleton trusted_nets object for the given configuration
# (id is always "__default__"). The returned trusted_nets list contains
# every entry currently configured, regardless of source type.

data "link11waap_config" "main" {}

data "link11waap_planet_trusted_nets" "current" {
  config_id = data.link11waap_config.main.id
}

# Full object, handy for debugging or feeding back into another module
output "planet_trusted_nets" {
  value = data.link11waap_planet_trusted_nets.current
}

# Raw list of entries
output "planet_trusted_nets_entries" {
  value = data.link11waap_planet_trusted_nets.current.trusted_nets
}

# Only the IP/CIDR-based trusted networks
output "planet_trusted_nets_ip_addresses" {
  value = [
    for n in data.link11waap_planet_trusted_nets.current.trusted_nets :
    n.address if n.source == "ip"
  ]
}

# Only the global filter references
output "planet_trusted_nets_global_filter_ids" {
  value = [
    for n in data.link11waap_planet_trusted_nets.current.trusted_nets :
    n.gf_id if n.source == "global_filter"
  ]
}

# Count of configured entries
output "planet_trusted_nets_entry_count" {
  value = length(data.link11waap_planet_trusted_nets.current.trusted_nets)
}
