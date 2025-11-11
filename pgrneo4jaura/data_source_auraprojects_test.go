package pgrneo4jaura

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// note this checks for a container with at least 1 network container in the project
func TestAccPGRNeo4jAuraProjectConfigurations(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "pgrneo4jaura_auraprojects" "projects" {
	tenant_id = "00000000-0000-0000-0000-000000000000"
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					ExpectNotEmpty(
						"data.pgrneo4jaura_auraprojects.projects",
						tfjsonpath.New("name"),
					),
					ExpectNotEmpty(
						"data.pgrneo4jaura_auraprojects.projects",
						tfjsonpath.New("instance_configurations"),
					),
				},
			},
		},
	})
}
