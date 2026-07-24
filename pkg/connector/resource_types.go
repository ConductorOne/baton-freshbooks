package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

var userResourceType = &v2.ResourceType{
	Id:          "user",
	DisplayName: "User",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
}

// RoleResourceTypeID identifies the role resource type. userBuilder.Grants
// references this constant when deciding whether to emit cross-type role
// grants (see WillSyncResourceType gating in connector.go).
const RoleResourceTypeID = "role"

var roleResourceType = &v2.ResourceType{
	Id:          RoleResourceTypeID,
	DisplayName: "Role",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_ROLE},
}
