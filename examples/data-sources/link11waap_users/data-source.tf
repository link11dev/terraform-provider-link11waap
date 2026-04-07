data "link11waap_users" "all" {}

output "users" {
  value = data.link11waap_users.all.users
}
