# Example: Flow Control Policy Resource

data "link11waap_config" "main" {}

resource "link11waap_flow_control_policy" "checkout_flow" {
  config_id   = data.link11waap_config.main.id
  name        = "Example Flow Control Policy"
  description = "Enforce a specific request flow for checkout process"

  # Whether the flow control policy is active
  active = true

  # Time window in seconds within which the flow steps must be completed
  timeframe = 60

  # Optional: tags to apply to matching requests
  # tags = ["checkout"]

  # Optional: tags describing requests to include / exclude
  include = ["all"]
  exclude = ["internal:devops-ci-cd"]

  # Flow control key configuration.
  # At least one key block is required.
  # Exactly one of attrs, args, plugins, cookies, or headers must be set per block.
  key {
    attrs = "session"
  }
  key {
    args = "flow_id"
  }
  key {
      plugins = "abcd.efgh.ijkl"
  }
  key {
      cookies = "session_id"
  }
  key {
      headers = "X-Test-Header"
  }


  # Ordered steps describing the restricted request flow.
  # At least one step is required; steps must be completed in order within the timeframe.
  steps {
    method = "GET"
    uri    = "/cart"
    headers = {
      "Host" = "shop.example.com"
    }
  }

  steps {
    method = "POST"
    uri    = "/checkout"

    # Optional: header/cookie/arg/plugin name-value pairs required at this step
    headers = {
      "X-Requested-With" = "XMLHttpRequest"
    }
    args    = {
      "test" = "value"
    }
    cookies = {
        "my_cookie" = "123456"
    }
  }
  steps {
    method = "GET"
    uri    = "/"
  }
}
