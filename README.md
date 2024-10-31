# Terraform Provider AutoDNS

This repository contains the source code for the AutoDNS terraform provider.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.9
- [Go](https://golang.org/doc/install) >= 1.22.7

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the make command:

```shell
make build
```

## Developing the Provider

To compile the provider, run `make install`, this will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `TF_AUTODNS_ZONE_ID="foo.dev@a.bar.net" TF_AUTODNS_ZONE_ORIGIN="foo.dev" make testacc`.

*Note:* Acceptance tests create real resources. Do not run them on your production zones.
