package availability

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	r.Route("/availability", func(r chi.Router) {

		// PUBLIC CLEANER AVAILABILITY
		r.Get(
			"/cleaners/{cleanerID}",
			handler.ListByCleaner,
		)

		r.Get(
			"/cleaners/{cleanerID}/calendar",
			handler.GetCalendar,
		)

		r.Get(
			"/cleaners/{cleanerID}/next-available",
			handler.NextAvailable,
		)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			// CLEANER DATE-SPECIFIC AVAILABILITY
			r.Post(
				"/",
				handler.Create,
			)

			r.Post(
				"/bulk",
				handler.CreateBulk,
			)

			r.Get(
				"/me",
				handler.ListMine,
			)

			// CALENDAR SETTINGS
			r.Get(
				"/settings/me",
				handler.GetSettings,
			)

			r.Put(
				"/settings/me",
				handler.UpdateSettings,
			)

			// RECURRING WEEKLY AVAILABILITY
			r.Post(
				"/recurring",
				handler.CreateRecurring,
			)

			r.Post(
				"/recurring/bulk",
				handler.CreateRecurringBulk,
			)

			r.Get(
				"/recurring/me",
				handler.ListRecurring,
			)

			r.Delete(
				"/recurring/{recurringID}",
				handler.DeleteRecurring,
			)

			// ONE-OFF DATE OVERRIDES
			r.Post(
				"/overrides",
				handler.CreateOverride,
			)

			r.Delete(
				"/overrides/{overrideID}",
				handler.DeleteOverride,
			)

			// AVAILABILITY BLOCK MANAGEMENT
			r.Post(
				"/blocks",
				handler.CreateBlock,
			)

			r.Get(
				"/blocks/me",
				handler.ListBlocks,
			)

			r.Delete(
				"/blocks/{blockID}",
				handler.DeleteBlock,
			)

			// INDIVIDUAL AVAILABILITY RECORDS
			r.Get(
				"/{availabilityID}",
				handler.GetByID,
			)

			r.Put(
				"/{availabilityID}",
				handler.Update,
			)

			r.Delete(
				"/{availabilityID}",
				handler.Delete,
			)
		})
	})
}
