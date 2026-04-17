# Example: Planet Trusted Nets Resource
#
# This is a singleton resource per configuration: its id is always
# "__default__". Managing it replaces the full list of trusted networks
# for the given config. Each entry must reference either an IP/CIDR
# (source = "ip") or a global filter (source = "global_filter"), never
# both. Delete is a no-op because the upstream API does not expose a
# DELETE endpoint for this resource.
#
# We recommend to import the existing trusted nets configuration into Terraform state before
# managing it, to avoid accidentally wiping out existing entries.
# For example, you can use the next command for importing:
# terraform import link11waap_planet_trusted_nets.$MY_RESOURCE_NAME $CONFIG/__default__

data "link11waap_config" "main" {}

# Get some existing global filters to reference from trusted_nets entries below.
data "link11waap_global_filter" "l11_cdn_trusted_source" {
  config_id = data.link11waap_config.main.id
  name = "Secure CDN Link11 Trusted Source"
}

# Get another global filter to reference from trusted_nets entries below.
data "link11waap_global_filter" "aws" {
  config_id = data.link11waap_config.main.id
  name      = "AWS"
}

# Mixed IPs and a global_filter reference
#
# The global filter below defines a set of trusted partner IPs/ASNs.
# Referencing it from trusted_nets keeps the trusted-IP policy in sync
# with the global filter definition rather than duplicating values.

resource "link11waap_global_filter" "trusted_partners" {
  config_id   = data.link11waap_config.main.id
  name        = "Trusted Partners"
  description = "IPs and ASNs of partners whose traffic is implicitly trusted"
  active      = true
  tags        = ["trusted", "partners"]
  action      = "action-monitor"

  rule {
    relation = "OR"

    entry {
      type    = "ip"
      value   = "10.20.30.0/24"
      comment = "Partner A egress"
    }

    entry {
      type    = "asn"
      value   = "12345"
      comment = "Partner B ASN"
    }
  }
}

# The trusted nets resource itself, with a mix of IP and global filter entries
# which is applied to the planet configuration. Note that the global filter entries can be managed
# in the same Terraform configuration or imported as data sources if managed outside of Terraform.
resource "link11waap_planet_trusted_nets" "main" {
  config_id = data.link11waap_config.main.id

  # Static IP entries
  trusted_nets {
      address = "127.0.0.0/8"
      comment = "Private subnet"
      source  = "ip"
  }
  trusted_nets {
      address = "172.16.0.0/12"
      comment = "Private subnet"
      source  = "ip"
  }
  trusted_nets {
      address = "10.0.0.0/8"
      comment = "Private subnet"
      source  = "ip"
  }

  # Reference to a global filter managed in the same configuration
  trusted_nets {
    source  = "global_filter"
    gf_id   = link11waap_global_filter.trusted_partners.id
    comment = "Partner IPs/ASNs (see global filter)"
  }

  # Reference to AWS global filter managed outside of Terraform (data source)
  trusted_nets {
      comment = "AWS IPs/ASNs (see global filter)"
      gf_id   = data.link11waap_global_filter.aws.id
      source  = "global_filter"
  }

  # Reference to a global filter managed outside of Terraform (data source)
  trusted_nets {
      comment = "Secure CDN Link11 Trusted Source"
      gf_id   = data.link11waap_global_filter.l11_cdn_trusted_source.id
      source  = "global_filter"
  }
}
