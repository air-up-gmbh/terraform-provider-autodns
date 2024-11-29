package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate the provider during acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"autodns": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck runs pre-checks before the tests run.
func testAccPreCheck(t *testing.T) {
	if os.Getenv("AUTODNS_USERNAME") == "" || os.Getenv("AUTODNS_PASSWORD") == "" {
		t.Fatalf("Please make sure that AUTODNS_USERNAME and AUTODNS_PASSWORD environment variables are set.")
	}

	if os.Getenv("TF_AUTODNS_ZONE_ORIGIN") == "" {
		t.Fatalf("Please make sure that TF_AUTODNS_ZONE_ORIGIN environment variable is set for zone_data_source tests to run.")
	}

	if os.Getenv("TF_AUTODNS_ZONE_ID") == "" {
		t.Fatalf("Please make sure that TF_AUTODNS_ZONE_ID environment variable is set for record_resource tests to run.")
	}
}
