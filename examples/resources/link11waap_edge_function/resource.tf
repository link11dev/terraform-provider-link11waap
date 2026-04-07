# Example: Edge Function Resource

data "link11waap_config" "main" {}

# Edge function that runs before the request is processed
resource "link11waap_edge_function" "rate_limit_api" {
  config_id   = data.link11waap_config.main.id
  name        = "Rate Limit API"
  description = "Custom rate limiting for API endpoints"

  # Phase at which the edge function executes.
  # Valid values: request_pre, request_post, response_pre, response_post
  phase = "request_pre"

  code = trimspace(<<-EOT
    -- Custom request-phase logic
    local uri = ngx.var.uri
    if string.match(uri, "^/api/") then
      ngx.req.set_header("X-Custom-Rate-Limit", "true")
    end
  EOT
  )
}

# Edge function that runs after the response is generated
resource "link11waap_edge_function" "add_security_headers" {
  config_id   = data.link11waap_config.main.id
  name        = "Add Security Headers"
  description = "Adds security headers to all responses"

  phase = "response_post"

  code = trimspace(<<-EOT
    -- Set custom cache-control headers for specific endpoints
    ngx.header['cache-control'] = 'max-age=3600, s-maxage=3600, public'
  EOT
  )
}
