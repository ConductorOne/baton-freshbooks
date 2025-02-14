package connector

import (
	"context"
	"github.com/conductorone/baton-freshbooks/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
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
func (u *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var rv []*v2.Resource

	if u.client.GetBusinessID() == "" {
		businessID, err := u.client.RequestBusinessID(ctx)
		if err != nil {
			return nil, "", nil, err
		}
		u.client.SetBusinessID(businessID)
	}

	bag, pageToken, err := getToken(pToken, userResourceType)
	if err != nil {
		return nil, "", nil, err
	}

	teamMembers, nextPageToken, annotation, err := u.client.ListTeamMembers(ctx, client.PageOptions{
		Page:    pageToken,
		PerPage: pToken.Size,
	})

	if err != nil {
		return nil, "", nil, err
	}
	err = bag.Next(nextPageToken)
	if err != nil {
		return nil, "", nil, err
	}

	for _, teamMember := range teamMembers {
		teamMemberCopy := teamMember
		userResource, err := parseIntoUserResource(&teamMemberCopy, parentResourceID)
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, userResource)
	}

	nextPageToken, err = bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return rv, nextPageToken, annotation, nil
}

// Entitlements always returns an empty slice for users.
func (u *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (u *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *client.FreshBooksClient) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       client,
	}
}

// parseIntoUserResource - This function parses a TeamMember (users from FreshBooks) into a User Resource.
func parseIntoUserResource(teamMember *client.TeamMember, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	var userStatus = v2.UserTrait_Status_STATUS_ENABLED

	profile := map[string]interface{}{
		"uuid":                teamMember.UUID,
		"email":               teamMember.Email,
		"first_name":          teamMember.FirstName,
		"last_name":           teamMember.LastName,
		"active":              teamMember.Active,
		"invitation_accepted": teamMember.InvitationDateAccepted,
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
