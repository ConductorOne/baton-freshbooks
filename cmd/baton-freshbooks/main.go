package main

import (
	"context"

	cfg "github.com/conductorone/baton-freshbooks/pkg/config"
	"github.com/conductorone/baton-freshbooks/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
)

var version = "dev"

func main() {
	ctx := context.Background()
	config.RunConnector(ctx, "baton-freshbooks", version, cfg.Config,
		connector.NewLambdaConnector,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilderV2(&connector.Connector{}),
	)
}
