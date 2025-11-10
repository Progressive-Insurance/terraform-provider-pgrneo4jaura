package pgrneo4jaura

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccPGRNeo4jInstance(t *testing.T) {
	t.Parallel()

	tenantID := "00000000-0000-0000-0000-000000000000"

	// test for graph customer_managed_key_id, vector_optimized
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(1, tenantID),
				ConfigStateChecks: []statecheck.StateCheck{
					ExpectNotEmpty(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("id"),
					),
					ExpectNotEmpty(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("metrics_integration_url"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("paused"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("name"),
						knownvalue.StringExact("testprovider"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("type"),
						knownvalue.StringExact("enterprise-db"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("memory"),
						knownvalue.StringExact("4GB"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("graph_analytics_plugin"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("vector_optimized"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("n4jusr"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectSensitiveValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("n4jpwd"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("secondary_count"),
						knownvalue.Int64Exact(0),
					),
				},
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(2, tenantID),
				ConfigStateChecks: []statecheck.StateCheck{
					ExpectNotEmpty(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("id"),
					),
					ExpectNotEmpty(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("metrics_integration_url"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("paused"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("name"),
						knownvalue.StringExact("testprovider2"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("type"),
						knownvalue.StringExact("enterprise-db"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("memory"),
						knownvalue.StringExact("8GB"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("graph_analytics_plugin"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("vector_optimized"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("n4jusr"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectSensitiveValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("n4jpwd"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("secondary_count"),
						knownvalue.Int64Exact(0),
					),
				},
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(3, tenantID),
				ConfigStateChecks: []statecheck.StateCheck{
					ExpectNotEmpty(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("id"),
					),
					ExpectNotEmpty(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("metrics_integration_url"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("paused"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("name"),
						knownvalue.StringExact("testprovider"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("type"),
						knownvalue.StringExact("enterprise-db"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("memory"),
						knownvalue.StringExact("4GB"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("graph_analytics_plugin"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("vector_optimized"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("n4jusr"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectSensitiveValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("n4jpwd"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("secondary_count"),
						knownvalue.Int64Exact(0),
					),
				},
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(4, tenantID),
				ConfigStateChecks: []statecheck.StateCheck{
					ExpectNotEmpty(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("id"),
					),
					ExpectNotEmpty(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("metrics_integration_url"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("paused"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("name"),
						knownvalue.StringExact("testprovider2"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("type"),
						knownvalue.StringExact("enterprise-db"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("memory"),
						knownvalue.StringExact("8GB"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("graph_analytics_plugin"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("vector_optimized"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("n4jusr"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectSensitiveValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("n4jpwd"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("secondary_count"),
						knownvalue.Int64Exact(0),
					),
				},
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(8, tenantID),
				ConfigStateChecks: []statecheck.StateCheck{
					ExpectNotEmpty(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("id"),
					),
					ExpectNotEmpty(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("metrics_integration_url"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("paused"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("name"),
						knownvalue.StringExact("testprovider2"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("type"),
						knownvalue.StringExact("enterprise-db"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("memory"),
						knownvalue.StringExact("8GB"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("graph_analytics_plugin"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("vector_optimized"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("n4jusr"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectSensitiveValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("n4jpwd"),
					),
					statecheck.ExpectKnownValue(
						"pgrneo4jaura_aurainstance.instance",
						tfjsonpath.New("secondary_count"),
						knownvalue.Int64Exact(2),
					),
				},
			},
		},
	})
}

func testAccCheckPGRNeo4jInstanceConfig(testid int, tenant_id string) string {
	itype, name, memory, paused, n4jusr := "enterprise-db", "", "", "", "true"
	cmk, vectorOptimized, gdsPlugin := "", "true", "false"
	secondaries := 0
	//enterprise-db tier
	if testid == 1 { // create
		name = "testprovider"
		memory = "4GB"
		paused = "false"
	} else if testid == 2 { // pause first, update memory, vector optimized, name at same time
		name = "testprovider2"
		memory = "8GB"
		vectorOptimized = "false"
		paused = "true"
	} else if testid == 3 { // unpause first, update memory, vector optimized, name at same time
		name = "testprovider"
		memory = "4GB"
		paused = "false"
		vectorOptimized = "true"
	} else if testid == 4 { // no pause/unpause, update memory, vector optimized, name at same time
		name = "testprovider2"
		memory = "8GB"
		paused = "false"
		vectorOptimized = "false"
	} else if testid == 5 { // create instance and include neo4j password in state
		name = "testproviderpwd"
		memory = "4GB"
		paused = "false"
	} else if testid == 6 { // replace instance / no neo4j user password in state
		name = "testproviderpwd"
		memory = "4GB"
		paused = "false"
		n4jusr = "false"
		// professional-db tier
	} else if testid == 7 {
		name = "testprovider-professionaldb"
		memory = "4GB"
		paused = "false"
		itype = "professional-db"
		// test sceondary_count with testid==1
	} else if testid == 8 {
		name = "testprovider2"
		memory = "8GB"
		paused = "false"
		vectorOptimized = "false"
		secondaries = 2
	}
	return fmt.Sprintf(`
	resource "pgrneo4jaura_aurainstance" "instance" {
		tenant_id = "%s"
		name = "%s"
		type = "%s"
		version = "5"
		cloud_provider = "aws"
		region = "us-east-1"
		memory = "%s"
		paused = %s
		n4jusr = %s
		customer_managed_key_id = "%s"
		vector_optimized = %s
		graph_analytics_plugin = %s
		secondary_count = %d
	}`, tenant_id, name, itype, memory, paused, n4jusr, cmk, vectorOptimized, gdsPlugin, secondaries)
}
