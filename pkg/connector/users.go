package connector

import (
	"context"

	"github.com/conductorone/baton-freshbooks/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.FreshBooksClient
}

func (u *userBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (u *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	var rv []*v2.Resource
	err := u.client.EnsureBusinessID(ctx)
	if err != nil {
		return nil, nil, err
	}

	pToken := &opts.PageToken

	bag, pageToken, err := getToken(pToken, userResourceType)
	if err != nil {
		return nil, nil, err
	}

	teamMembers, nextPageToken, annos, err := u.client.ListTeamMembers(ctx, client.PageOptions{
		Page:    pageToken,
		PerPage: pToken.Size,
	})
	if err != nil {
		return nil, nil, err
	}

	err = bag.Next(nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	for _, teamMember := range teamMembers {
		userResource, err := parseIntoUserResource(teamMember, parentResourceID)
		if err != nil {
			return nil, nil, err
		}
		rv = append(rv, userResource)
	}

	nextPageTokenStr, err := bag.Marshal()
	if err != nil {
		return nil, nil, err
	}

	if nextPageTokenStr == "" {
		return rv, &rs.SyncOpResults{Annotations: annos}, nil
	}
	return rv, &rs.SyncOpResults{NextPageToken: nextPageTokenStr, Annotations: annos}, nil
}

// Entitlements always returns an empty slice for users.
func (u *userBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants emits the user's role assignment. The role is read from the
// business_role_name stored on the user's profile during List, so no
// additional API call (nor a cached team-member list) is needed here.
func (u *userBuilder) Grants(_ context.Context, user *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	userTrait, err := rs.GetUserTrait(user)
	if err != nil {
		return nil, nil, err
	}

	businessRoleName, ok := rs.GetProfileStringValue(userTrait.GetProfile(), "business_role_name")
	if !ok || businessRoleName == "" {
		// User has no role assignment.
		return nil, nil, nil
	}

	roleResource, err := newRoleResource(businessRoleName)
	if err != nil {
		return nil, nil, err
	}
	if roleResource == nil {
		// Unknown role name; nothing to grant.
		return nil, nil, nil
	}

	roleGrant := grant.NewGrant(roleResource, permissionName, user.Id)
	return []*v2.Grant{roleGrant}, nil, nil
}

func newUserBuilder(client *client.FreshBooksClient) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       client,
	}
}

// parseIntoUserResource parses a TeamMember (users from FreshBooks) into a User Resource.
func parseIntoUserResource(teamMember client.TeamMember, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	var userStatus = v2.UserTrait_Status_STATUS_ENABLED

	profile := map[string]interface{}{
		"uuid":                teamMember.UUID,
		"email":               teamMember.Email,
		"first_name":          teamMember.FirstName,
		"last_name":           teamMember.LastName,
		"active":              teamMember.Active,
		"invitation_accepted": teamMember.InvitationDateAccepted,
		"business_role_name":  teamMember.BusinessRoleName,
	}

	userTraits := []rs.UserTraitOption{
		rs.WithUserProfile(profile),
		rs.WithStatus(userStatus),
		rs.WithUserLogin(teamMember.Email),
		rs.WithEmail(teamMember.Email, true),
	}

	displayName := teamMember.FirstName + " " + teamMember.LastName
	if displayName == "" {
		displayName = teamMember.Email
	}

	ret, err := rs.NewUserResource(
		displayName,
		userResourceType,
		teamMember.UUID,
		userTraits,
		rs.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
