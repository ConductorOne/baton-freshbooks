package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

const (
	token          = "access-token"
	refreshToken   = "refresh-token"
	fbClientID     = "freshbooks-client-id"
	fbClientSecret = "freshbooks-client-secret"
	baseURL        = "base-url"
)

var (
	TokenField = field.StringField(
		token,
		field.WithRequired(false),
		field.WithDisplayName("Access Token"),
		field.WithDescription("Token to request data from the FreshBooks APIs"),
		field.WithIsSecret(true),
		field.WithExportTarget(field.ExportTargetCLIOnly),
		field.WithHidden(true),
	)
	RefreshTokenField = field.StringField(
		refreshToken,
		field.WithRequired(false),
		field.WithDisplayName("Refresh Token"),
		field.WithDescription("Refresh token used to get a new access token from FreshBooks"),
		field.WithIsSecret(true),
		field.WithExportTarget(field.ExportTargetCLIOnly),
		field.WithHidden(true),
	)

	OAuth2Field = field.Oauth2Field(
		"oauth2",
		field.WithDisplayName("OAuth2"),
		field.WithDescription("OAuth configuration for FreshBooks"),
	)

	ClientIDField = field.StringField(
		fbClientID,
		field.WithRequired(false),
		field.WithDisplayName("Client ID"),
		field.WithDescription("Client ID from the Freshbooks app"),
	)

	ClientSecretField = field.StringField(
		fbClientSecret,
		field.WithRequired(false),
		field.WithDisplayName("Client Secret"),
		field.WithDescription("Client Secret from the Freshbooks app"),
		field.WithIsSecret(true),
	)

	BaseURLField = field.StringField(
		baseURL,
		field.WithDescription("Override the FreshBooks API URL (for testing)"),
		field.WithExportTarget(field.ExportTargetCLIOnly),
		field.WithHidden(true))

	configFields = []field.SchemaField{TokenField, RefreshTokenField, OAuth2Field, ClientIDField, ClientSecretField, BaseURLField}
)

func ValidateConfig(_ *viper.Viper) error {
	return nil
}

//go:generate go run ./gen
var Config = field.NewConfiguration(configFields,
	field.WithConnectorDisplayName("FreshBooks"),
	field.WithHelpUrl("/docs/baton/freshbooks"),
	field.WithIconUrl("/static/app-icons/freshbooks.svg"))
