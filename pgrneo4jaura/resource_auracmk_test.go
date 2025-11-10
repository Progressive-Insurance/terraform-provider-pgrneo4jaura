package pgrneo4jaura

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccPGRNeo4jCMK(t *testing.T) {
	t.Parallel()

	tenantID := "00000000-0000-0000-0000-000000000000"
	keyID := "arn:aws:kms:us-east-1:123456789012:key/mrk-00000000000000000000000000000000"

	// quick test
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccCheckPGRNeo4jCMKConfig(1, tenantID, keyID),
				ConfigStateChecks: []statecheck.StateCheck{
					ExpectNotEmpty(
						"pgrneo4jaura_auracmk.cmk",
						tfjsonpath.New("id"),
					),
					ExpectNotEmpty(
						"pgrneo4jaura_auracmk.cmk",
						tfjsonpath.New("created"),
					),
				},
			},
		},
	})
}

func testAccCheckPGRNeo4jCMKConfig(testid int, tenant_id string, keyId string) string {
	name := ""
	if testid == 1 {
		name = "mycmk"
	}
	return fmt.Sprintf(`
	resource "pgrneo4jaura_auracmk" "cmk" {
		tenant_id = "%s"
		cloud_provider = "aws"
		instance_type = "enterprise-db"
		name = "%s"
		region = "us-east-1"
		key_id = "%s"
	}`, tenant_id, name, keyId)
}
