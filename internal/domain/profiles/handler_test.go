package profiles

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rodrigueghenda/jobira/internal/security/identity"
)

func profilesTestAuth(
	userID uint,
	role string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := identity.WithUser(
					r.Context(),
					identity.UserIdentity{
						UserID: userID,
						Email:  "test@example.com",
						Role:   role,
					},
				)

				next.ServeHTTP(
					w,
					r.WithContext(ctx),
				)
			},
		)
	}
}

func profilesPassthroughAdmin(
	next http.Handler,
) http.Handler {
	return next
}

func newProfilesTestRouter(
	repo *MockRepository,
	userID uint,
	role string,
) *chi.Mux {
	service := NewService(repo)
	handler := NewHandler(service)

	router := chi.NewRouter()

	RegisterRoutes(
		router,
		handler,
		profilesTestAuth(userID, role),
		profilesPassthroughAdmin,
	)

	return router
}

func TestHandler_Create_Success(t *testing.T) {
	repo := &MockRepository{}

	router := newProfilesTestRouter(
		repo,
		10,
		"cleaner",
	)

	body := `{
		"bio":"Experienced cleaner",
		"location":"East London",
		"country":"UK",
		"city":"London",
		"region":"East London",
		"postcode_area":"E14",
		"availability_status":"weekends_only",
		"travel_radius_miles":20,
		"years_experience":5,
		"hourly_rate":25,
		"services_offered":"Domestic and Airbnb"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/profiles/",
		bytes.NewBufferString(body),
	)
	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	if repo.createdProfile == nil {
		t.Fatal("expected profile to be created")
	}

	if repo.createdProfile.UserID != 10 {
		t.Fatalf(
			"expected user ID 10, got %d",
			repo.createdProfile.UserID,
		)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	service := NewService(&MockRepository{})
	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profiles/",
		bytes.NewBufferString(`{}`),
	)

	res := httptest.NewRecorder()

	handler.Create(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_Create_InvalidBody(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profiles/",
		bytes.NewBufferString("{invalid"),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_Create_InvalidInput(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profiles/",
		bytes.NewBufferString(`{
			"bio":"",
			"location":"London",
			"services_offered":"Domestic"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_Create_InvalidAvailability(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profiles/",
		bytes.NewBufferString(`{
			"bio":"Cleaner",
			"location":"London",
			"services_offered":"Domestic",
			"availability_status":"whenever"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_Create_AlreadyExists(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{
			createErr: ErrProfileAlreadyExists,
		},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profiles/",
		bytes.NewBufferString(`{
			"bio":"Cleaner",
			"location":"London",
			"services_offered":"Domestic"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf(
			"expected 409 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_Create_ServiceError(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{
			createErr: errors.New("database failure"),
		},
		1,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profiles/",
		bytes.NewBufferString(`{
			"bio":"Cleaner",
			"location":"London",
			"services_offered":"Domestic"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetMine_Success(t *testing.T) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			ID:     5,
			UserID: 10,
			Bio:    "Cleaner",
		},
	}

	router := newProfilesTestRouter(
		repo,
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	var profile CleanerProfile

	if err := json.NewDecoder(res.Body).Decode(&profile); err != nil {
		t.Fatalf(
			"failed decoding profile: %v",
			err,
		)
	}

	if profile.ID != 5 {
		t.Fatalf(
			"expected profile ID 5 got %d",
			profile.ID,
		)
	}
}

func TestHandler_GetMine_Unauthorized(t *testing.T) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me",
		nil,
	)

	res := httptest.NewRecorder()

	handler.GetMine(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetMine_NotFound(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{
			getByUserIDErr: ErrProfileNotFound,
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf(
			"expected 404 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetMine_ServiceError(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{
			getByUserIDErr: errors.New("database failure"),
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetMyProfileStrength_Success(t *testing.T) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			UserID:             10,
			Bio:                "Experienced cleaner",
			Location:           "East London",
			ServicesOffered:    "Domestic cleaning",
			AvailabilityStatus: "available_immediately",
			HourlyRate:         25,
			YearsExperience:    5,
			TravelRadiusMiles:  20,
			IsVerified:         true,
			VerificationStatus: "verified",
		},
	}

	router := newProfilesTestRouter(
		repo,
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me/strength",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	var result ProfileStrength

	if err := json.NewDecoder(
		res.Body,
	).Decode(
		&result,
	); err != nil {
		t.Fatalf(
			"failed decoding profile strength: %v",
			err,
		)
	}

	if result.Percentage != 90 {
		t.Fatalf(
			"expected percentage 90 got %d",
			result.Percentage,
		)
	}

	if result.IsComplete {
		t.Fatal("expected profile to be incomplete")
	}

	if result.CompletedItems != 8 {
		t.Fatalf(
			"expected 8 completed items got %d",
			result.CompletedItems,
		)
	}
}

func TestHandler_GetMyProfileStrength_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me/strength",
		nil,
	)

	res := httptest.NewRecorder()

	handler.GetMyProfileStrength(
		res,
		req,
	)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetMyProfileStrength_NotFound(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{
			getByUserIDErr: ErrProfileNotFound,
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me/strength",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if res.Code != http.StatusNotFound {
		t.Fatalf(
			"expected 404 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_GetMyProfileStrength_ServiceError(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{
			getByUserIDErr: errors.New(
				"database failure",
			),
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me/strength",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	existing := &CleanerProfile{
		ID:     5,
		UserID: 10,
	}

	repo := &MockRepository{
		profile: existing,
	}

	router := newProfilesTestRouter(
		repo,
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/profiles/me",
		bytes.NewBufferString(`{
			"bio":"Updated profile",
			"location":"London",
			"availability_status":"weekends_only",
			"travel_radius_miles":15,
			"services_offered":"Domestic"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	if repo.updatedProfile == nil {
		t.Fatal("expected profile update")
	}
}

func TestHandler_Update_InvalidBody(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/profiles/me",
		bytes.NewBufferString("{bad"),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_Update_InvalidAvailability(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/profiles/me",
		bytes.NewBufferString(`{
			"availability_status":"never-ever"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{
			getByUserIDErr: ErrProfileNotFound,
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/profiles/me",
		bytes.NewBufferString(`{}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf(
			"expected 404 got %d",
			res.Code,
		)
	}
}

func TestHandler_UpdateVerificationStatus_Success(t *testing.T) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			UserID:             15,
			IsVerified:         true,
			VerificationStatus: "verified",
		},
	}

	router := newProfilesTestRouter(
		repo,
		1,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/profiles/15/verification",
		bytes.NewBufferString(`{
			"verification_status":"verified"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	if repo.lastVerificationUserID != 15 {
		t.Fatalf(
			"expected user ID 15 got %d",
			repo.lastVerificationUserID,
		)
	}

	if !repo.lastVerificationVerified {
		t.Fatal("expected verified=true")
	}
}

func TestHandler_UpdateVerificationStatus_InvalidUserID(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{},
		1,
		"admin",
	)

	tests := []string{
		"/profiles/abc/verification",
		"/profiles/0/verification",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPatch,
				path,
				bytes.NewBufferString(`{
					"verification_status":"verified"
				}`),
			)

			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected 400 got %d",
					res.Code,
				)
			}
		})
	}
}

func TestHandler_UpdateVerificationStatus_InvalidBody(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{},
		1,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/profiles/15/verification",
		bytes.NewBufferString("{bad"),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_UpdateVerificationStatus_InvalidState(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{},
		1,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/profiles/15/verification",
		bytes.NewBufferString(`{
			"verification_status":"maybe"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d",
			res.Code,
		)
	}
}

func TestHandler_UpdateVerificationStatus_NotFound(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{
			updateVerificationStatusErr: ErrProfileNotFound,
		},
		1,
		"admin",
	)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/profiles/99/verification",
		bytes.NewBufferString(`{
			"verification_status":"verified"
		}`),
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf(
			"expected 404 got %d",
			res.Code,
		)
	}
}

func TestHandler_Search_Success(t *testing.T) {
	verified := true

	repo := &MockRepository{
		profiles: []CleanerProfile{
			{
				ID:                 5,
				UserID:             20,
				Country:            "UK",
				City:               "London",
				Region:             "East",
				PostcodeArea:       "E14",
				AvailabilityStatus: "weekends_only",
				ServicesOffered:    "Airbnb",
				YearsExperience:    5,
				HourlyRate:         25,
				TravelRadiusMiles:  10,
				IsVerified:         true,
			},
		},
	}

	router := newProfilesTestRouter(
		repo,
		10,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/search?country=UK&city=London"+
			"&region=East&postcode_area=E14"+
			"&availability_status=weekends_only"+
			"&services_offered=Airbnb"+
			"&min_experience=3"+
			"&max_hourly_rate=30"+
			"&max_travel_radius=20"+
			"&is_verified=true",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	if repo.lastSearchRequest.ClientID != 10 {
		t.Fatalf(
			"expected client ID 10 got %d",
			repo.lastSearchRequest.ClientID,
		)
	}

	if repo.lastSearchRequest.MinExperience != 3 {
		t.Fatalf(
			"expected min experience 3 got %d",
			repo.lastSearchRequest.MinExperience,
		)
	}

	if repo.lastSearchRequest.MaxHourlyRate != 30 {
		t.Fatalf(
			"expected max rate 30 got %d",
			repo.lastSearchRequest.MaxHourlyRate,
		)
	}

	if repo.lastSearchRequest.MaxTravelRadius != 20 {
		t.Fatalf(
			"expected radius 20 got %d",
			repo.lastSearchRequest.MaxTravelRadius,
		)
	}

	if repo.lastSearchRequest.IsVerified == nil ||
		*repo.lastSearchRequest.IsVerified != verified {
		t.Fatal("expected verified=true")
	}
}

func TestHandler_Search_InvalidQueryUsesDefaults(
	t *testing.T,
) {
	repo := &MockRepository{}

	router := newProfilesTestRouter(
		repo,
		10,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/search?min_experience=nope"+
			"&max_hourly_rate=nope"+
			"&max_travel_radius=nope"+
			"&is_verified=nope",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	if repo.lastSearchRequest.MinExperience != 0 {
		t.Fatal("expected invalid experience to use 0")
	}

	if repo.lastSearchRequest.MaxHourlyRate != 0 {
		t.Fatal("expected invalid rate to use 0")
	}

	if repo.lastSearchRequest.MaxTravelRadius != 0 {
		t.Fatal("expected invalid radius to use 0")
	}

	if repo.lastSearchRequest.IsVerified != nil {
		t.Fatal("expected invalid verified value to be ignored")
	}
}

func TestHandler_Search_ServiceError(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{
			searchErr: errors.New("search failure"),
		},
		10,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/search",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetCompletedJobs_Success(t *testing.T) {
	now := time.Now()

	repo := &MockRepository{
		history: []JobHistoryItem{
			{
				ApplicationID: 1,
				JobID:         5,
				Status:        "completed",
				CreatedAt:     now,
			},
		},
	}

	router := newProfilesTestRouter(
		repo,
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me/completed-jobs",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetCompletedJobs_ServiceError(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{
			completedJobsErr: errors.New("history failure"),
		},
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me/completed-jobs",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetCancelledJobs_Success(t *testing.T) {
	repo := &MockRepository{
		history: []JobHistoryItem{
			{
				ApplicationID: 1,
				Status:        "cancelled",
			},
		},
	}

	router := newProfilesTestRouter(
		repo,
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me/cancelled-jobs",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetFullHistory_Success(t *testing.T) {
	repo := &MockRepository{
		history: []JobHistoryItem{
			{
				ApplicationID: 1,
			},
		},
	}

	router := newProfilesTestRouter(
		repo,
		10,
		"cleaner",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/me/history",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			res.Code,
		)
	}
}

func TestHandler_SendLongServiceErrorSafely(t *testing.T) {
	router := newProfilesTestRouter(
		&MockRepository{
			searchErr: errors.New(
				strings.Repeat("x", 100),
			),
		},
		1,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/search",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d",
			res.Code,
		)
	}
}

func TestHandler_GetNewOnJobiraStatus_Success(
	t *testing.T,
) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			UserID:        15,
			JobsCompleted: 1,
			CreatedAt:     time.Now().AddDate(0, 0, -5),
		},
	}

	router := newProfilesTestRouter(
		repo,
		20,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/15/new-on-jobira",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	var result NewOnJobiraStatus

	if err := json.NewDecoder(
		res.Body,
	).Decode(
		&result,
	); err != nil {
		t.Fatalf(
			"failed decoding response: %v",
			err,
		)
	}

	if !result.IsNew {
		t.Fatal(
			"expected cleaner to be new on Jobira",
		)
	}

	if result.Label != "New on Jobira" {
		t.Fatalf(
			"expected label %q got %q",
			"New on Jobira",
			result.Label,
		)
	}

	if result.JobsCompleted != 1 {
		t.Fatalf(
			"expected 1 completed job got %d",
			result.JobsCompleted,
		)
	}
}

func TestHandler_GetNewOnJobiraStatus_NotNew(
	t *testing.T,
) {
	repo := &MockRepository{
		profile: &CleanerProfile{
			UserID:        15,
			JobsCompleted: 4,
			CreatedAt:     time.Now().AddDate(0, 0, -5),
		},
	}

	router := newProfilesTestRouter(
		repo,
		20,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/15/new-on-jobira",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if res.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}

	var result NewOnJobiraStatus

	if err := json.NewDecoder(
		res.Body,
	).Decode(
		&result,
	); err != nil {
		t.Fatalf(
			"failed decoding response: %v",
			err,
		)
	}

	if result.IsNew {
		t.Fatal(
			"expected cleaner not to be new",
		)
	}

	if result.Label != "" {
		t.Fatalf(
			"expected empty label got %q",
			result.Label,
		)
	}
}

func TestHandler_GetNewOnJobiraStatus_InvalidUserID(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{},
		20,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/invalid/new-on-jobira",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_GetNewOnJobiraStatus_ZeroUserID(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{},
		20,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/0/new-on-jobira",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if res.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}
func TestHandler_GetNewOnJobiraStatus_Unauthorized(
	t *testing.T,
) {
	handler := NewHandler(
		NewService(
			&MockRepository{},
		),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/15/new-on-jobira",
		nil,
	)

	res := httptest.NewRecorder()

	handler.GetNewOnJobiraStatus(
		res,
		req,
	)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_GetNewOnJobiraStatus_NotFound(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{
			getByUserIDErr: ErrProfileNotFound,
		},
		20,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/999/new-on-jobira",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if res.Code != http.StatusNotFound {
		t.Fatalf(
			"expected 404 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_GetNewOnJobiraStatus_ServiceError(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{
			getByUserIDErr: errors.New(
				"database failure",
			),
		},
		20,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/profiles/15/new-on-jobira",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}

func TestHandler_GetNewOnJobiraStatus_RouteMethodNotAllowed(
	t *testing.T,
) {
	router := newProfilesTestRouter(
		&MockRepository{},
		20,
		"client",
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/profiles/15/new-on-jobira",
		nil,
	)

	res := httptest.NewRecorder()

	router.ServeHTTP(
		res,
		req,
	)

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected 405 got %d: %s",
			res.Code,
			res.Body.String(),
		)
	}
}
