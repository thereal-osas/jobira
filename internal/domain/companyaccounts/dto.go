package companyaccounts

type CreateCompanyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddMemberRequest struct {
	UserID uint `json:"user_id"`

	Role string `json:"role"`
}

type UpdateMemberStatusRequest struct {
	Status string `json:"status"`
}

type UpdateMemberRolesRequest struct {
	Role string `json:"role"`
}
