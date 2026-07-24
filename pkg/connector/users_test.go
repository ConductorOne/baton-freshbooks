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

// Grants itself is unconditional regardless of the sync filter; the SDK
// decides whether to call it at all based on the annotation ResourceType
// sets, exercised below.
func TestUserBuilderGrants_EmitsRoleGrant(t *testing.T) {
	u := newUserBuilder(nil, true)

	userResource, err := parseIntoUserResource(testTeamMember(), nil)
	require.NoError(t, err)

	grants, _, err := u.Grants(context.Background(), userResource, rs.SyncOpAttrs{})
	require.NoError(t, err)
	assert.Len(t, grants, 1)
}

func TestUserBuilderResourceType_RoleSyncEnabled_SetsSkipEntitlements(t *testing.T) {
	u := newUserBuilder(nil, true)

	rt := u.ResourceType(context.Background())

	annos := annotations.Annotations(rt.GetAnnotations())
	assert.True(t, annos.Contains(&v2.SkipEntitlements{}),
		"expected SkipEntitlements annotation when role sync is enabled")
	assert.False(t, annos.Contains(&v2.SkipEntitlementsAndGrants{}),
		"did not expect SkipEntitlementsAndGrants annotation when role sync is enabled, since Grants still needs to run")
}

func TestUserBuilderResourceType_RoleSyncDisabled_SetsSkipEntitlementsAndGrants(t *testing.T) {
	u := newUserBuilder(nil, false)

	rt := u.ResourceType(context.Background())

	annos := annotations.Annotations(rt.GetAnnotations())
	assert.True(t, annos.Contains(&v2.SkipEntitlementsAndGrants{}),
		"expected SkipEntitlementsAndGrants annotation when role sync is disabled")
}

func TestUserBuilderResourceType_DoesNotMutatePackageLevelVar(t *testing.T) {
	u := newUserBuilder(nil, false)
	_ = u.ResourceType(context.Background())

	assert.Empty(t, userResourceType.GetAnnotations(),
		"ResourceType must not mutate the shared package-level userResourceType var")
}
