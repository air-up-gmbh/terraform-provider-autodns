package provider

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var zoneID = os.Getenv("TF_AUTODNS_ZONE_ID")

var testDataApexRecord = `
resource "autodns_record" "test" {
	zone_id = "` + zoneID + `"
  name  = ""
  ttl   = 60
  type  = "A"
  values = ["2.2.2.2"]
}
`
var testDataApexRecordUpdated = `
resource "autodns_record" "test" {
	zone_id = "` + zoneID + `"
  name  = ""
  ttl   = 90
  type  = "A"
  values = ["4.4.4.4"]
}
`

var testDataARecord = `
resource "autodns_record" "test" {
	zone_id = "` + zoneID + `"
  name  = "acctest_a"
  ttl   = 60
  type  = "A"
  values = ["1.1.1.1"]
}
`

var testDataARecordUpdated = `
resource "autodns_record" "test" {
	zone_id = "` + zoneID + `"
  name  = "acctest_a"
  ttl   = 90
  type  = "A"
  values = ["2.2.2.2"]
}
`

var testDataTXTRecord = `
resource "autodns_record" "test" {
	zone_id = "` + zoneID + `"
  name  = "acctest_txt"
  ttl   = 60
  type  = "TXT"
  values = ["foo", "baz"]
}
`

var testDataTXTRecordUpdated = `
resource "autodns_record" "test" {
	zone_id = "` + zoneID + `"
  name  = "acctest_txt"
  ttl   = 60
  type  = "TXT"
  values = ["foo", "bar", "baz"]
}
`

var testDataMXRecord = `
resource "autodns_record" "test" {
	zone_id = "` + zoneID + `"
  name  = "acctest_mx"
  ttl   = 60
  type  = "MX"
  values = ["10 foo", "30 baz"]
}
`

var testDataMXRecordUpdated = `
resource "autodns_record" "test" {
	zone_id = "` + zoneID + `"
  name  = "acctest_mx"
  ttl   = 60
  type  = "MX"
  values = ["10 foo", "20 bar", "30 baz"]
}
`

var testDataMXBadRecord = `
resource "autodns_record" "test" {
  zone_id = "` + zoneID + `"

  name   = "acctest_mx_bad"
  ttl    = 60
  type   = "MX"
  values = ["foo.bar"]
}
`

func TestAccRecordResourceApexRecord(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testDataApexRecord,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID+"____A")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("name"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("ttl"), knownvalue.Int32Exact(60)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("2.2.2.2"),
					})),
				},
			},
			// There should be no changes if we try with the same data
			{
				Config: testDataApexRecord,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// ImportState testing
			{
				ResourceName:      "autodns_record.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testDataApexRecordUpdated,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID+"____A")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("name"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("ttl"), knownvalue.Int32Exact(90)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("4.4.4.4"),
					})),
				},
			},
			// There should be no changes if we try with the same data
			{
				Config: testDataApexRecordUpdated,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func TestAccRecordResourceARecord(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testDataARecord,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID+"__acctest_a__A")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("name"), knownvalue.StringExact("acctest_a")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("ttl"), knownvalue.Int32Exact(60)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("1.1.1.1"),
					})),
				},
			},
			// There should be no changes if we try with the same data
			{
				Config: testDataARecord,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// ImportState testing
			{
				ResourceName:      "autodns_record.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testDataARecordUpdated,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID+"__acctest_a__A")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("name"), knownvalue.StringExact("acctest_a")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("ttl"), knownvalue.Int32Exact(90)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("2.2.2.2"),
					})),
				},
			},
			// There should be no changes if we try with the same data
			{
				Config: testDataARecordUpdated,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRecordResourceTXTRecord(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testDataTXTRecord,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID+"__acctest_txt__TXT")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("name"), knownvalue.StringExact("acctest_txt")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("ttl"), knownvalue.Int32Exact(60)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("type"), knownvalue.StringExact("TXT")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("foo"),
						knownvalue.StringExact("baz"),
					})),
				},
			},
			// There should be no changes if we try with the same data
			{
				Config: testDataTXTRecord,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// ImportState testing
			{
				ResourceName:      "autodns_record.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testDataTXTRecordUpdated,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID+"__acctest_txt__TXT")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("name"), knownvalue.StringExact("acctest_txt")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("ttl"), knownvalue.Int32Exact(60)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("type"), knownvalue.StringExact("TXT")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListSizeExact(3)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("foo"),
						knownvalue.StringExact("bar"),
						knownvalue.StringExact("baz"),
					})),
				},
			},
			// There should be no changes if we try with the same data
			{
				Config: testDataTXTRecordUpdated,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRecordResourceMXRecord(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testDataMXRecord,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID+"__acctest_mx__MX")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("name"), knownvalue.StringExact("acctest_mx")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("ttl"), knownvalue.Int32Exact(60)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("type"), knownvalue.StringExact("MX")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("10 foo"),
						knownvalue.StringExact("30 baz"),
					})),
				},
			},
			// There should be no changes if we try with the same data
			{
				Config: testDataMXRecord,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// ImportState testing
			{
				ResourceName:      "autodns_record.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testDataMXRecordUpdated,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID+"__acctest_mx__MX")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("name"), knownvalue.StringExact("acctest_mx")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("ttl"), knownvalue.Int32Exact(60)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("type"), knownvalue.StringExact("MX")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListSizeExact(3)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("10 foo"),
						knownvalue.StringExact("20 bar"),
						knownvalue.StringExact("30 baz"),
					})),
				},
			},
			// There should be no changes if we try with the same data
			{
				Config: testDataMXRecordUpdated,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
func TestAccRecordResourceMXToARecord(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testDataMXRecord,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID+"__acctest_mx__MX")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("name"), knownvalue.StringExact("acctest_mx")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("ttl"), knownvalue.Int32Exact(60)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("type"), knownvalue.StringExact("MX")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("10 foo"),
						knownvalue.StringExact("30 baz"),
					})),
				},
			},
			// Update and Read testing
			{
				Config: testDataARecordUpdated,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("id"), knownvalue.StringExact(zoneID+"__acctest_a__A")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("name"), knownvalue.StringExact("acctest_a")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("ttl"), knownvalue.Int32Exact(90)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("type"), knownvalue.StringExact("A")),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("autodns_record.test", tfjsonpath.New("values"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("2.2.2.2"),
					})),
				},
			},
			// There should be no changes if we try with the same data
			{
				Config: testDataARecordUpdated,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRecordResourceMXBadRecord(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:      testDataMXBadRecord,
				ExpectError: regexp.MustCompile(".*MX, SRV, NAPTR format is:.*"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
