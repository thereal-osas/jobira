package companyaccounts

type CreateCompanyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddMemberRequest struct {
	UserID uint `json:"user_id"`
}
