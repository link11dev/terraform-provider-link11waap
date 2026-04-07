# Example: ACL Profile Resource

data "link11waap_config" "main" {}

resource "link11waap_acl_profile" "web" {
  config_id   = data.link11waap_config.main.id
  name        = "Web ACL Profile"
  description = "ACL profile for web traffic"

  # Default action when a request matches a deny rule.
  # Valid values: action-acl-block, action-waap-feed-block, action-https-redirect
  action = "action-acl-block"

  # Optional: tags for categorization
  # tags = ["web", "production"]

  # Tag identifiers to allow through
  # allow = ["trusted-networks"]

  # Tag identifiers to deny
  deny = ["acl-deny"]

  # Tag identifiers to deny (bot-specific)
  deny_bot = ["apple-crawler"]

  # Tag identifiers to allow (bot-specific)
  allow_bot = ["api"]

  # Tag identifiers to force deny (overrides allow)
  force_deny = ["enforce-acl-deny"]

  # Tag identifiers to pass through without inspection
  passthrough = ["skip-waf"]
}
