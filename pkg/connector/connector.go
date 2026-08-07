package connector

import (
	"context"
	"fmt"
	"io"

	"github.com/conductorone/baton-freshbooks/pkg/client"
	cfg "github.com/conductorone/baton-freshbooks/pkg/config"

	"golang.org/x/oauth2"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type Connector struct {
	client *client.FreshBooksClient
	// skipRoleResourceType reports whether role is excluded from the sync
	// filter. Named for the skip condition so the zero value is safe: main.go
	// registers a zero-value Connector{} as the capabilities factory.
	skipRoleResourceType bool
}

// WithSkipRoleResourceType records that role is excluded from this sync.
func WithSkipRoleResourceType(skip bool) Option {
	return func(c *Connector) error {
		c.skipRoleResourceType = skip
		return nil
	}
}

type Option func(*Connector) error

// ResourceSyncers returns a ResourceSyncerV2 for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(_ context.Context) []connectorbuilder.ResourceSyncerV2 {
	return []connectorbuilder.ResourceSyncerV2{
		newUserBuilder(d.client, d.skipRoleResourceType),
		newRoleBuilder(d.client),
	}
}

func WithRefreshToken(ctx context.Context, refreshToken, clientID, clientSecret, baseURL string) Option {
	return func(c *Connector) error {
		var clientOpts []client.Option
		if baseURL != "" {
			clientOpts = append(clientOpts, client.WithBaseURL(baseURL))
		}
		clientOpts = append(clientOpts, client.WithRefreshToken(ctx, refreshToken, clientID, clientSecret))
		fbc, err := client.New(ctx, clientOpts...)
		if err != nil {
			return fmt.Errorf("error applying option WithRefreshToken: %w", err)
		}

		c.client = fbc
		return nil
	}
}

func WithAccessToken(ctx context.Context, accessToken, baseURL string) Option {
	return func(c *Connector) error {
		clientOpts := []client.Option{
			client.WithBearerToken(accessToken),
		}
		if baseURL != "" {
			clientOpts = append(clientOpts, client.WithBaseURL(baseURL))
		}
		fbc, err := client.New(ctx, clientOpts...)
		if err != nil {
			return fmt.Errorf("error applying option WithAccessToken: %w", err)
		}

		c.client = fbc
		return nil
	}
}

func WithTokenSource(ctx context.Context, tokenSource oauth2.TokenSource) Option {
	return func(c *Connector) error {
		fbc, err := client.New(ctx, client.WithTokenSource(tokenSource))
		if err != nil {
			return fmt.Errorf("error applying option WithTokenSource: %w", err)
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

// NewLambdaConnector is the entry point used by config.RunConnector.
func NewLambdaConnector(ctx context.Context, fbc *cfg.Freshbooks, cliOpts *cli.ConnectorOpts) (connectorbuilder.ConnectorBuilderV2, []connectorbuilder.Opt, error) {
	var connectorOpts []Option

	tokenSource := cliOpts.TokenSource

	switch {
	case tokenSource != nil:
		// Running under C1 with the OAuth2 field configured: the SDK supplies a
		// managed token source that handles refresh.
		connectorOpts = append(connectorOpts, WithTokenSource(ctx, tokenSource))
	case fbc.AccessToken != "":
		connectorOpts = append(connectorOpts, WithAccessToken(ctx, fbc.AccessToken, fbc.BaseUrl))
	case fbc.RefreshToken != "" && fbc.FreshbooksClientId != "" && fbc.FreshbooksClientSecret != "":
		connectorOpts = append(connectorOpts, WithRefreshToken(ctx, fbc.RefreshToken, fbc.FreshbooksClientId, fbc.FreshbooksClientSecret, fbc.BaseUrl))
	}

	if len(connectorOpts) == 0 {
		return nil, nil, fmt.Errorf("[token] or [refresh-token, fb-client-id, fb-client-secret] argumetns must provided")
	}

	// nil cliOpts means no filter, so nothing is skipped.
	connectorOpts = append(connectorOpts,
		WithSkipRoleResourceType(cliOpts != nil && !cliOpts.WillSyncResourceType(RoleResourceTypeID)))

	c, err := New(ctx, connectorOpts...)
	if err != nil {
		return nil, nil, err
	}
	return c, nil, nil
}
