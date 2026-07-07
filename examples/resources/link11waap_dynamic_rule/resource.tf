# Example: Dynamic Rule Resource

data "link11waap_config" "main" {}

resource "link11waap_dynamic_rule" "burst_protection" {
  config_id   = data.link11waap_config.main.id
  name        = "Burst Protection"
  description = "Ban clients that exceed the request threshold"

  # Maximum number of matching requests within the timeframe
  threshold = 100

  # Time window in seconds
  timeframe = 60

  # Ban duration in seconds; must be a positive multiple of 3600 (full hours)
  ttl = 3600

  # Whether the rule is active
  active = true

  # Whether IP filtering is offloaded to the edge
  offload_ip_filtering = false

  # What the rule counts on (e.g., "ip", "session", "uri")
  target = "ip"

  # Action to take when the threshold is exceeded
  action = "action-monitor"

  # Optional: tags for categorization
  # tags = ["api"]

  # Exactly one "include" block and exactly one "exclude" block are required.

  # Include filter: requests matching these tags are counted
  include {
    relation = "OR"
    tags     = ["facebook"]
  }

  # Exclude filter: requests matching these tags are excluded from counting
  exclude {
    relation = "OR"
    tags     = []
  }
}
