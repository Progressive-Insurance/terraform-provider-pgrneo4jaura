package pgrneo4jaura

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `provider "pgrneo4jaura" {
  client_id = "<YOUR CLIENT ID>"
  client_secret = "<YOUR CLIENT SECRET>"
}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"pgrneo4jaura": providerserver.NewProtocol6WithError(New()),
	}
)
