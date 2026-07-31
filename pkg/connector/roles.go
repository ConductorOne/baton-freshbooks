package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"

	"github.com/conductorone/baton-freshbooks/pkg/client"
)

const permissionName = "assigned"

// availableRoles is the fixed set of FreshBooks roles. They are not creatable
// or modifiable via the API, so we declare them statically.
var availableRoles = []client.Role{
	{RoleName: "admin", BusinessRoleName: "owner"},                 // Admin Role.
	{RoleName: "manager", BusinessRoleName: "business_manager"},    // Manager Role.
	{RoleName: "employee", BusinessRoleName: "business_employee"},  // Employee Role.
	{RoleName: "contractor", BusinessRoleName: "contractor"},       // Contractor Role.
	{RoleName: "accountant", BusinessRoleName: "no_seat_employee"}, // Accountant Role.
}

type roleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.FreshBooksClient
}

func (r *roleBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return roleResourceType
}

// List retrieves a hardcoded list of available Roles, since they are fixed (not modifications neither creation allowed by the platform) and cannot be requested to the API.
func (r *roleBuilder) List(_ context.Context, _ *v2.ResourceId, _ rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	var ret []*v2.Resource
	for _, role := range availableRoles {
		roleResource, err := parseIntoRoleResource(role, nil)
		if err != nil {
			return nil, nil, err
		}

		ret = append(ret, roleResource)
	}

	return ret, nil, nil
}

func (r *roleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	var ret []*v2.Entitlement

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(resource.Description),
		entitlement.WithDisplayName(resource.DisplayName),
	}
	ret = append(ret, entitlement.NewPermissionEntitlement(resource, permissionName, assigmentOptions...))

	return ret, nil, nil
}

// Grants for role assignments are emitted from userBuilder.Grants instead of
// here: each team member carries its role in business_role_name, which is
// stored on the user resource during sync. Emitting from the user side avoids
// re-paginating the full team-member list (already fetched by userBuilder.List)
// and keeps the connector stateless across lambda invocations.
func (r *roleBuilder) Grants(_ context.Context, _ *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	return nil, nil, nil
}

// parseIntoRoleResource parses a role from FreshBooks into a Role Resource.
func parseIntoRoleResource(role client.Role, _ *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"id":   role.BusinessRoleName,
		"name": role.RoleName,
	}

	roleTraits := []rs.RoleTraitOption{}

	ret, err := rs.NewRoleResource(role.RoleName, roleResourceType, role.BusinessRoleName, roleTraits, rs.WithResourceProfile(profile))
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// newRoleResource builds the role resource for a given business_role_name, or
// returns nil when the name does not map to a known role.
func newRoleResource(businessRoleName string) (*v2.Resource, error) {
	for _, role := range availableRoles {
		if role.BusinessRoleName == businessRoleName {
			return parseIntoRoleResource(role, nil)
		}
	}

	return nil, nil
}

func newRoleBuilder(c *client.FreshBooksClient) *roleBuilder {
	return &roleBuilder{
		resourceType: roleResourceType,
		client:       c,
	}
}
