package client

import "time"

type TeamMember struct {
	UUID                   string    `json:"uuid,omitempty"`
	FirstName              string    `json:"first_name,omitempty"`
	MiddleName             string    `json:"middle_name,omitempty"`
	LastName               string    `json:"last_name,omitempty"`
	Email                  string    `json:"email,omitempty"`
	JobTitle               string    `json:"job_tittle,omitempty"`
	Street1                string    `json:"street_1,omitempty"`
	Street2                string    `json:"street_2,omitempty"`
	City                   string    `json:"city,omitempty"`
	Province               string    `json:"province,omitempty"`
	Country                string    `json:"country,omitempty"`
	PostalCode             string    `json:"postal_code,omitempty"`
	PhoneNumber            string    `json:"phone_number,omitempty"`
	BusinessID             int       `json:"business_id,omitempty"`
	BusinessRoleName       string    `json:"business_role_name,omitempty"`
	Active                 bool      `json:"active,omitempty"`
	IdentityId             int       `json:"identity_it,omitempty"`
	InvitationDateAccepted time.Time `json:"invitation_date_accepted,omitempty"`
	CreatedAt              time.Time `json:"created_at,omitempty"`
	UpdatedAt              time.Time `json:"updated_at,omitempty"`
}

type Role struct {
	RoleName         string
	BusinessRoleName string
}

var (
	adminRole      = Role{RoleName: "admin", BusinessRoleName: "owner"}
	managerRole    = Role{RoleName: "manager", BusinessRoleName: "business_manager"}
	employeeRole   = Role{RoleName: "employee", BusinessRoleName: "business_employee"}
	contractorRole = Role{RoleName: "contractor", BusinessRoleName: "contractor"}
	accountantRole = Role{RoleName: "accountant", BusinessRoleName: "no_seat_employee"}
)
