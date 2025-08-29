resource "nacos_namespace" "example" {
  namespace_id = "id-123"
  name         = "some-name"
  description  = "managed by terraform"
}
