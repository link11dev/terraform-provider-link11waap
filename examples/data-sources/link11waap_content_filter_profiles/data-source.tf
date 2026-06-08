# Example: Content Filter Profiles Data Source

data "link11waap_config" "main" {}

# List all content filter profiles
data "link11waap_content_filter_profiles" "all" {
  config_id = data.link11waap_config.main.id
}

output "content_filter_profiles" {
  value = data.link11waap_content_filter_profiles.all.content_filter_profiles
}

output "content_filter_profile_names" {
  value = [for p in data.link11waap_content_filter_profiles.all.content_filter_profiles : p.name]
}

# Get a specific content filter profile by name
data "link11waap_content_filter_profiles" "example_by_name" {
  config_id = data.link11waap_config.main.id
  name      = "Default Content Filter Profile"
}

output "example_by_name" {
  value = data.link11waap_content_filter_profiles.example_by_name
}

# Get a specific content filter profile by ID
data "link11waap_content_filter_profiles" "example_by_id" {
  config_id = data.link11waap_config.main.id
  id        = "your-profile-id"
}

output "example_by_id" {
  value = data.link11waap_content_filter_profiles.example_by_id
}
