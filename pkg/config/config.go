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
		token, field.WithRequired(false),
		field.WithDisplayName("Access Token"),
		field.WithDescription("Token to request data from the FreshBooks APIs"),
	)
	RefreshTokenField = field.StringField(
		refreshToken, field.WithRequired(false),
		field.WithDisplayName("Refresh Token"),
		field.WithDescription("Refresh token used to get a new access token from FreshBooks"))

	ClientIDField = field.StringField(
		fbClientID, field.WithRequired(false),
		field.WithDisplayName("Client ID"),
		field.WithDescription("Client ID from the Freshbooks app"))

	ClientSecretField = field.StringField(
		fbClientSecret, field.WithRequired(false),
		field.WithDisplayName("Client Secret"),
		field.WithDescription("Client Secret from the Freshbooks app"))

	BaseURLField = field.StringField(
		baseURL,
		field.WithDescription("Override the FreshBooks API URL (for testing)"),
		field.WithHidden(true))

	configFields = []field.SchemaField{TokenField, RefreshTokenField, ClientIDField, ClientSecretField, BaseURLField}

	fieldRelationships = []field.SchemaFieldRelationship{
		field.FieldsAtLeastOneUsed(TokenField, RefreshTokenField),
		field.FieldsRequiredTogether(RefreshTokenField, ClientIDField, ClientSecretField),
	}
)

func ValidateConfig(_ *viper.Viper) error {
	return nil
}

//go:generate go run ./gen
var Config = field.NewConfiguration(configFields,
	field.WithConstraints(fieldRelationships...),
	field.WithConnectorDisplayName("Freshbooks v2"),
	field.WithHelpUrl("/docs/baton/freshbooks"),
	field.WithIconUrl("/static/app-icons/freshbooks.svg"))
