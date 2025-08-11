package main

import (
	"context"
	"fmt"
	"os"

	cfg "github.com/conductorone/baton-freshbooks/pkg/config"
	"github.com/conductorone/baton-freshbooks/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-github",
		getConnector,
		cfg.Config,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, fbc *cfg.Freshbooks) (types.ConnectorServer, error) {
	err := field.Validate(cfg.Config, fbc)
	if err != nil {
		return nil, err
	}

	var connectorOpts []connector.Option

	if fbc.AccessToken != "" {
		connectorOpts = append(connectorOpts, connector.WithAccessToken(ctx, fbc.AccessToken))
	} else if fbc.RefreshToken != "" && fbc.FreshbooksClientId != "" && fbc.FreshbooksClientSecret != "" {
		connectorOpts = append(connectorOpts, connector.WithRefreshToken(ctx, fbc.RefreshToken, fbc.FreshbooksClientId, fbc.FreshbooksClientSecret))
	}

	if len(connectorOpts) == 0 {
		return nil, fmt.Errorf("[token] or [refresh-token, fb-client-id, fb-client-secret] argumetns must provided")
	}

	l := ctxzap.Extract(ctx)

	cb, err := connector.New(ctx, connectorOpts...)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	connector, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	return connector, nil
}
