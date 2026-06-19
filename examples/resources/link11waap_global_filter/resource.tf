# Example: Global Filter Resource

data "link11waap_config" "main" {}

# Example 1: Block by ASN with simple value
resource "link11waap_global_filter" "block_by_asn" {
  config_id   = data.link11waap_config.main.id
  name        = "Block Suspicious ASN"
  description = "Block traffic from a known malicious ASN"
  active      = false
  tags        = ["security", "asn-block"]
  action      = "action-global-filter-block"

  rule {
    relation = "OR"

    entry {
      type    = "asn"
      value   = "12345"
      comment = "Malicious ASN"
    }

    entry {
      type    = "ip"
      value   = "192.168.1.0/24"
      comment = "Malicious IP range"
    }
  }
}

# Example 2: Monitor API traffic (uses default action: action-monitor)
resource "link11waap_global_filter" "monitor_api" {
  config_id = data.link11waap_config.main.id
  name      = "Monitor API Traffic"
  active    = true

  rule {
    relation = "OR"

    entry {
      type    = "path"
      value   = "/api/"
      comment = "API path"
    }

    # headers and cookies use name + value for the [field_name, field_value] structure
    entry {
      type    = "headers"
      name    = "content-type"
      value   = "application/json"
      comment = "JSON content type header"
    }

    entry {
      type    = "method"
      value   = "(POST|PUT|DELETE|PATCH)"
      comment = "Mutating HTTP methods"
    }
  }
}

# Example 3: Complex filter with groups
resource "link11waap_global_filter" "complex_filter" {
  config_id   = data.link11waap_config.main.id
  name        = "Complex Traffic Filter"
  description = "Filter combining multiple conditions"
  active      = true
  tags        = ["complex"]
  action      = "action-challenge"

  rule {
    relation = "AND"

    group {
      relation = "OR"

      entry {
        type    = "path"
        value   = "/admin/"
        comment = "Admin paths"
      }

      entry {
        type    = "uri"
        value   = "/.+\\.php"
        comment = "PHP files"
      }
    }
    group {
      relation = "AND"
      entry {
        type = "country"
        value = "China"
        comment = "Traffic from China"
      }
    }
  }
}

# Example 4: Using entries_json for large blocklists (e.g., IP blocklist, ASN blocklist)
resource "link11waap_global_filter" "large_ip_blocklist" {
  config_id = data.link11waap_config.main.id
  name      = "Large IP Blocklist"
  active    = true
  action    = "action-global-filter-block"

  rule {
    relation     = "OR"
    entries_json = file("${path.module}/ip_blocklist.json")
  }
}

# Where ip_blocklist.json contains:
# [
#   ["ip", "1.2.3.4",       ""],
#   ["ip", "5.6.7.8",       ""],
#   ["asn", "12345",        "Bad ASN"],
#   ["country", "XX",       ""],
#   ["ip", "192.168.1.0",   ""],
#   ["ip", "192.168.1.1",   ""],
#   ["ip", "192.168.1.2",   ""],
#   .....
# ]


# Example 5: Using entries_json for groups
resource "link11waap_global_filter" "nested_json_groups" {
  config_id   = data.link11waap_config.main.id
  name        = "JSON encoded entries example"
  description = "Global filter with complex rules using JSON encoding for entries"
  active      = false
  tags        = ["security", "asn-block"]
  action      = "action-global-filter-block"
  rule {
    relation = "AND"

    group {
      relation = "OR"
      entries_json = jsonencode([
        ["ip",      "203.0.113.0/24", "Suspicious subnet"],
        ["country", "Brazil",         "Blocked country"],
      ])
    }
    group {
      relation = "OR"
      entries_json = jsonencode([
            ["path",   "/admin/",    "Admin path"],
            ["method", "POST",       "POST method"],
      ])
    }
  }
}
