package connector

import (
	"context"
	"testing"

	"github.com/conductorone/baton-freshbooks/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"google.golang.org/protobuf/proto"
)

func hasAnno(rt *v2.ResourceType, msg proto.Message) bool {
	for _, a := range rt.GetAnnotations() {
		if a.MessageIs(msg) {
			return true
		}
	}
	return false
}

// The user type's only grants are cross-type role grants, so when role is
// excluded the whole grants pass is skipped.
func TestUserResourceType_SkipAnnotation(t *testing.T) {
	inScope := newUserBuilder(nil, false).ResourceType(context.Background())
	if !hasAnno(inScope, &v2.SkipEntitlements{}) || hasAnno(inScope, &v2.SkipEntitlementsAndGrants{}) {
		t.Fatalf("role in scope: want SkipEntitlements only, got %v", inScope.Annotations)
	}

	filtered := newUserBuilder(nil, true).ResourceType(context.Background())
	if !hasAnno(filtered, &v2.SkipEntitlementsAndGrants{}) {
		t.Fatalf("role filtered: want SkipEntitlementsAndGrants, got %v", filtered.Annotations)
	}

	// Both branches annotate, so check for either leaking onto the shared
	// package-level value if the proto.Clone is ever dropped.
	if hasAnno(userResourceType, &v2.SkipEntitlementsAndGrants{}) || hasAnno(userResourceType, &v2.SkipEntitlements{}) {
		t.Fatal("package-level userResourceType was mutated")
	}
}

// Grants suppresses the role grant itself when role is filtered out, rather
// than relying only on the SkipEntitlementsAndGrants annotation.
func TestUserGrants_SkipsRoleGrantWhenFiltered(t *testing.T) {
	user, err := parseIntoUserResource(client.TeamMember{
		UUID:             "1",
		Email:            "user@example.com",
		BusinessRoleName: "owner",
	}, nil)
	if err != nil {
		t.Fatalf("parseIntoUserResource: %v", err)
	}

	inScope, _, err := newUserBuilder(nil, false).Grants(context.Background(), user, rs.SyncOpAttrs{})
	if err != nil {
		t.Fatalf("role in scope: %v", err)
	}
	if len(inScope) != 1 {
		t.Fatalf("role in scope: want 1 grant, got %d", len(inScope))
	}

	filtered, _, err := newUserBuilder(nil, true).Grants(context.Background(), user, rs.SyncOpAttrs{})
	if err != nil {
		t.Fatalf("role filtered: %v", err)
	}
	if len(filtered) != 0 {
		t.Fatalf("role filtered: want 0 grants, got %d", len(filtered))
	}
}

// main.go registers a zero-value Connector{} as the capabilities factory, which
// bypasses New; it must report the unfiltered capability set.
func TestZeroValueConnector_DoesNotSkipGrants(t *testing.T) {
	var found bool
	for _, s := range (&Connector{}).ResourceSyncers(context.Background()) {
		rt := s.ResourceType(context.Background())
		if rt.GetId() != userResourceType.Id {
			continue
		}
		found = true
		if hasAnno(rt, &v2.SkipEntitlementsAndGrants{}) {
			t.Fatal("zero-value Connector advertised SkipEntitlementsAndGrants")
		}
	}
	if !found {
		t.Fatalf("user syncer missing from ResourceSyncers; nothing was asserted")
	}
}
