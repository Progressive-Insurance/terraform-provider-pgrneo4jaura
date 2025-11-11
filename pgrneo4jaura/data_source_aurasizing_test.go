package pgrneo4jaura

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// note this checks for a container with at least 1 network container in the project
func TestAccPGRNeo4jAuraSizing(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "pgrneo4jaura_aurasizing" "sizing" {
	node_count = 1000000
	relationship_count = 5000000
	instance_type = "enterprise-ds"
	algorithm_categories = ["path-finding", "community-detection"]
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					ExpectNotEmpty(
						"data.pgrneo4jaura_aurasizing.sizing",
						tfjsonpath.New("did_exceed_maximum"),
					),
					ExpectNotEmpty(
						"data.pgrneo4jaura_aurasizing.sizing",
						tfjsonpath.New("recommended_size"),
					),
					ExpectNotEmpty(
						"data.pgrneo4jaura_aurasizing.sizing",
						tfjsonpath.New("min_required_memory"),
					),
				},
			},
		},
	})
}
