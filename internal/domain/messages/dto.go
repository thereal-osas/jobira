package messages

type SendMessageRequest struct {
	ReceiverID	uint	`json:"receiver_id"`
	Content 	string	`json:"content"`
}
