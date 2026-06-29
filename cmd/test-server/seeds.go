package main

import "fmt"

// roleEmployee is the business_role_name for the "employee" role. It is used
// for several seed members, so it lives in a constant.
const roleEmployee = "business_employee"

// seed populates the test server with team members covering every grant path
// the connector exercises. The business_role_name values mirror the
// connector's static role list (pkg/connector/roles.go):
//
//	owner             -> admin
//	business_manager  -> manager
//	business_employee -> employee
//	contractor        -> contractor
//	no_seat_employee  -> accountant
//
// Invariant: every role grant the connector emits in production must have at
// least one seed member that produces it.
func seed(s *State) {
	s.business = business{
		ID:           12345,
		BusinessUUID: "biz-uuid-12345",
		Name:         "Acme Test Co",
	}

	// Diverse members: one per known role plus edge cases.
	members := []*teamMember{
		{UUID: "tm-0001", FirstName: "Alice", LastName: "Owner", Email: "alice@example.com", BusinessRoleName: "owner", Active: true, InvitationDateAccepted: "2023-01-01"},
		{UUID: "tm-0002", FirstName: "Bob", LastName: "Manager", Email: "bob@example.com", BusinessRoleName: "business_manager", Active: true, InvitationDateAccepted: "2023-01-02"},
		{UUID: "tm-0003", FirstName: "Carol", LastName: "Employee", Email: "carol@example.com", BusinessRoleName: roleEmployee, Active: true, InvitationDateAccepted: "2023-01-03"},
		{UUID: "tm-0004", FirstName: "Dave", LastName: "Contractor", Email: "dave@example.com", BusinessRoleName: "contractor", Active: true, InvitationDateAccepted: "2023-01-04"},
		{UUID: "tm-0005", FirstName: "Eve", LastName: "Accountant", Email: "eve@example.com", BusinessRoleName: "no_seat_employee", Active: true, InvitationDateAccepted: "2023-01-05"},
		// Second employee — exercises a role assigned to >= 2 users. Inactive
		// to cover a deactivated member.
		{UUID: "tm-0006", FirstName: "Frank", LastName: "Employee2", Email: "frank@example.com", BusinessRoleName: roleEmployee, Active: false, InvitationDateAccepted: "2023-01-06"},
		// No role — exercises the empty-grants path (no grant emitted).
		{UUID: "tm-0007", FirstName: "Grace", LastName: "NoRole", Email: "grace@example.com", BusinessRoleName: "", Active: true},
		// Unknown role — exercises newRoleResource returning nil (no grant).
		{UUID: "tm-0008", FirstName: "Heidi", LastName: "Unknown", Email: "heidi@example.com", BusinessRoleName: "ghost_role", Active: true, InvitationDateAccepted: "2023-01-08"},
	}
	for _, m := range members {
		m.BusinessID = int(s.business.ID)
		s.addMember(m)
	}

	// Filler members to push the total past the connector's page size (50) so
	// List pagination crosses a page boundary. They carry a known role so they
	// also emit grants.
	for i := 9; i <= 55; i++ {
		s.addMember(&teamMember{
			UUID:             fmt.Sprintf("tm-%04d", i),
			FirstName:        fmt.Sprintf("User%d", i),
			LastName:         "Filler",
			Email:            fmt.Sprintf("user%d@example.com", i),
			BusinessID:       int(s.business.ID),
			BusinessRoleName: roleEmployee,
			Active:           true,
		})
	}
}
