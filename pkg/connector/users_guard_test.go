package connector

import (
	"context"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
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

	if hasAnno(userResourceType, &v2.SkipEntitlementsAndGrants{}) {
		t.Fatal("package-level userResourceType was mutated")
	}
}

// main.go registers a zero-value Connector{} as the capabilities factory, which
// bypasses New; it must report the unfiltered capability set.
func TestZeroValueConnector_DoesNotSkipGrants(t *testing.T) {
	for _, s := range (&Connector{}).ResourceSyncers(context.Background()) {
		rt := s.ResourceType(context.Background())
		if rt.GetId() != userResourceType.Id {
			continue
		}
		if hasAnno(rt, &v2.SkipEntitlementsAndGrants{}) {
			t.Fatal("zero-value Connector advertised SkipEntitlementsAndGrants")
		}
	}
}
