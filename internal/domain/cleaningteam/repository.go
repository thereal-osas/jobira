package cleaningteam

import "context"

type Repository interface {
	ListTeamMembers(ctx context.Context, clientID uint) ([]TeamMemberData, error)
}
