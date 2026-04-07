# Example: Security Policy Resource

data "link11waap_config" "main" {}

# Create supporting resources first
resource "link11waap_acl_profile" "web" {
  config_id   = data.link11waap_config.main.id
  name        = "Web ACL Profile"
  description = "ACL profile for web traffic"
  action      = "action-acl-block"
  deny        = ["acl-deny"]
  deny_bot    = ["apple-crawler"]
  allow_bot   = ["api"]
  force_deny  = ["enforce-acl-deny"]
  passthrough = ["skip-waf"]
}

resource "link11waap_backend_service" "web" {
  config_id      = data.link11waap_config.main.id
  name           = "Web Backend Service"
  http11         = true
  transport_mode = "default"
  sticky         = "none"
  least_conn     = false

  back_hosts {
    host          = "origin.example.com"
    http_ports    = [80]
    https_ports   = [443]
    weight        = 1
    max_fails     = 3
    fail_timeout  = 10
    down          = false
    monitor_state = ""
    backup        = false
  }
}

resource "link11waap_rate_limit_rule" "api" {
  config_id = data.link11waap_config.main.id
  name      = "API Rate Limit"
  global    = false
  active    = true
  timeframe = 60
  threshold = 100
  ttl       = 300
  action    = "action-monitor"

  key {
    attrs = "session"
  }

  include {
    relation = "OR"
    tags     = ["all"]
  }

  exclude {
    relation = "OR"
    tags     = ["tor"]
  }
}

resource "link11waap_edge_function" "cache" {
  config_id   = data.link11waap_config.main.id
  name        = "Cache Control"
  description = "Custom caching for endpoints"
  phase       = "response_post"

  code = trimspace(<<-EOT
    ngx.header['cache-control'] = 'max-age=3600, s-maxage=3600, public'
  EOT
  )
}

# Security policy with map entries referencing other resources
resource "link11waap_security_policy" "web" {
  config_id   = data.link11waap_config.main.id
  name        = "Web Security Policy"
  description = "Security policy for web applications"

  # Exactly one session block is required.
  # Exactly one of attrs, args, plugins, cookies, or headers must be set.
  session {
    attrs = "ip"
  }

  # Optional: session IDs for additional session identification
  # session_ids {
  #   cookies = "session_cookie"
  # }
  # session_ids {
  #   headers = "X-Session-ID"
  # }

  # Optional: tags for categorization
  # tags = ["web", "production"]

  # Security profile map entries
  map {
    id                            = "default_entry1"
    name                          = "Default"
    match                         = "/"
    acl_profile                   = link11waap_acl_profile.web.id
    acl_profile_active            = true
    # In current version, content filter profile is not yet supported, but we include it here for completeness and future compatibility
    # In examples, we use the default content filter profile provided by Link11
    content_filter_profile        = "__defaultcontentfilter__"
    content_filter_profile_active = true
    backend_service               = link11waap_backend_service.web.id
    rate_limit_rules              = [link11waap_rate_limit_rule.api.id]
    edge_functions                = [link11waap_edge_function.cache.id]
  }

  map {
    id                            = "api-entry"
    name                          = "API"
    match                         = "/api/"
    acl_profile                   = link11waap_acl_profile.web.id
    acl_profile_active            = true
    # In current version, content filter profile is not yet supported, but we include it here for completeness and future compatibility
    # In examples, we use the default content filter profile provided by Link11
    content_filter_profile        = "__defaultcontentfilter__"
    content_filter_profile_active = true
    backend_service               = link11waap_backend_service.web.id
    rate_limit_rules              = [link11waap_rate_limit_rule.api.id]
  }
}
