package connector

import (
	"context"
	"fmt"
	"io"

	"github.com/conductorone/baton-freshbooks/pkg/client"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type Freshbooks struct {
	client *client.FreshBooksClient
}

type Option func(*Freshbooks) error

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Freshbooks) ResourceSyncers(_ context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(d.client),
		newRoleBuilder(d.client),
	}
}

func WithRefreshToken(ctx context.Context, refreshToken, clientID, clientSecret string) Option {
	return func(c *Freshbooks) error {
		clientOpts := []client.Option{
			client.WithRefreshToken(ctx, refreshToken, clientID, clientSecret),
		}
		fbc, err := client.New(ctx, clientOpts...)
		if err != nil {
			return fmt.Errorf("error applying option WithRefreshToken: %w", err)
		}

		c.client = fbc
		return nil
	}
}

func WithAccessToken(ctx context.Context, accessToken string) Option {
	return func(c *Freshbooks) error {
		fbc, err := client.New(ctx, client.WithBearerToken(accessToken))
		if err != nil {
			return fmt.Errorf("error applying option WithAccessToken: %w", err)
		}

		c.client = fbc
		return nil
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Freshbooks) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Freshbooks) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Baton-FreshBooks Connector",
		Description: "Connector to sync data from the FreshBooks Platform",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Freshbooks) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(_ context.Context, opts ...Option) (*Freshbooks, error) {
	connector := &Freshbooks{}
	for _, opt := range opts {
		err := opt(connector)
		if err != nil {
			return nil, err
		}
	}

	return connector, nil
}
