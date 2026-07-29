package admin

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) DashboardMessage() string {
	return "Admin dashboard ready"
}