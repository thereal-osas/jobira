package profilemedia

import "time"

type MediaType string

const (
	MediaTypeProfilePhoto MediaType = "profile_photo"
	MediaTypePortfolio    MediaType = "portfolio"
	MediaTypePricing      MediaType = "pricing"
	MediaTypeCompanyLogo  MediaType = "company_logo"
)

type ProfileMedia struct {
	ID uint `json:"id"`

	OwnerUserID *uint `json:"Owner_user_id,omitempty"`

	CompanyID *uint `json:"company_id,omitempty"`

	MediaType MediaType `json:"media_type"`

	URL string `json:"url"`

	Caption string `json:"caption"`

	SortOrder int `json:"sort_order"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`
}

type MediaLimits struct {
	ProfilePhoto int `json:"profile_photo"`

	Portfolio int `json:"portfolio"`

	Pricing int `json:"pricing"`

	CompanyLogo int `json:"company_logo"`
}

func DefaultMediaLimits() MediaLimits {
	return MediaLimits{
		ProfilePhoto: 1,
		Portfolio:    6,
		Pricing:      1,
		CompanyLogo:  1,
	}
}
