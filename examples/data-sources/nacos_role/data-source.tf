data "nacos_role" "example" {
  role_name = "example_role"
}

output "role_details" {
  value = data.nacos_role.example
}