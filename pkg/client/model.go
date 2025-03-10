package client

type Response struct {
	Response []TeamMember `json:"response,omitempty"`
	Metadata Meta         `json:"meta,omitempty"`
}

type Meta struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
	Total   int `json:"total,omitempty"`
}

type TeamMember struct {
	UUID                   string `json:"uuid,omitempty"`
	FirstName              string `json:"first_name,omitempty"`
	MiddleName             string `json:"middle_name,omitempty"`
	LastName               string `json:"last_name,omitempty"`
	Email                  string `json:"email,omitempty"`
	JobTitle               string `json:"job_tittle,omitempty"`
	Street1                string `json:"street_1,omitempty"`
	Street2                string `json:"street_2,omitempty"`
	City                   string `json:"city,omitempty"`
	Province               string `json:"province,omitempty"`
	Country                string `json:"country,omitempty"`
	PostalCode             string `json:"postal_code,omitempty"`
	CountryCode            string `json:"country_code,omitempty"`
	PhoneNumber            string `json:"phone_number,omitempty"`
	BusinessID             int    `json:"business_id,omitempty"`
	BusinessRoleName       string `json:"business_role_name,omitempty"`
	Active                 bool   `json:"active,omitempty"`
	IdentityId             int    `json:"identity_it,omitempty"`
	InvitationDateAccepted string `json:"invitation_date_accepted,omitempty"`
	CreatedAt              string `json:"created_at,omitempty"`
	UpdatedAt              string `json:"updated_at,omitempty"`
	Invited                bool   `json:"invited,omitempty"`
}

type Role struct {
	RoleName         string
	BusinessRoleName string
}

type ResponseBID struct {
	Response UserResponse `json:"response"`
}

type UserResponse struct {
	ID                  int64                `json:"id"`
	IdentityID          int64                `json:"identity_id"`
	IdentityUUID        string               `json:"identity_uuid"`
	BusinessMemberships []BusinessMembership `json:"business_memberships"`
}

type BusinessMembership struct {
	ID       int64    `json:"id"`
	Business Business `json:"business"`
}

type Business struct {
	ID           int64  `json:"id"`
	BusinessUUID string `json:"business_uuid"`
	Name         string `json:"name"`
}
