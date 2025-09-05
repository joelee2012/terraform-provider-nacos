[![test](https://github.com/joelee2012/terraform-provider-nacos/actions/workflows/test.yml/badge.svg)](https://github.com/joelee2012/terraform-provider-nacos/actions/workflows/test.yml)
[![goreleaser](https://github.com/joelee2012/terraform-provider-nacos/actions/workflows/release.yml/badge.svg)](https://github.com/joelee2012/terraform-provider-nacos/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/joelee2012/terraform-provider-nacos/graph/badge.svg?token=PY470EX7J6)](https://codecov.io/gh/joelee2012/terraform-provider-nacos)


# Terraform Provider for Nacos


## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23
- [Nacos](https://nacos.io/), version must be compatible with [api v1](https://nacos.io/docs/v1/open-api/?spm=5238cd80.2ef5001f.0.0.3f613b7cibLcyN)

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

HCL example:
```terraform
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

resource "nacos_user" "test" {
  username = "user1"
  password = "abcd2"
}

resource "nacos_role" "test" {
  name     = "role1"
  username = nacos_user.test.username
}

resource "nacos_permission" "test" {
  role_name  = nacos_role.test.name
  resource   = "${resource.nacos_namespace.example.namespace_id}:*:*"
  action = "r"
}

```

more examples can be found in [examples](examples) directory.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
