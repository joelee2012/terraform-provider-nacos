provider "nacos" {
  host     = "https://<nacos-url>"
  username = "<nacos username>"
  password = "<nacos password>"
}

resource "nacos_namespace" "example" {
  namespace_id = "some-id"
  name         = "some-name"
  description  = "managed by terraform"
}

resource "nacos_configuration" "example" {
  namespace_id = resource.nacos_namespace.example.namespace_id
  data_id      = "configreation-test"
  group        = "DEFAULT_GROUP"
  content      = <<EOF
server:
  port: 8080
  address: 0.0.0.0
EOF
  type         = "yaml"
  description  = "test terraform"
  tags         = ["terraform"]
  application  = "terraform-nacos"
}
