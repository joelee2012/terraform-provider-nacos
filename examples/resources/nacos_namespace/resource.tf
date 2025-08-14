resource "nacos_namespace" "example" {
  namespace_id = "some-id"
  name         = "some-name"
  description  = "managed by terraform"
}
