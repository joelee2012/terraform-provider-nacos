resource "nacos_permission" "example" {
  role_name = "some-role-name"
  resource  = "<namespace>:*:*"
  action    = "r"
}
