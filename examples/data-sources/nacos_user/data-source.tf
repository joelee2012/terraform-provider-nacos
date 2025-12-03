data "nacos_user" "example" {
  username = "example_user"
}

output "user_details" {
  value = data.nacos_user.example
}