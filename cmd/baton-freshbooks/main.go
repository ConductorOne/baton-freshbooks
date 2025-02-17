package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-freshbooks/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-freshbooks",
		getConnector,
		field.Configuration{
			Fields: ConfigurationFields,
		},
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

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	// Get arguments from Viper
	argAccessToken := v.GetString(token)
	argRefreshToken := v.GetString(refreshToken)
	argClientID := v.GetString(fbClientID)
	argClientSecret := v.GetString(fbClientSecret)

	var connectorOpts []connector.Option

	if argAccessToken != "" {
		connectorOpts = append(connectorOpts, connector.WithAccessToken(ctx, argAccessToken))
	} else {
		if argRefreshToken != "" && argClientID != "" && argClientSecret != "" {
			connectorOpts = append(connectorOpts, connector.WithRefreshToken(ctx, argRefreshToken, argClientID, argClientSecret))
		}
	}

	if len(connectorOpts) == 0 {
		return nil, fmt.Errorf("[token] or [refresh-token, fb-client-id, fb-client-secret] argumetns must provided")
	}

	l := ctxzap.Extract(ctx)

	if err := ValidateConfig(v); err != nil {
		return nil, err
	}

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
