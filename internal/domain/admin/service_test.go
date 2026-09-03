package admin

import "testing"

func TestNewService(t *testing.T) {
	service := NewService()

	if service == nil {
		t.Fatal("expected service")
	}
}

func TestService_DashboardMessage(t *testing.T) {
	service := NewService()

	message := service.DashboardMessage()

	expected := "Admin dashboard ready"

	if message != expected {
		t.Fatalf(
			"expected %q, got %q",
			expected,
			message,
		)
	}
}
