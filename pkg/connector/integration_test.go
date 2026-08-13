//go:build integration

package connector

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/conductorone/baton-freshbooks/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/stretchr/testify/assert"
)

var (
	ctx              = context.Background()
	message          = ""
	accessToken, _   = os.LookupEnv("FRESHBOOKS_ACCESS_TOKEN")
	refreshToken, _  = os.LookupEnv("FRESHBOOKS_REFRESH_TOKEN")
	clientID, _      = os.LookupEnv("FRESHBOOKS_CLIENT_ID")
	clientSecret, _  = os.LookupEnv("FRESHBOOKS_CLIENT_SECRET")
	parentResourceID = &v2.ResourceId{}
	syncOpAttrs      = rs.SyncOpAttrs{PageToken: pagination.Token{Size: 50}}
)

func TestUserBuilderListWithAcessToken(t *testing.T) {
	if accessToken == "" {
		t.Skip("FRESHBOOKS_ACCESS_TOKEN not set, skipping integration test")
	}

	c, err := client.New(
		ctx,
		client.WithBearerToken(accessToken),
	)
	if err != nil {
		message = fmt.Sprintf("error creating client: %v", err)
		t.Fatal(message)
	}
	u := newUserBuilder(c, false)

	users, _, err := u.List(ctx, parentResourceID, syncOpAttrs)
	assert.Nil(t, err)
	assert.NotNil(t, users)
}

func TestUserBuilderListWithRefreshToken(t *testing.T) {
	if refreshToken == "" && clientID == "" && clientSecret == "" {
		t.Skip("FRESHBOOKS_REFRESH_TOKEN, FRESHBOOKS_CLIENT_ID, FRESHBOOKS_CLIENT_SECRET not set, skipping integration test")
	}

	c, err := client.New(
		ctx,
		client.WithRefreshToken(ctx, refreshToken, clientID, clientSecret),
	)
	if err != nil {
		message = fmt.Sprintf("error creating client: %v", err)
		t.Fatal(message)
	}

	r := newRoleBuilder(c)
	roles, _, err := r.List(ctx, parentResourceID, syncOpAttrs)
	assert.Nil(t, err)
	assert.NotNil(t, roles)
}
