package connector

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/conductorone/baton-freshbooks/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
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
	paginationToken  = &pagination.Token{Size: 50, Token: ""}
)

func TestUserBuilderListWithAcessToken(t *testing.T) {
	if accessToken == "" {
		t.Fatal("param token missing")
	}

	c, err := client.New(
		ctx,
		client.WithBearerToken(accessToken),
	)
	if err != nil {
		message = fmt.Sprintf("error creating client: %v", err)
		t.Fatal(message)
	}
	u := newUserBuilder(c)

	users, _, _, err := u.List(ctx, parentResourceID, paginationToken)
	assert.Nil(t, err)
	assert.NotNil(t, users)
	assert.Greater(t, len(users), 0)
}

func TestUserBuilderListWithRefreshToken(t *testing.T) {
	if refreshToken == "" && clientID == "" && clientSecret == "" {
		t.Fatal("the params refresh-token, fb-client-id and fb-client-secret must be used")
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
	roles, _, _, err := r.List(ctx, parentResourceID, paginationToken)
	assert.Nil(t, err)
	assert.NotNil(t, roles)
	assert.Greater(t, len(roles), 0)
}
