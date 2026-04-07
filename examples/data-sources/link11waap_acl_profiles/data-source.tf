# Example: ACL Profiles Data Source

data "link11waap_config" "main" {}

data "link11waap_acl_profiles" "all" {
  config_id = data.link11waap_config.main.id
}

output "acl_profiles" {
  value = data.link11waap_acl_profiles.all.acl_profiles
}

output "acl_profile_names" {
  value = [for acl in data.link11waap_acl_profiles.all.acl_profiles : acl.name]
}
