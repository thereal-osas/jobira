package cleaningteam

import (
	"context"
	"errors"
	"sort"
	"strings"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetMyCleaningTeam(ctx context.Context, clientID uint) (*CleaningTeam, error) {
	if clientID == 0 {
		return nil, ErrInvalidInput
	}

	data, err := s.repo.ListTeamMembers(
		ctx, clientID,
	)
	if err != nil {
		return nil, err
	}

	members := make(
		[]CleaningTeamMember,
		0,
		len(data),
	)

	preferredCount := 0
	favoriteCount := 0
	housekeeperCount := 0

	for _, item := range data {
		member := buildCleaningTeamMember(
			item,
		)

		if member.IsPreferred {
			preferredCount++
		}

		if member.IsFavorite {
			favoriteCount++
		}

		if member.ServiceRole == "Housekeeper" {
			housekeeperCount++
		}

		members = append(members, member)
	}

	sortCleaningTeam(
		members,
	)

	return &CleaningTeam{
		ClientID: clientID,

		TotalMembers: len(members),

		PreferredCount: preferredCount,
		FavoriteCount:  favoriteCount,

		HouseKeeperCount: housekeeperCount,

		Members: members,
	}, nil
}

func buildCleaningTeamMember(data TeamMemberData) CleaningTeamMember {
	services := parseServices(
		data.ServicesOffered,
	)

	serviceRole := determineServiceRole(
		data.LastJobType,
		services,
	)

	return CleaningTeamMember{
		CleanerID: data.CleanerID,

		CleanerName: data.CleanerName,

		ServiceRole: serviceRole,

		Services: services,

		Location: data.Location,

		AvailabilityStatus: data.AvailabilityStatus,

		IsVerified: data.IsVerified,

		IsFavorite: data.IsFavorite,

		IsPreferred: data.IsPreferred,

		RelationshipType: determineRelationshipType(
			data,
		),

		CompletedJobsTogether: data.CompletedJobsTogether,

		LastBookingID: data.LastBookingID,

		LastBookedAt: data.LastBookedAt,

		CanRebook: canRebook(data),

		AverageRating: data.AverageRating,

		TotalReviews: data.TotalReviews,

		ReliabilityScore: data.ReliabilityScore,

		Badge: data.Badge,
	}
}

func determineRelationshipType(data TeamMemberData) string {
	if data.IsPreferred {
		return "preferred"
	}

	if data.CompletedJobsTogether > 0 {
		return "previously_booked"
	}

	if data.IsFavorite {
		return "shortlisted"
	}

	return "saved"
}

func canRebook(data TeamMemberData) bool {
	return data.CompletedJobsTogether > 0 && data.LastBookingID != nil
}

func determineServiceRole(
	LastJobType string,
	services []string,
) string {
	LastJobType = normalizeService(
		LastJobType,
	)

	if isHousekeepingService(
		LastJobType,
	) {
		return "Housekeeper"
	}

	if isAirbnbService(
		LastJobType,
	) {
		return "Airbnb Cleaner"
	}

	if isEndOfTenancyService(
		LastJobType,
	) {
		return "End of Tenancy Cleaner"
	}

	if isDeepCleanService(
		LastJobType,
	) {
		return "Deep Clean Specialist"
	}

	if isCommercialService(
		LastJobType,
	) {
		return "Commercial Cleaner"
	}

	for _, service := range services {
		if isHousekeepingService(
			service,
		) {
			return "Housekeeper"
		}
	}

	for _, service := range services {
		if isAirbnbService(
			service,
		) {
			return "Airbnb Cleaner"
		}
	}

	for _, service := range services {
		if isEndOfTenancyService(
			service,
		) {
			return "Deep Clean Specialist"
		}
	}

	for _, service := range services {
		if isCommercialService(
			service,
		) {
			return "Commercial Cleaner"
		}
	}

	return "Cleaner"
}

func parseServices(value string) []string {
	value = strings.TrimSpace(
		value,
	)

	if value == "" {
		return []string{}
	}

	replacer := strings.NewReplacer(
		";", ",",
		"|", ",",
	)

	value = replacer.Replace(
		value,
	)

	parts := strings.Split(
		value,
		",",
	)

	services := make(
		[]string,
		0,
		len(parts),
	)

	seen := make(
		map[string]bool,
	)

	for _, part := range parts {
		service := normalizeService(
			part,
		)

		if service == "" {
			continue
		}

		if seen[service] {
			continue
		}

		seen[service] = true

		services = append(
			services,
			service,
		)
	}

	return services
}

func normalizeService(value string) string {
	value = strings.ToLower(
		strings.TrimSpace(value),
	)

	replacer := strings.NewReplacer(
		"_", " ",
		"_", " ",
		"/", " ",
	)

	value = replacer.Replace(
		value,
	)

	return strings.Join(strings.Fields(value),
		" ",
	)
}

func isHousekeepingService(
	value string,
) bool {
	switch normalizeService(value) {
	case "housekeeping",
		"housekeeper",
		"house keeping",
		"laundry",
		"ironing":
		return true

	default:
		return false
	}
}

func isAirbnbService(
	value string,
) bool {
	switch normalizeService(value) {
	case "airbnb",
		"short let",
		"shortlet",
		"holiday let":
		return true

	default:
		return false
	}
}

func isEndOfTenancyService(
	value string,
) bool {
	switch normalizeService(value) {
	case "end of tenancy",
		"end tenancy":
		return true

	default:
		return false
	}
}

func isDeepCleanService(
	value string,
) bool {
	switch normalizeService(value) {
	case "deep clean",
		"deep cleaning":
		return true

	default:
		return false
	}
}

func isCommercialService(
	value string,
) bool {
	switch normalizeService(value) {
	case "commercial",
		"commercial cleaning",
		"office",
		"office cleaning":
		return true

	default:
		return false
	}
}

func sortCleaningTeam(
	members []CleaningTeamMember,
) {
	sort.SliceStable(
		members,
		func(i, j int) bool {
			if members[i].IsPreferred !=
				members[j].IsPreferred {
				return members[i].IsPreferred
			}

			if members[i].CompletedJobsTogether !=
				members[j].CompletedJobsTogether {
				return members[i].CompletedJobsTogether >
					members[j].CompletedJobsTogether
			}

			if members[i].ReliabilityScore !=
				members[j].ReliabilityScore {
				return members[i].ReliabilityScore >
					members[j].ReliabilityScore
			}

			return members[i].CleanerName <
				members[j].CleanerName
		},
	)
}
