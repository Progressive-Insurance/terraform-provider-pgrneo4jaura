package main

import (
	"context"
	"terraform-provider-pgrneo4jaura/pgrneo4jaura"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	providerserver.Serve(context.Background(), pgrneo4jaura.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/progressive/pgrneo4jaura",
	})
}
