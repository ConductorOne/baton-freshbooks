package main

import "sync"

// The JSON types below mirror the FreshBooks API response shapes (see the
// pkg/client model). They are defined locally so the test server replicates
// the API's documented contract independently of the connector's structs.

// teamMember mirrors a FreshBooks team member.
// Doc URL: https://www.freshbooks.com/api/team_members
type teamMember struct {
	UUID                   string `json:"uuid"`
	FirstName              string `json:"first_name"`
	LastName               string `json:"last_name"`
	Email                  string `json:"email"`
	BusinessID             int    `json:"business_id"`
	BusinessRoleName       string `json:"business_role_name"`
	Active                 bool   `json:"active"` // no omitempty: false must serialize
	InvitationDateAccepted string `json:"invitation_date_accepted,omitempty"`
}

type meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

type teamMembersResponse struct {
	Response []teamMember `json:"response"`
	Meta     meta         `json:"meta"`
}

// business mirrors the FreshBooks business object embedded in the identity
// endpoint. Doc URL: https://www.freshbooks.com/api/me_endpoint
type business struct {
	ID           int64  `json:"id"`
	BusinessUUID string `json:"business_uuid"`
	Name         string `json:"name"`
}

type businessMembership struct {
	ID       int64    `json:"id"`
	Business business `json:"business"`
}

type meUser struct {
	ID                  int64                `json:"id"`
	IdentityID          int64                `json:"identity_id"`
	IdentityUUID        string               `json:"identity_uuid"`
	BusinessMemberships []businessMembership `json:"business_memberships"`
}

type meResponse struct {
	Response meUser `json:"response"`
}

// State is the in-memory store. The connector is read-only, so after seeding
// the data never mutates; reads still go through the lock to stay safe under
// the parallel requests CI generates.
type State struct {
	mu sync.RWMutex

	business   business
	members    map[string]*teamMember // uuid -> member
	memberList []*teamMember          // insertion order for deterministic pagination
}

func NewState() *State {
	s := &State{
		members: make(map[string]*teamMember),
	}
	seed(s)
	return s
}

// Business returns a copy of the seeded business.
func (s *State) Business() business {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.business
}

// ListTeamMembers returns one 1-based page of members and the total count.
func (s *State) ListTeamMembers(page, perPage int) ([]teamMember, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.memberList)
	if page < 1 {
		page = 1
	}
	start := (page - 1) * perPage
	if start >= total {
		return []teamMember{}, total
	}
	end := start + perPage
	if end > total {
		end = total
	}

	out := make([]teamMember, 0, end-start)
	for _, m := range s.memberList[start:end] {
		out = append(out, *m)
	}
	return out, total
}

// addMember is called from seed() only, before the server starts serving, so
// it takes no lock.
func (s *State) addMember(m *teamMember) {
	cp := *m
	s.members[m.UUID] = &cp
	s.memberList = append(s.memberList, &cp)
}
