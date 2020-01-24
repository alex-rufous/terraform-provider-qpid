package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/terraform-providers/terraform-provider-qpid/qpid"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: qpid.Provider})
}
