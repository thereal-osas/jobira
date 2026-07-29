package jobinvitations

type CreateJobInvitationRequest struct {
	CleanerID 	uint	`json:"cleaner_id"`
	Message 	string	`json:"message"`
}