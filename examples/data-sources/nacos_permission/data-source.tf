data "nacos_permission" "example" {
  resource = "example_resource"
  action   = "READ"
}

output "permission_details" {
  value = data.nacos_permission.example
}