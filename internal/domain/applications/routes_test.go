package applications

import (
	"net/http"
	"sort"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes(t *testing.T) {
	router := chi.NewRouter()
	handler := &Handler{}

	authMiddleware := func(next http.Handler) http.Handler {
		return next
	}

	RegisterRoutes(router, handler, authMiddleware)

	var registeredRoutes []string

	err := chi.Walk(
		router,
		func(
			method string,
			route string,
			handler http.Handler,
			middlewares ...func(http.Handler) http.Handler,
		) error {
			registeredRoutes = append(
				registeredRoutes,
				method+" "+route,
			)

			return nil
		},
	)

	if err != nil {
		t.Fatalf("failed to walk application routes: %v", err)
	}

	expectedRoutes := []string{
		http.MethodGet + " /applications/{id}",
		http.MethodGet + " /applications/jobs/{jobID}",
		http.MethodGet + " /applications/mine",
		http.MethodPatch + " /applications/{id}/status",
		http.MethodPost + " /applications/jobs/{jobID}",
	}

	sort.Strings(registeredRoutes)
	sort.Strings(expectedRoutes)

	if len(registeredRoutes) != len(expectedRoutes) {
		t.Fatalf(
			"expected %d registered routes, got %d: %v",
			len(expectedRoutes),
			len(registeredRoutes),
			registeredRoutes,
		)
	}

	for index := range expectedRoutes {
		if registeredRoutes[index] != expectedRoutes[index] {
			t.Errorf(
				"expected route %q, got %q",
				expectedRoutes[index],
				registeredRoutes[index],
			)
		}
	}
}
