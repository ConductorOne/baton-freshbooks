package connector

import (
	"context"
	"testing"

	"github.com/conductorone/baton-freshbooks/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTeamMember() client.TeamMember {
	return client.TeamMember{
		UUID:             "user-1",
		FirstName:        "Ada",
		LastName:         "Lovelace",
		Email:            "ada@example.com",
		Active:           true,
		BusinessRoleName: "owner",
	}
}

func TestUserBuilderGrants_RoleSyncEnabled_EmitsRoleGrant(t *testing.T) {
	u := newUserBuilder(nil, true)

	userResource, err := parseIntoUserResource(testTeamMember(), nil, true)
	require.NoError(t, err)

	grants, _, err := u.Grants(context.Background(), userResource, rs.SyncOpAttrs{})
	require.NoError(t, err)
	assert.Len(t, grants, 1)
}

func TestUserBuilderGrants_RoleSyncDisabled_EmitsNoGrant(t *testing.T) {
	u := newUserBuilder(nil, false)

	userResource, err := parseIntoUserResource(testTeamMember(), nil, false)
	require.NoError(t, err)

	grants, _, err := u.Grants(context.Background(), userResource, rs.SyncOpAttrs{})
	require.NoError(t, err)
	assert.Empty(t, grants)
}

func TestParseIntoUserResource_RoleSyncDisabled_SetsSkipEntitlementsAndGrants(t *testing.T) {
	userResource, err := parseIntoUserResource(testTeamMember(), nil, false)
	require.NoError(t, err)

	annos := annotations.Annotations(userResource.GetAnnotations())
	assert.True(t, annos.Contains(&v2.SkipEntitlementsAndGrants{}),
		"expected SkipEntitlementsAndGrants annotation when role sync is disabled")
}

func TestParseIntoUserResource_RoleSyncEnabled_OmitsSkipEntitlementsAndGrants(t *testing.T) {
	userResource, err := parseIntoUserResource(testTeamMember(), nil, true)
	require.NoError(t, err)

	annos := annotations.Annotations(userResource.GetAnnotations())
	assert.False(t, annos.Contains(&v2.SkipEntitlementsAndGrants{}),
		"did not expect SkipEntitlementsAndGrants annotation when role sync is enabled")
}
