# Example: ACL Profile Resource

data "link11waap_config" "main" {}

resource "link11waap_acl_profile" "web" {
  config_id   = data.link11waap_config.main.id
  name        = "Web ACL Profile"
  description = "ACL profile for web traffic"

  # The ID of the action that is performed when a request matches a tag
  # in Enforce Deny or Block / Apply, or the requestor fails a bot challenge.
  action = "123a1bc01c3a"

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
