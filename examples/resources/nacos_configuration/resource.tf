resource "nacos_configuration" "example" {
  namespace_id = "some-namespace-id"
  data_id      = "some-data-id"
  group        = "DEFAULT_GROUP"
  content      = <<EOF
server:
  port: 8080
  address: 0.0.0.0
EOF
  type         = "yaml"
  description  = "managed by terraform"
  tags         = ["terraform"]
  application  = "application-name"
}
