package companyaccounts

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route(
		"/companies",
		func(r chi.Router) {
			r.Use(
				authMiddleware,
			)

			// Company creation.
			r.Post(
				"/",
				handler.CreateCompany,
			)

			// Company dashboard.
			r.Get(
				"/{companyID}/dashboard",
				handler.GetDashboard,
			)

			// Company workforce.
			r.Get(
				"/{companyID}/members",
				handler.ListMembers,
			)

			r.Post(
				"/{companyID}/members",
				handler.AddMember,
			)

			r.Delete(
				"/{companyID}/members/{userID}",
				handler.RemoveMember,
			)

			// Activate / deactivate staff.
			r.Patch(
				"/{companyID}/members/{userID}/status",
				handler.UpdateMemberStatus,
			)

			// Promote / change member role.
			r.Patch(
				"/{companyID}/members/{userID}/role",
				handler.UpdateMemberRole,
			)
		},
	)
}