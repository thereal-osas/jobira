package billing

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v84/customer"

	"github.com/stripe/stripe-go/v84"

	portalsession "github.com/stripe/stripe-go/v84/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/webhook"
)

type StripeConfig struct {
	SecretKey       string
	WebhookSecret   string
	SuccessURL      string
	CancelURL       string
	PortalReturnURL string
}

type Service struct {
	repo Repository
	cfg  StripeConfig
}

func NewService(repo Repository, cfg StripeConfig) *Service {
	return &Service{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *Service) CreateCheckoutSession(ctx context.Context, userID uint, req CreateCheckoutSessionRequest) (*CheckoutSessionResponse, error) {
	req.PromoCode = strings.TrimSpace(req.PromoCode)

	if userID == 0 || req.PlanID == 0 {
		return nil, ErrInvalidInput
	}

	if s.cfg.SecretKey == "" || s.cfg.SuccessURL == "" || s.cfg.CancelURL == "" {
		return nil, ErrCheckoutSessionError
	}

	plan, err := s.repo.GetPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(plan.StripePriceID) == "" {
		return nil, ErrStripePriceMissing
	}

	email, err := s.repo.GetUserEmail(ctx, userID)
	if err != nil {
		return nil, err
	}

	stripe.Key = s.cfg.SecretKey

	stripeCustomerID, err := s.repo.GetStripeCustomerID(ctx, userID)
	if errors.Is(err, ErrCustomerNotFound) {
		customerParams := &stripe.CustomerParams{
			Email: stripe.String(email),
			Metadata: map[string]string{
				"user_id": strconv.FormatUint(uint64(userID), 10),
			},
		}

		createdCustomer, err := customer.New(customerParams)
		if err != nil {
			return nil, err
		}

		stripeCustomerID = createdCustomer.ID

		if err := s.repo.UpsertBillingCustomer(ctx, userID, email, stripeCustomerID); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	params := &stripe.CheckoutSessionParams{
		Mode:                stripe.String("subscription"),
		Customer:            stripe.String(stripeCustomerID),
		ClientReferenceID:   stripe.String(strconv.FormatUint(uint64(userID), 10)),
		SuccessURL:          stripe.String(s.cfg.SuccessURL),
		CancelURL:           stripe.String(s.cfg.CancelURL),
		AllowPromotionCodes: stripe.Bool(true),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(plan.StripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"user_id": strconv.FormatUint(uint64(userID), 10),
			"plan_id": strconv.FormatUint(uint64(req.PlanID), 10),
		},
	}

	session, err := checkoutsession.New(params)
	if err != nil {
		return nil, err
	}

	record := &CheckoutSessionRecord{
		UserID:               userID,
		PlanID:               req.PlanID,
		StripeSessionID:      session.ID,
		StripeCustomerID:     stripeCustomerID,
		StripeSubscriptionID: "",
		Status:               "created",
	}
	if err := s.repo.SaveCheckoutSession(ctx, record); err != nil {
		return nil, err
	}

	return &CheckoutSessionResponse{
		URL:       session.URL,
		SessionID: session.ID,
	}, nil

}

func (s *Service) CreateBillingPortalSession(ctx context.Context, userID uint) (*BillingPortalResponse, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}

	if s.cfg.SecretKey == "" || s.cfg.PortalReturnURL == "" {
		return nil, ErrPortalSessionError
	}

	stripeCustomerID, err := s.repo.GetStripeCustomerID(ctx, userID)
	if err != nil {
		return nil, err
	}

	stripe.Key = s.cfg.SecretKey

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(stripeCustomerID),
		ReturnURL: stripe.String(s.cfg.PortalReturnURL),
	}

	session, err := portalsession.New(params)
	if err != nil {
		return nil, err
	}

	return &BillingPortalResponse{
		URL: session.URL,
	}, nil
}

func (s *Service) HandleWebhook(ctx context.Context, payload []byte, signatureHeader string) error {
	if s.cfg.WebhookSecret == "" {
		return ErrWebhookInvalid
	}

	event, err := webhook.ConstructEvent(payload, signatureHeader, s.cfg.WebhookSecret)
	if err != nil {
		return ErrWebhookInvalid
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession

		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return err
		}

		userID, err := metadataUint(session.Metadata, "user_id")
		if err != nil {
			return err
		}

		planID, err := metadataUint(session.Metadata, "plan_id")
		if err != nil {
			return err
		}

		stripeCustomerID := ""
		if session.Customer != nil {
			stripeCustomerID = session.Customer.ID
		}

		stripeSubscriptionID := ""
		if session.Subscription != nil {
			stripeSubscriptionID = session.Subscription.ID
		}

		if err := s.repo.MarkCheckoutSessionComplete(
			ctx,
			session.ID,
			stripeCustomerID,
			stripeSubscriptionID,
		); err != nil {
			return err
		}

		return s.repo.ActivateUserSubscription(
			ctx,
			userID,
			planID,
			stripeSubscriptionID,
		)
	}

	if event.Type == "customer.subscription.deleted" {
		var subscription stripe.Subscription

		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			return err
		}

		stripeCustomerID := ""
		if subscription.Customer != nil {
			stripeCustomerID = subscription.Customer.ID
		}

		if stripeCustomerID == "" {
			return nil
		}

		return s.repo.UpdateUserSubscriptionStatusByCustomer(ctx, stripeCustomerID, "cancelled")
	}

	if event.Type == "customer.subscription.updated" {
		var subscription stripe.Subscription

		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			return err
		}

		stripeCustomerID := ""
		if subscription.Customer != nil {
			stripeCustomerID = subscription.Customer.ID
		}

		if stripeCustomerID == "" {
			return nil
		}

		return s.repo.UpdateUserSubscriptionStatusByCustomer(ctx, stripeCustomerID, string(subscription.Status))
	}

	return nil
}

func metadataUint(metadata map[string]string, key string) (uint, error) {
	rawValue := metadata[key]
	if rawValue == "" {
		return 0, ErrInvalidInput
	}

	parsedValue, err := strconv.ParseUint(rawValue, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(parsedValue), nil
}


func (s *Service) ValidatePromoCode(ctx context.Context, userID uint, req ValidatePromoCodeRequest) (*ValidatePromoCodeRespone, error) {
	req.Code = strings.TrimSpace(req.Code)

	if userID == 0 || req.Code == "" {
		return nil, ErrInvalidInput
	}

	promo, err := s.repo.GetPromoCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	if !promo.IsActive {
		return nil, ErrPromoCodeInvalid
	}

	now := time.Now()

	if promo.StartsAt != nil && now.Before(*promo.StartsAt) {
		return nil, ErrPromoCodeInvalid
	}

	if promo.ExpiresAt != nil && now.After(*promo.ExpiresAt) {
		return nil, ErrPromoCodeInvalid
	}

	if promo.MaxUses > 0 && promo.TimesUsed >= promo.MaxUses {
		return nil, ErrPromoCodeInvalid
	}

	if promo.PlanID != nil && req.PlanID != 0 && *promo.PlanID != req.PlanID {
		return nil, ErrPromoCodeInvalid
	}

	return &ValidatePromoCodeRespone{
		Code: promo.Code,
		Valid: true,
		Description: promo.Description,
		DiscountType: promo.DiscountType,
		PercentageOff: promo.PercentageOff,
		FixedAmountPence: promo.FixedAmountPence,
		FreeMonths: promo.FreeMonths,
	}, nil
}