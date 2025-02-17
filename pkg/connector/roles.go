package connector

import (
	"context"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"sync"

	"github.com/conductorone/baton-freshbooks/pkg/client"
)

const permissionName = "assigned"

type roleBuilder struct {
	resourceType     *v2.ResourceType
	teamMembers      []client.TeamMember
	teamMembersMutex sync.RWMutex
	client           *client.FreshBooksClient
}

func (r *roleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return roleResourceType
}

// List retrieves a hardcoded list of available Roles, since they are fixed (not modifications neither creation allowed by the platform) and cannot be requested to the API.
func (r *roleBuilder) List(_ context.Context, _ *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	availableRoles := []client.Role{
		{RoleName: "admin", BusinessRoleName: "owner"},                 // Admin Role.
		{RoleName: "manager", BusinessRoleName: "business_manager"},    // Manager Role.
		{RoleName: "employee", BusinessRoleName: "business_employee"},  // Employee Role.
		{RoleName: "contractor", BusinessRoleName: "contractor"},       // Contractor Role.
		{RoleName: "accountant", BusinessRoleName: "no_seat_employee"}, // Accountant Role.
	}

	var ret []*v2.Resource
	for _, role := range availableRoles {
		roleCopy := role
		roleResource, err := parseIntoRoleResource(&roleCopy, nil)
		if err != nil {
			return nil, "", nil, err
		}

		ret = append(ret, roleResource)
	}

	return ret, "", nil, nil
}

func (r *roleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var ret []*v2.Entitlement

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(resource.Description),
		entitlement.WithDisplayName(resource.DisplayName),
	}
	ret = append(ret, entitlement.NewPermissionEntitlement(resource, permissionName, assigmentOptions...))

	return ret, "", nil, nil
}

func (r *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var ret []*v2.Grant

	teamMembers, err := r.GetAllTeamMembers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	for _, teamMember := range teamMembers {
		if teamMember.BusinessRoleName == resource.Id.Resource {
			userResource, err := parseIntoUserResource(&teamMember, nil)
			if err != nil {
				return nil, "", nil, err
			}

			membershipGrant := grant.NewGrant(resource, permissionName, userResource.Id)
			ret = append(ret, membershipGrant)
		}
	}

	return ret, "", nil, nil
}

func (r *roleBuilder) GetAllTeamMembers(ctx context.Context) ([]client.TeamMember, error) {
	r.teamMembersMutex.Lock()
	defer r.teamMembersMutex.Unlock()

	var ret []client.TeamMember
	if r.teamMembers != nil && len(r.teamMembers) > 0 {
		return r.teamMembers, nil
	}

	err := r.client.EnsureBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	
	paginationToken := pagination.Token{Size: 50, Token: ""}
	for {
		bag, pageToken, err := getToken(&paginationToken, userResourceType)
		if err != nil {
			return nil, err
		}

		teamMembers, nextPageToken, _, err := r.client.ListTeamMembers(ctx, client.PageOptions{
			Page:    pageToken,
			PerPage: paginationToken.Size,
		})

		for _, tm := range teamMembers {
			ret = append(ret, tm)
		}

		err = bag.Next(nextPageToken)
		if err != nil {
			return nil, err
		}

		if nextPageToken == "" {
			break
		}
		paginationToken.Token = nextPageToken
	}

	return ret, nil
}

// This function parses a role from FreshBooks into a Role Resource.
func parseIntoRoleResource(role *client.Role, _ *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"id":   role.BusinessRoleName,
		"name": role.RoleName,
	}

	roleTraits := []rs.RoleTraitOption{
		rs.WithRoleProfile(profile),
	}

	ret, err := rs.NewRoleResource(role.RoleName, roleResourceType, role.BusinessRoleName, roleTraits)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func newRoleBuilder(c *client.FreshBooksClient) *roleBuilder {
	return &roleBuilder{
		resourceType: roleResourceType,
		client:       c,
	}
}
