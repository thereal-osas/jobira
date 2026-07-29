package auth

import (
	"context"
	"strings"

	jwtsecurity "github.com/rodrigueghenda/jobira/internal/security/jwt"
	"github.com/rodrigueghenda/jobira/internal/security/password"
)


type Service struct {
	repo Repository
	issuer *jwtsecurity.Issuer
}

func NewService(repo Repository, issuer *jwtsecurity.Issuer) *Service {
	return &Service{
		repo: repo,
		issuer: issuer,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.FullName == "" || req.Email == "" || req.Password == "" {
		return nil, ErrInvalidInput
	}

	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &User{
		FullName: 		req.FullName,
		Email:			req.Email,
		PasswordHash: 	hashedPassword,
		Role: 			"user",
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.issuer.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: UserPayload{
			ID:			user.ID,
			FullName: 	user.FullName,
			Email:		user.Email,
			Role: 		user.Role,
		},
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if err := password.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.issuer.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: UserPayload{
			ID: 	    user.ID,
			FullName:	user.FullName,
			Email:		user.Email,
			Role:		user.Role,
		},
	}, nil
}
