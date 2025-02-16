package connector

import (
	"context"
	"fmt"
	"github.com/conductorone/baton-freshbooks/pkg/client"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type Connector struct {
	client *client.FreshBooksClient
}

type Option func(*Connector) error

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(d.client),
		newRoleBuilder(d.client),
	}
}

func WithRefreshToken(ctx context.Context, refreshToken, clientID, clientSecret string) Option {
	return func(c *Connector) error {
		clientOpts := []client.Option{
			client.WithRefreshToken(refreshToken),
			client.WithClientID(clientID),
			client.WithClientSecret(clientSecret),
		}
		fbc, err := client.New(ctx, clientOpts...)
		if err != nil {
			return fmt.Errorf("error applying option WithRefreshToken: %v", err)
		}

		c.client = fbc
		return nil
	}
}

func WithAccessToken(ctx context.Context, accessToken string) Option {
	return func(c *Connector) error {
		fbc, err := client.New(ctx, client.WithBearerToken(accessToken))
		if err != nil {
			return fmt.Errorf("error applying option WithAccessToken: %v", err)
		}

		c.client = fbc
		return nil
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Baton-FreshBooks Connector",
		Description: "Connector to sync data from the FreshBooks Platform",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(_ context.Context, opts ...Option) (*Connector, error) {
	connector := &Connector{}
	for _, opt := range opts {
		err := opt(connector)
		if err != nil {
			return nil, err
		}
	}

	return connector, nil
}
