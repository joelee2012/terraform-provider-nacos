resource "nacos_permission" "example" {
  role_name  = "some-role-name"
  resource   = "<namespace>:*:*"
  permission = "r"
}
