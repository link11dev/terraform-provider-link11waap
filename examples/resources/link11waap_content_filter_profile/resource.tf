# Example: Content Filter Profile Resource

data "link11waap_config" "main" {}

# Example content filter rule that can be referenced in the profile
resource "link11waap_content_filter_rule" "example_cf_rule" {
    category    = "malware"
    config_id   = data.link11waap_config.main.id
    description = "Blocks requests targeting known malicious domains"
    msg         = "Blocked by content filter rule"
    name        = "My Block Malicious Domains"
    operand     = ".*\\.malicious\\.example\\.com$"
    risk        = 4
    subcategory = "phishing-test"
    tags        = [
        "security-test",
        "malware-test",
    ]
}

# Example content filter profile that references the above rule
# Here is documentation about Content Filter Profiles:
# https://waap.docs.link11.com/console-walkthrough/security/content-filter/profiles

resource "link11waap_content_filter_profile" "example_cf_profile" {
  config_id    = data.link11waap_config.main.id
  name         = "Example Content Filter Profile"
  description  = "Example profile with various settings"
  masking_seed = "change-me-to-a-random-seed"

  ignore_alphanum = true
  ignore_body     = false
  content_type    = ["application/json", "application/x-www-form-urlencoded"]
  tags            = ["security", "owasp"]
  action          = "some-action-id"

  active = ["cf-rule-name:920021"]
  ignore = ["apple-crawler"]
  report = ["cf-rule-category:sqli"]

  allsections = {
    max_count         = 512
    max_length        = 1024
    enable_max_count  = true
    enable_max_length = true

    names = [
      {
        parameter  = "my-data"
        value      = "data_value"
        mask             = true
        active           = true
        case_insensitive = true
      }
    ]

    regex = [
      {
        parameter = "secret_token"
        value     = "^secret_token_.*$"
        active    = true
      }
    ]
  }
  args = {
    max_count         = 512
    max_length        = 1024
    enable_max_count  = true
    enable_max_length = true

    names = [
      {
        parameter = "super-arg"
        value     = "abcdevf"
        mask             = true
        active           = true
        case_insensitive = true
      },
      {
        parameter = "another-arg"
        value     = "another_value"
        mask             = false
        active           = false
        case_insensitive = false
      }
    ]

    regex = [
      {
        parameter = "token"
        value     = "^token_.*$"
        active    = true
      }
    ]
  }

  headers = {
    max_count         = 64
    max_length        = 1024
    enable_max_count  = true
    enable_max_length = true
    names = [
      {
        parameter = "My-Header"
        value     = "header_value"
        mask             = true
        active           = true
        case_insensitive = true
      }
    ]
    regex = [
      {
        parameter = "X-Forwarded-For"
        value     = "^192\\.168\\..*$"
        active    = true
      }
    ]
  }

  cookies = {
    max_count         = 32
    max_length        = 512
    enable_max_count  = true
    enable_max_length = true
    regex = [
      {
        parameter = "session_id"
        value     = "^sess_.*$"
        active    = true
      }
    ]
  }

  decoding = {
    base64  = true
    dual    = false
    html    = false
    unicode = false
  }

  url = {
    max_count        = 1
    max_length       = 1024
    text = [{
      active = false
      case_insensitive = false
      ignore_cf_rule_tags = ["cf-rule-name:100001"]
      mask = false
      path = "/data"
      domain = "test.com"
      key = "url"
      reg = "^/data$"
    }]
    regex = [{
      active = true
      case_insensitive = false
      mask = false
      path = "/data.*js$"
      domain = "^test1.com$"
      ignore_cf_rule_tags = [link11waap_content_filter_rule.example_cf_rule.tags[1]]
    }]
  }
  path = {
    max_length        = 1024
    enable_max_length = true
  }
}
