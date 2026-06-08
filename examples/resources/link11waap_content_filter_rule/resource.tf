# Example: Content Filter Rule Resource

data "link11waap_config" "main" {}

resource "link11waap_content_filter_rule" "example" {
  config_id   = data.link11waap_config.main.id
  name        = "Block Malicious Domains"
  msg         = "Blocked by content filter rule"
  operand     = ".*\\.malicious\\.example\\.com$"
  category    = "malware"
  subcategory = "phishing"
  risk        = 4
  description = "Blocks requests targeting known malicious domains"

  # Optional: tags for categorization
  tags = ["security", "malware"]
}
