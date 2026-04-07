# Example: Rate Limit Rule Resource

data "link11waap_config" "main" {}

resource "link11waap_rate_limit_rule" "api_rate_limit" {
  config_id   = data.link11waap_config.main.id
  name        = "API Rate Limit"
  description = "Rate limit for API endpoints"

  # Whether this is a global rate limit rule
  global = false

  # Whether the rate limit rule is active
  active = true

  # Time window in seconds for counting requests
  timeframe = 60

  # Maximum number of requests allowed within the timeframe
  threshold = 100

  # Time-to-live in seconds for the rate limit ban
  ttl = 300

  # Action to take when the rate limit is exceeded
  action = "action-monitor"

  # Optional: whether the action is a ban action (default: false)
  # is_action_ban = false

  # Optional: tags for categorization
  # tags = ["api"]

  # Rate limit key configuration.
  # At least one key block is required.
  # Exactly one of attrs, args, plugins, cookies, or headers must be set per block.
  key {
    attrs = "session"
  }

  key {
    plugins = "jwt.somedata"
  }

  # Include filter: only requests matching these tags are counted
  include {
    relation = "OR"
    tags     = ["facebook"]
  }

  # Exclude filter: requests matching these tags are excluded from counting
  exclude {
    relation = "OR"
    tags     = ["tor"]
  }
}
