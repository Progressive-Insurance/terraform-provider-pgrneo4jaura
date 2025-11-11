package pgrneo4jaura

import (
	"context"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"

	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

type notEmptyStateCheck struct {
	resourceAddress string
	attributePath   tfjsonpath.Path
}

func (n notEmptyStateCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	var resource *tfjson.StateResource

	if req.State == nil {
		resp.Error = fmt.Errorf("state is nil")
	}

	if req.State.Values == nil {
		resp.Error = fmt.Errorf("state does not contain any state values")
	}

	if req.State.Values.RootModule == nil {
		resp.Error = fmt.Errorf("state does not contain a root module")
	}

	for _, r := range req.State.Values.RootModule.Resources {
		if n.resourceAddress == r.Address {
			resource = r

			break
		}
	}

	if resource == nil {
		resp.Error = fmt.Errorf("%s - Resource not found in state", n.resourceAddress)
		return
	}

	// Traverse the attribute path to get the value
	value, err := tfjsonpath.Traverse(resource.AttributeValues, n.attributePath)
	if err != nil {
		resp.Error = fmt.Errorf("error traversing attribute path %s: %s", n.attributePath.String(), err.Error())
		return
	}

	// Check if the value is empty
	if value == nil || value == "" {
		resp.Error = fmt.Errorf("attribute %s is empty", n.attributePath.String())
	}
}

// Helper function to create the custom state check
func ExpectNotEmpty(resourceAddress string, attributePath tfjsonpath.Path) statecheck.StateCheck {
	return notEmptyStateCheck{
		resourceAddress: resourceAddress,
		attributePath:   attributePath,
	}
}
