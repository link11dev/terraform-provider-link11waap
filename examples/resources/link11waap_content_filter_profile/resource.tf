resource "link11waap_content_filter_profile" "example" {
  config_id    = "my-config-id"
  name         = "Default Content Filter Profile"
  description  = "Inspects arguments, headers, cookies and path"
  masking_seed = "change-me-to-a-random-seed"

  ignore_alphanum = false
  ignore_body     = false
  content_type    = ["application/json", "application/x-www-form-urlencoded"]
  tags            = ["security", "owasp"]

  args = {
    max_count         = 512
    max_length        = 1024
    enable_max_count  = true
    enable_max_length = true

    names = [
      {
        key              = "password"
        mask             = true
        active           = true
        case_insensitive = true
      }
    ]

    regex = [
      {
        reg    = "^token_.*$"
        active = true
      }
    ]
  }

  headers = {
    max_count         = 64
    max_length        = 1024
    enable_max_count  = true
    enable_max_length = true
  }

  cookies = {
    max_count         = 32
    max_length        = 512
    enable_max_count  = true
    enable_max_length = true
  }

  path = {
    max_length        = 2048
    enable_max_length = true
  }

  decoding = {
    base64  = true
    dual    = false
    html    = false
    unicode = false
  }
}
