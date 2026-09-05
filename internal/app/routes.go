package app

import (
	"net/http"

	"github.com/rodrigueghenda/jobira/internal/domain/admin"
	adminmoderationdomain "github.com/rodrigueghenda/jobira/internal/domain/adminmoderation"
	adminsubscriptionsdomain "github.com/rodrigueghenda/jobira/internal/domain/adminsubscriptions"
	analyticsdomain "github.com/rodrigueghenda/jobira/internal/domain/analytics"
	applicationsdomain "github.com/rodrigueghenda/jobira/internal/domain/applications"
	applicationtimelinedomain "github.com/rodrigueghenda/jobira/internal/domain/applicationtimeline"
	"github.com/rodrigueghenda/jobira/internal/domain/auth"
	availabilitydomain "github.com/rodrigueghenda/jobira/internal/domain/availability"
	availablenowdomain "github.com/rodrigueghenda/jobira/internal/domain/availablenow"
	billingdomain "github.com/rodrigueghenda/jobira/internal/domain/billing"
	blockedcleanersdomain "github.com/rodrigueghenda/jobira/internal/domain/blockedcleaners"
	bookingsdomain "github.com/rodrigueghenda/jobira/internal/domain/bookings"
	bookingtimelinedomain "github.com/rodrigueghenda/jobira/internal/domain/bookingtimeline"
	chatdomain "github.com/rodrigueghenda/jobira/internal/domain/chat"
	cleanerdashboarddomain "github.com/rodrigueghenda/jobira/internal/domain/cleanerdashboard"
	cleanerprogressiondomain "github.com/rodrigueghenda/jobira/internal/domain/cleanerprogression"
	cleanerreportsdomain "github.com/rodrigueghenda/jobira/internal/domain/cleanerreports"
	cleaningteamdomain "github.com/rodrigueghenda/jobira/internal/domain/cleaningteam"
	clientdashboarddomain "github.com/rodrigueghenda/jobira/internal/domain/clientdashboard"
	clientnotesdomain "github.com/rodrigueghenda/jobira/internal/domain/clientnotes"
	companyaccountsdomain "github.com/rodrigueghenda/jobira/internal/domain/companyaccounts"
	emaildomain "github.com/rodrigueghenda/jobira/internal/domain/email"
	favoritesdomain "github.com/rodrigueghenda/jobira/internal/domain/favorites"
	"github.com/rodrigueghenda/jobira/internal/domain/health"
	jobalertsdomain "github.com/rodrigueghenda/jobira/internal/domain/jobalerts"
	jobinvitationsdomain "github.com/rodrigueghenda/jobira/internal/domain/jobinvitations"
	jobpulsedomain "github.com/rodrigueghenda/jobira/internal/domain/jobpulse"
	workhistorydomain "github.com/rodrigueghenda/jobira/internal/domain/workhistory"
	"github.com/rodrigueghenda/jobira/internal/domain/jobs"
	matchscoredomain "github.com/rodrigueghenda/jobira/internal/domain/matchscore"
	messagesdomain "github.com/rodrigueghenda/jobira/internal/domain/messages"
	notificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/notifications"
	preferredcleanersdomain "github.com/rodrigueghenda/jobira/internal/domain/preferredcleaners"
	profilemediadomain "github.com/rodrigueghenda/jobira/internal/domain/profilemedia"
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
	topofferdomain "github.com/rodrigueghenda/jobira/internal/domain/topoffer"
	userdomain "github.com/rodrigueghenda/jobira/internal/domain/users"
	verificationsdomain "github.com/rodrigueghenda/jobira/internal/domain/verifications"
	workproofdomain "github.com/rodrigueghenda/jobira/internal/domain/workproof"
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

	var emailSender emaildomain.Sender

	if a.Cfg.SMTPHost != "" &&
		a.Cfg.SMTPPort != "" &&
		a.Cfg.SMTPFrom != "" {

		emailSender = emaildomain.NewSMTPSender(
			emaildomain.SMTPConfig{
				Host:     a.Cfg.SMTPHost,
				Port:     a.Cfg.SMTPPort,
				Username: a.Cfg.SMTPUsername,
				Password: a.Cfg.SMTPPassword,
				From:     a.Cfg.SMTPFrom,
			},
		)
	} else {
		emailSender = emaildomain.NewNoopSender()
	}

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

	jobPulseRepo := jobpulsedomain.NewSQLRepository(
		a.DB,
	)

	jobPulseService := jobpulsedomain.NewService(
		jobPulseRepo,
	)

	jobPulseHandler := jobpulsedomain.NewHandler(
		jobPulseService,
	)

	jobpulsedomain.RegisterRoutes(
		a.R,
		jobPulseHandler,
	)

	notificationsRepo := notificationsdomain.NewSQLRepository(a.DB)
	notificationsService := notificationsdomain.NewService(notificationsRepo)
	notificationHandler := notificationsdomain.NewHandler(notificationsService)
	notificationsdomain.RegisterRoutes(a.R, notificationHandler, authMiddleware)

	profileMediaRepo := profilemediadomain.NewSQLRepository(a.DB)

	profilesRepo := profilesdomain.NewSQLRepository(a.DB)
	profilesService := profilesdomain.NewService(profilesRepo)

	profilesService.SetPortfolioCounter(
		profileMediaRepo,
	)

	profilesHandler := profilesdomain.NewHandler(profilesService)

	profilesdomain.RegisterRoutes(
		a.R,
		profilesHandler,
		authMiddleware,
		adminOnlyMiddleware,
	)

	profileMediaService := profilemediadomain.NewService(
		profileMediaRepo,
	)

	profileMediaHandler := profilemediadomain.NewHandler(
		profileMediaService,
	)

	profilemediadomain.RegisterRoutes(
		a.R,
		profileMediaHandler,
		authMiddleware,
	)

	applicationTimelineRepo := applicationtimelinedomain.NewSQLRepository(a.DB)
	applicationTimelineService := applicationtimelinedomain.NewService(applicationTimelineRepo)
	applicationTimelineHandler := applicationtimelinedomain.NewHandler(applicationTimelineService)

	applicationtimelinedomain.RegisterRoutes(
		a.R,
		applicationTimelineHandler,
		authMiddleware,
	)

	applicationsRepo := applicationsdomain.NewSQLRepository(a.DB)
	applicationsService := applicationsdomain.NewService(applicationsRepo, notificationsService, profilesService, usageService, emailService, subscriptionaccessService)
	applicationsService.SetTimelineRecorder(
		applicationTimelineService,
	)
	applicationsHandler := applicationsdomain.NewHandler(applicationsService)
	applicationsdomain.RegisterRoutes(a.R, applicationsHandler, authMiddleware)

	reviewsRepo := reviewsdomain.NewSQLRepository(a.DB)
	reviewsService := reviewsdomain.NewService(reviewsRepo)
	reviewsHandler := reviewsdomain.NewHandler(reviewsService)
	reviewsdomain.RegisterRoutes(a.R, reviewsHandler, authMiddleware)

	workProofRepo := workproofdomain.NewSQLRepository(a.DB)

	workProofService := workproofdomain.NewService(
		workProofRepo,
	)

	workProofHandler := workproofdomain.NewHandler(
		workProofService,
	)

	workproofdomain.RegisterRoutes(
		a.R,
		workProofHandler,
		authMiddleware,
	)

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

	availabilityRepo := availabilitydomain.NewSQLRepository(a.DB)
	availabilityService := availabilitydomain.NewService(availabilityRepo)
	availabilityHandler := availabilitydomain.NewHandler(availabilityService)
	availabilitydomain.RegisterRoutes(a.R, availabilityHandler, authMiddleware)

	availableNowRepo := availablenowdomain.NewSQLRepository(a.DB)

	availableNowService := availablenowdomain.NewService(
		availableNowRepo,
	)

	availableNowHandler := availablenowdomain.NewHandler(
		availableNowService,
	)

	availablenowdomain.RegisterRoutes(
		a.R,
		availableNowHandler,
		authMiddleware,
	)

	repeatBookingsRepo := repeatbookingsdomain.NewSQLRepository(a.DB)
	repeatBookingsService := repeatbookingsdomain.NewService(repeatBookingsRepo, notificationsService, blockChecker, availabilityService)
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

	bookingsService.SetAvailabilityChecker(availabilityService)

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

	billingRepo := billingdomain.NewSQLRepository(a.DB)

	billingService := billingdomain.NewService(
		billingRepo,
		billingdomain.StripeConfig{
			SecretKey:       a.Cfg.StripeSecretKey,
			WebhookSecret:   a.Cfg.StripeWebhookSecret,
			SuccessURL:      a.Cfg.StripeSuccessURL,
			CancelURL:       a.Cfg.StripeCancelURL,
			PortalReturnURL: a.Cfg.StripePortalReturnURL,
		},
	)
	billingHandler := billingdomain.NewHandler(billingService)
	billingdomain.RegisterRoutes(a.R, billingHandler, authMiddleware)

	referralsRepo := referralsdomain.NewSQLRepository(a.DB)
	referralsService := referralsdomain.NewService(referralsRepo)
	referralsHandler := referralsdomain.NewHandler(referralsService)
	referralsdomain.RegisterRoutes(a.R, referralsHandler, authMiddleware)

	matchScoreRepo := matchscoredomain.NewSQLRepository(a.DB)

	matchScoreService := matchscoredomain.NewService(
		matchScoreRepo,
	)

	matchScoreHandler := matchscoredomain.NewHandler(
		matchScoreService,
	)

	matchscoredomain.RegisterRoutes(
		a.R,
		matchScoreHandler,
		authMiddleware,
	)

	topOfferRepo := topofferdomain.NewSQLRepository(a.DB)

	topOfferService := topofferdomain.NewService(
		topOfferRepo,
	)

	topOfferHandler := topofferdomain.NewHandler(
		topOfferService,
	)

	topofferdomain.RegisterRoutes(
		a.R,
		topOfferHandler,
		authMiddleware,
	)

	cleaningTeamRepo := cleaningteamdomain.NewSQLRepository(a.DB)

	cleaningTeamService := cleaningteamdomain.NewService(
		cleaningTeamRepo,
	)

	cleaningTeamHandler := cleaningteamdomain.NewHandler(
		cleaningTeamService,
	)

	cleaningteamdomain.RegisterRoutes(
		a.R,
		cleaningTeamHandler,
		authMiddleware,
	)

	reputationRepo := reputationdomain.NewSQLRepository(a.DB)
	reputationService := reputationdomain.NewService(reputationRepo)
	reputationHandler := reputationdomain.NewHandler(reputationService)

	reputationdomain.RegisterRoutes(a.R, reputationHandler)

	cleanerProgressionRepo := cleanerprogressiondomain.NewSQLRepository(a.DB)

	cleanerProgressionService := cleanerprogressiondomain.NewService(cleanerProgressionRepo)

	cleanerProgressionHandler := cleanerprogressiondomain.NewHandler(cleanerProgressionService)

	cleanerprogressiondomain.RegisterRoutes(a.R, cleanerProgressionHandler, authMiddleware)

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

	workHistoryRepo := workhistorydomain.NewSQLRepository(a.DB)

workHistoryService := workhistorydomain.NewService(
	workHistoryRepo,
)

workHistoryHandler := workhistorydomain.NewHandler(
	workHistoryService,
)

workhistorydomain.RegisterRoutes(
	a.R,
	workHistoryHandler,
	authMiddleware,
)

}

func (a *App) Router() http.Handler {
	return a.R
}
