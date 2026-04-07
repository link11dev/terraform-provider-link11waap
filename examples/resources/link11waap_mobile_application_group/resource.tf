# Example: Mobile Application Group Resource

data "link11waap_config" "main" {}

resource "link11waap_mobile_application_group" "example" {
  config_id   = data.link11waap_config.main.id
  name        = "Example Mobile Application Group"
  description = "Mobile application group for example app"

  # Optional: UID header name for device identification
  uid_header = "X-Device-UID"

  # Optional: grace period value
  grace = "3600"

  # Active configuration entries
  active_config {
    active = true
    json   = "{\"test\": 123}"
    name   = "default-config"
  }

  # Application signatures
  signatures {
    active = true
    hash   = "596f75724d657373616765"
    name   = "example-signature"
  }
}
