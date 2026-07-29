package app

import (
	"net/http"
	"os"

	"github.com/rodrigueghenda/jobira/internal/domain/admin"
	applicationsdomain "github.com/rodrigueghenda/jobira/internal/domain/applications"
	"github.com/rodrigueghenda/jobira/internal/domain/auth"

	adminmoderationdomain "github.com/rodrigueghenda/jobira/internal/domain/adminmoderation"
	adminsubscriptionsdomain "github.com/rodrigueghenda/jobira/internal/domain/adminsubscriptions"
	analyticsdomain "github.com/rodrigueghenda/jobira/internal/domain/analytics"
	availabilitydomain "github.com/rodrigueghenda/jobira/internal/domain/availability"
	billingdomain "github.com/rodrigueghenda/jobira/internal/domain/billing"
	blockedcleanersdomain "github.com/rodrigueghenda/jobira/internal/domain/blockedcleaners"
	bookingsdomain "github.com/rodrigueghenda/jobira/internal/domain/bookings"
	bookingtimelinedomain "github.com/rodrigueghenda/jobira/internal/domain/bookingtimeline"
	chatdomain "github.com/rodrigueghenda/jobira/internal/domain/chat"
	cleanerdashboarddomain "github.com/rodrigueghenda/jobira/internal/domain/cleanerdashboard"
	cleanerreportsdomain "github.com/rodrigueghenda/jobira/internal/domain/cleanerreports"
	clientdashboarddomain "github.com/rodrigueghenda/jobira/internal/domain/clientdashboard"
	clientnotesdomain "github.com/rodrigueghenda/jobira/internal/domain/clientnotes"
	companyaccountsdomain "github.com/rodrigueghenda/jobira/internal/domain/companyaccounts"
	emaildomain "github.com/rodrigueghenda/jobira/internal/domain/email"
	favoritesdomain "github.com/rodrigueghenda/jobira/internal/domain/favorites"
	"github.com/rodrigueghenda/jobira/internal/domain/health"
	jobalertsdomain "github.com/rodrigueghenda/jobira/internal/domain/jobalerts"
	jobinvitationsdomain "github.com/rodrigueghenda/jobira/internal/domain/jobinvitations"
	"github.com/rodrigueghenda/jobira/internal/domain/jobs"
	messagesdomain "github.com/rodrigueghenda/jobira/internal/domain/messages"
	notificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/notifications"
	preferredcleanersdomain "github.com/rodrigueghenda/jobira/internal/domain/preferredcleaners"
	profilesdomain "github.com/rodrigueghenda/jobira/internal/domain/profiles"
	recentviewsdomain "github.com/rodrigueghenda/jobira/internal/domain/recentviews"
	referralsdomain "github.com/rodrigueghenda/jobira/internal/domain/referrals"
	repeatbookingsdomain "github.com/rodrigueghenda/jobira/internal/domain/repeatbookings"
	reportsdomain "github.com/rodrigueghenda/jobira/internal/domain/reports"
	reputationdomain "github.com/rodrigueghenda/jobira/internal/domain/reputation"
	reviewsdomain "github.com/rodrigueghenda/jobira/internal/domain/reviews"
	savedjobsdomain "github.com/rodrigueghenda/jobira/internal/domain/savedjobs"
	subscriptionaccessdomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptionaccess"
	subscriptionsdomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptions"
	usagedomain "github.com/rodrigueghenda/jobira/internal/domain/subscriptions/usage"
	userdomain "github.com/rodrigueghenda/jobira/internal/domain/users"
	verificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/verifications"
	jwtsecurity "github.com/rodrigueghenda/jobira/internal/security/jwt"
	"github.com/rodrigueghenda/jobira/internal/transport/http/middleware"
)

