package pgrneo4jaura

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPGRNeo4jInstance(t *testing.T) {
	tenant_id := "00000000-0000-0000-0000-000000000000"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(1, tenant_id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "id"),
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "connection_url"),
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "storage"),
				),
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(2, tenant_id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "id"),
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "connection_url"),
				),
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(3, tenant_id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "id"),
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "connection_url"),
				),
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(4, tenant_id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "id"),
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "connection_url"),
				),
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(5, tenant_id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "id"),
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "connection_url"),
				),
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(6, tenant_id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "id"),
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "connection_url"),
				),
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(7, tenant_id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "id"),
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "connection_url"),
				),
			},
			{
				Config: providerConfig + testAccCheckPGRNeo4jInstanceConfig(8, tenant_id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "id"),
					resource.TestCheckResourceAttrSet("pgrneo4jaura_aurainstance.instance", "connection_url"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccCheckPGRNeo4jInstanceConfig(testid int, tenant_id string) string {
	name, memory, paused := "", "", ""
	if testid == 1 { // create
		name = "testprovider"
		memory = "4GB"
		paused = "false"
	} else if testid == 2 { // pause
		name = "testprovider"
		memory = "4GB"
		paused = "true"
	} else if testid == 3 { // unpause
		name = "testprovider"
		memory = "4GB"
		paused = "false"
	} else if testid == 4 { // rename
		name = "testprovider2"
		memory = "4GB"
		paused = "false"
	} else if testid == 5 { // update memory
		name = "testprovider2"
		memory = "8GB"
		paused = "false"
	} else if testid == 6 { // update memory and pause
		name = "testprovider2"
		memory = "4GB"
		paused = "true"
	} else if testid == 7 { // update memory and unpause
		name = "testprovider2"
		memory = "8GB"
		paused = "false"
	} else if testid == 8 { // update memory, pause, and rename
		name = "testprovider3"
		memory = "4GB"
		paused = "true"
	}
	return fmt.Sprintf(`
	resource "pgrneo4jaura_aurainstance" "instance" {
		tenant_id = "%s"
		name = "%s"
		type = "enterprise-db"
		version = "5"
		cloud_provider = "aws"
		region = "us-east-1"
		memory = "%s"
		paused = %s
	}`, tenant_id, name, memory, paused)
}
