# Example: Action Resource

data "link11waap_config" "main" {}

# A custom block action that returns a 403 with a custom body and headers.
resource "link11waap_action" "block_forbidden" {
  config_id   = data.link11waap_config.main.id
  name        = "Block Forbidden"
  description = "Returns a custom 403 response"

  # One of: skip, block, challenge, ichallenge, monitor
  type = "block"

  tags = ["custom", "blocking"]

  params = {
    content = "Access denied."
    status  = 403
    headers = {
      "X-Blocked-By" = "link11-waap"
    }
  }
}

# A monitor action that only logs matching requests.
resource "link11waap_action" "monitor_only" {
  config_id = data.link11waap_config.main.id
  name      = "Monitor Only"
  type      = "monitor"
}