func (a *App) registerRoutes() {
	a.R.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Jobira API is running"))
	})

	health.RegisterRoutes(a.R, health.NewHandler())

	jwtIssuer := jwtsecurity.NewIssuer(a.Cfg.JWTSecret)
	authMiddleware := middleware.Auth(jwtIssuer)
	blockChecker := blockedcleanersdomain.NewChecker(a.DB)
	adminOnlyMiddleware := middleware.RequireRole("admin")

	emailSender := emaildomain.NewNoopSender()
	emailService := emaildomain.NewService(emailSender)

	authRepo := auth.NewSQLRepository(a.DB)
	authService := auth.NewService(authRepo, jwtIssuer)
	authHandler := auth.NewHandler(authService)
	auth.RegisterRoutes(a.R, authHandler)

	userRepo := userdomain.NewSQLRepository(a.DB)
	userService := userdomain.NewService(userRepo)
	userHandler := userdomain.NewHandler(userService)
	userdomain.RegisterRoutes(a.R, userHandler, authMiddleware)

	adminService := admin.NewService()
	adminHandler := admin.NewHandler(adminService)
	admin.RegisterRoutes(a.R, adminHandler, authMiddleware, adminOnlyMiddleware)

	usageRepo := usagedomain.NewSQLRepository(a.DB)
	usageService := usagedomain.NewService(usageRepo)

	subscriptionaccessRepo := subscriptionaccessdomain.NewSQLRepository(a.DB)
	subscriptionaccessService := subscriptionaccessdomain.NewService(subscriptionaccessRepo)
	subscriptionaccessHandler := subscriptionaccessdomain.NewHandler(subscriptionaccessService)
	subscriptionaccessdomain.RegisterRoutes(a.R, subscriptionaccessHandler, authMiddleware)

	jobRepo := jobs.NewSQLRepository(a.DB)
	jobService := jobs.NewService(jobRepo, usageService, subscriptionaccessService)
	jobHandler := jobs.NewHandler(jobService)
	jobs.RegisterRoutes(a.R, jobHandler, authMiddleware)

	notificationsRepo := notificationsdomain.NewSQLRepository(a.DB)
	notificationsService := notificationsdomain.NewService(notificationsRepo)
	notificationHandler := notificationsdomain.NewHandler(notificationsService)
	notificationsdomain.RegisterRoutes(a.R, notificationHandler, authMiddleware)

	profilesRepo := profilesdomain.NEWSQLRepository(a.DB)
	profilesService := profilesdomain.NewService(profilesRepo)
	profilesHandler := profilesdomain.NewHandler(profilesService)
	profilesdomain.RegisterRoutes(a.R, profilesHandler, authMiddleware, adminOnlyMiddleware)

	applicationsRepo := applicationsdomain.NewSQLRepository(a.DB)
	applicationsService := applicationsdomain.NewService(applicationsRepo, notificationsService, profilesService, usageService, emailService, subscriptionaccessService)
	applicationsHandler := applicationsdomain.NewHandler(applicationsService)
	applicationsdomain.RegisterRoutes(a.R, applicationsHandler, authMiddleware)

	reviewsRepo := reviewsdomain.NewSQLRepository(a.DB)
	reviewsService := reviewsdomain.NewService(reviewsRepo)
	reviewsHandler := reviewsdomain.NewHandler(reviewsService)
	reviewsdomain.RegisterRoutes(a.R, reviewsHandler, authMiddleware)

	messagesRepo := messagesdomain.NewSQLRepository(a.DB)
	messagesService := messagesdomain.NewService(messagesRepo, notificationsService)
	messagesHandler := messagesdomain.NewHandler(messagesService)
	messagesdomain.RegisterRoutes(a.R, messagesHandler, authMiddleware)

	blockedCleanersRepo := blockedcleanersdomain.NewSQLRepository(a.DB)
	blockedCleanersService := blockedcleanersdomain.NewService(blockedCleanersRepo)
	blockedCleanersHandler := blockedcleanersdomain.NewHandler(blockedCleanersService)
	blockedcleanersdomain.RegisterRoutes(a.R, blockedCleanersHandler, authMiddleware)

	favoritesRepo := favoritesdomain.NewSQLRepository(a.DB)
	favoritesService := favoritesdomain.NewService(favoritesRepo, notificationsService, blockChecker)
	favoritesHandler := favoritesdomain.NewHandler(favoritesService)
	favoritesdomain.RegisterRoutes(a.R, favoritesHandler, authMiddleware)

	jobInvitationsRepo := jobinvitationsdomain.NewSQLRepository(a.DB)
	jobInvitationsService := jobinvitationsdomain.NewService(jobInvitationsRepo, notificationsService, blockChecker)
	jobInvitationsHandler := jobinvitationsdomain.NewHandler(jobInvitationsService)

	jobinvitationsdomain.RegisterRoutes(a.R, jobInvitationsHandler, authMiddleware)

	repeatBookingsRepo := repeatbookingsdomain.NewSQLRepository(a.DB)
	repeatBookingsService := repeatbookingsdomain.NewService(repeatBookingsRepo, notificationsService, blockChecker)
	repeatBookingsHandler := repeatbookingsdomain.NewHandler(repeatBookingsService)

	repeatbookingsdomain.RegisterRoutes(a.R, repeatBookingsHandler, authMiddleware)

	clientNotesRepo := clientnotesdomain.NewSQLRepository(a.DB)
	clientNotesService := clientnotesdomain.NewService(clientNotesRepo)
	clientNotesHandler := clientnotesdomain.NewHandler(clientNotesService)

	clientnotesdomain.RegisterRoutes(a.R, clientNotesHandler, authMiddleware)

	recentViewsRepo := recentviewsdomain.NewSQLRepository(a.DB)
	recentViewsService := recentviewsdomain.NewService(recentViewsRepo)
	recentViewsHandler := recentviewsdomain.NewHandler(recentViewsService)

	recentviewsdomain.RegisterRoutes(a.R, recentViewsHandler, authMiddleware)

	cleanerReportsRepo := cleanerreportsdomain.NewSQLRepository(a.DB)
	cleanerReportsService := cleanerreportsdomain.NewService(cleanerReportsRepo)
	cleanerReportsHandler := cleanerreportsdomain.NewHandler(cleanerReportsService)
	cleanerreportsdomain.RegisterRoutes(a.R, cleanerReportsHandler, authMiddleware)

	preferredcleanersRepo := preferredcleanersdomain.NewSQLRepository(a.DB)
	preferredCleanersService := preferredcleanersdomain.NewService(preferredcleanersRepo, blockChecker)
	preferredCleanersHandler := preferredcleanersdomain.NewHandler(preferredCleanersService)

	preferredcleanersdomain.RegisterRoutes(a.R, preferredCleanersHandler, authMiddleware)

	adminModerationRepo := adminmoderationdomain.NewSQLRepository(a.DB)
	adminModerationService := adminmoderationdomain.NewService(adminModerationRepo)
	adminModerationHandler := adminmoderationdomain.NewHandler(adminModerationService)

	adminmoderationdomain.RegisterRoutes(a.R, adminModerationHandler, authMiddleware, adminOnlyMiddleware)

	adminSubscriptionsRepo := adminsubscriptionsdomain.NewSQLRepository(a.DB)
	adminSubscriptionsService := adminsubscriptionsdomain.NewService(adminSubscriptionsRepo)
	adminSubscriptionsHandler := adminsubscriptionsdomain.NewHandler(adminSubscriptionsService)

	adminsubscriptionsdomain.RegisterRoutes(a.R, adminSubscriptionsHandler, authMiddleware, adminOnlyMiddleware)

	subscriptionsRepo := subscriptionsdomain.NewSQLRepository(a.DB)
	subscriptionsService := subscriptionsdomain.NewService(subscriptionsRepo)
	subscriptionHandler := subscriptionsdomain.NewHandler(subscriptionsService)
	subscriptionsdomain.RegisterRoutes(a.R, subscriptionHandler, authMiddleware)

	companyAccountRepo := companyaccountsdomain.NewSQLRepository(a.DB)
	companyAccountService := companyaccountsdomain.NewService(companyAccountRepo)
	companyAccountHandler := companyaccountsdomain.NewHandler(companyAccountService)
	companyaccountsdomain.RegisterRoutes(a.R, companyAccountHandler, authMiddleware)

	bookingTimelinerepo := bookingtimelinedomain.NewSQLRepository(a.DB)
	bookingTimelineService := bookingtimelinedomain.NewService(bookingTimelinerepo)
	bookingTimelineHandler := bookingtimelinedomain.NewHandler(bookingTimelineService)

	bookingtimelinedomain.RegisterRoutes(
		a.R,
		bookingTimelineHandler,
		authMiddleware,
	)

	bookingsRepo := bookingsdomain.NewSQLRepository(a.DB)
	bookingsService := bookingsdomain.NewService(
		bookingsRepo,
		emailService,
		reviewsService,
		favoritesService,
		preferredCleanersService,
		bookingTimelineService,
		notificationsService,
	)
	bookingsHandler := bookingsdomain.NewHandler(bookingsService)

	bookingsdomain.RegisterRoutes(
		a.R,
		bookingsHandler,
		authMiddleware,
	)

	chatRepo := chatdomain.NewSQLRepository(a.DB)

	chatService := chatdomain.NewService(
		chatRepo,
		bookingsService,
	)

	chatHandler := chatdomain.NewHandler(chatService)

	chatdomain.RegisterRoutes(
		a.R,
		chatHandler,
		authMiddleware,
	)

	availabilityRepo := availabilitydomain.NewSQLRepository(a.DB)
	availabilityService := availabilitydomain.NewService(availabilityRepo)
	availabilityHandler := availabilitydomain.NewHandler(availabilityService)
	availabilitydomain.RegisterRoutes(a.R, availabilityHandler, authMiddleware)

	savedJobsRepo := savedjobsdomain.NewSQLRepository(a.DB)
	savedJobsService := savedjobsdomain.NewService(savedJobsRepo)
	savedJobHandler := savedjobsdomain.NewHandler(savedJobsService)
	savedjobsdomain.RegisterRoutes(a.R, savedJobHandler, authMiddleware)

	jobAlertsRepo := jobalertsdomain.NewSQLRepository(a.DB)
	jobAlertsService := jobalertsdomain.NewService(jobAlertsRepo)
	jobAlertsHandler := jobalertsdomain.NewHandler(jobAlertsService)
	jobalertsdomain.RegisterRoutes(a.R, jobAlertsHandler, authMiddleware)

	verificationsRepo := verificationsdomain.NewSQLRepository(a.DB)
	verificationsService := verificationsdomain.NewService(verificationsRepo)
	verificationsHandler := verificationsdomain.NewHandler(verificationsService)
	verificationsdomain.RegisterRoutes(a.R, verificationsHandler, authMiddleware, adminOnlyMiddleware)

	reportsRepo := reportsdomain.NewSQLRepository(a.DB)
	reportService := reportsdomain.NewService(reportsRepo)
	reportsHandler := reportsdomain.NewHandler(reportService)

	reportsdomain.RegisterRoutes(a.R, reportsHandler, authMiddleware, adminOnlyMiddleware)

	analyticsRepo := analyticsdomain.NewSQLRepository(a.DB)
	analyticsService := analyticsdomain.NewService(analyticsRepo)
	analyticsHandler := analyticsdomain.NewHandler(analyticsService)
	analyticsdomain.RegisterRoutes(a.R, analyticsHandler, authMiddleware, adminOnlyMiddleware)

	billingRepo := billingdomain.NEWSQLRepository(a.DB)
	billingService := billingdomain.NewService(
		billingRepo, billingdomain.StripeConfig{
			SecretKey:       os.Getenv("STRIPE_SECRET_KEY"),
			WebhookSecret:   os.Getenv("STRIPE_WEBHOOK_SECRET"),
			SuccessURL:      os.Getenv("STRIPE_SUCCESS_URL"),
			CancelURL:       os.Getenv("STRIPE_CANCEL_URL"),
			PortalReturnURL: os.Getenv("STRIPE_PORTAL_RETURN_URL"),
		},
	)
	billingHandler := billingdomain.NewHandler(billingService)
	billingdomain.RegisterRoutes(a.R, billingHandler, authMiddleware)

	referralsRepo := referralsdomain.NewSQLRepository(a.DB)
	referralsService := referralsdomain.NewService(referralsRepo)
	referralsHandler := referralsdomain.NewHandler(referralsService)
	referralsdomain.RegisterRoutes(a.R, referralsHandler, authMiddleware)

	reputationRepo := reputationdomain.NewSQLRepository(a.DB)
	reputationService := reputationdomain.NewService(reputationRepo)
	reputationHandler := reputationdomain.NewHandler(reputationService)

	reputationdomain.RegisterRoutes(a.R, reputationHandler)

	cleanerDashboardRepo := cleanerdashboarddomain.NewSQLRepository(a.DB)

	cleanerDashboardService := cleanerdashboarddomain.NewService(
		cleanerDashboardRepo,
		reputationService,
		subscriptionaccessService,
	)

	cleanerDashboardHandler := cleanerdashboarddomain.NewHandler(cleanerDashboardService)

	cleanerdashboarddomain.RegisterRoutes(
		a.R,
		cleanerDashboardHandler,
		authMiddleware,
	)

	clientDashboardRepo := clientdashboarddomain.NewSQLRepository(a.DB)

	clientDashboardService := clientdashboarddomain.NewService(
		clientDashboardRepo,
		subscriptionaccessService,
	)

	clientDashboardHandler := clientdashboarddomain.NewHandler(
		clientDashboardService,
	)

	clientdashboarddomain.RegisterRoutes(
		a.R,
		clientDashboardHandler,
		authMiddleware,
	)

}

func (a *App) Router() http.Handler {
	return a.R
}
