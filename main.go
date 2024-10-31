package main

import (
	"context"
	"flag"
	"log"

	"terraform-provider-autodns/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// This will be set by the goreleaser.
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/air-up-gmbh/autodns",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
