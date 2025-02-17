package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

const (
	token          = "token"
	refreshToken   = "refresh-token"
	fbClientID     = "fb-client-id"
	fbClientSecret = "fb-client-secret"
)

var (
	TokenField        = field.StringField(token, field.WithRequired(false), field.WithDescription("Token to request data from the FreshBooks APIs"))
	RefreshTokenField = field.StringField(refreshToken, field.WithRequired(false), field.WithDescription("Refresh token used to get a new access token from FreshBooks"))
	ClientIDField     = field.StringField(fbClientID, field.WithRequired(false), field.WithDescription("Refresh token used to get a new access token from FreshBooks"))
	ClientSecretField = field.StringField(fbClientSecret, field.WithRequired(false), field.WithDescription("Refresh token used to get a new access token from FreshBooks"))

	// ConfigurationFields defines the external configuration required for the
	// connector to run. Note: these fields can be marked as optional or
	// required.
	ConfigurationFields = []field.SchemaField{TokenField, RefreshTokenField, ClientIDField, ClientSecretField}

	// FieldRelationships defines relationships between the fields listed in
	// ConfigurationFields that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{
		field.FieldsAtLeastOneUsed(TokenField, RefreshTokenField),
		field.FieldsRequiredTogether(ClientIDField, ClientSecretField),
	}

	// ConfigurationSchema = field.NewConfiguration(ConfigurationFields, FieldRelationships...)
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(v *viper.Viper) error {
	return nil
}
