resource "link11waap_user" "engineer" {
  acl          = 5
  contact_name = "Jane Doe"
  email        = "jane.doe@example.com"
  mobile       = "+1234567890"
  org_id       = "my-org-id"
}

output "user_id" {
  value = link11waap_user.engineer.id
}
