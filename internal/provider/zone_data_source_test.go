package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var zoneOrigin = os.Getenv("TF_AUTODNS_ZONE_ORIGIN")

var testAccExampleDataSourceConfig = `
data "autodns_zone" "test" {
  origin = "` + zoneOrigin + `"
}
`

func TestAccExampleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.autodns_zone.test", "id", zoneID),
					resource.TestCheckResourceAttr("data.autodns_zone.test", "origin", zoneOrigin),
					resource.TestCheckResourceAttr("data.autodns_zone.test", "name_server_group", "ns14.net"),
					resource.TestCheckResourceAttr("data.autodns_zone.test", "virtual_name_server", "a.ns14.net"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.autodns_zone.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID)),
					statecheck.ExpectKnownValue("data.autodns_zone.test", tfjsonpath.New("origin"), knownvalue.StringExact(zoneOrigin)),
					statecheck.ExpectKnownValue("data.autodns_zone.test", tfjsonpath.New("name_server_group"), knownvalue.StringExact("ns14.net")),
					statecheck.ExpectKnownValue("data.autodns_zone.test", tfjsonpath.New("virtual_name_server"), knownvalue.StringExact("a.ns14.net")),
				},
			},
		},
	})
}
