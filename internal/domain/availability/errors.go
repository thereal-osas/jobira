package availability

import "errors"

var (
	ErrInvalidInput                  = errors.New("invalid input")
	ErrAvailabilityNotFound          = errors.New("availability not found")
	ErrAvailabilityExists            = errors.New("availability already exists")
	ErrAvailabilityConflict          = errors.New("availability conflict")
	ErrInvalidStatus                 = errors.New("invalid availability status")
	ErrForbidden                     = errors.New("forbidden")
	ErrRecurringAvailabilityNotFound = errors.New("recurring availability not found")
	ErrAvailabilityOverrideNotFound  = errors.New("availability override not found")
	ErrAvailabilitySettingsNotFound  = errors.New("availability settings not found")
	ErrInvalidTimeRange              = errors.New("invalid availability time range")
	ErrMinimumNotice                 = errors.New("minimum booking notice not met")
	ErrMinimumDuration               = errors.New("minimum booking duration not met")
	ErrMaximumDuration               = errors.New("maximum booking duration exceeded")
	ErrBookingHorizonExceeded        = errors.New("booking horizon exceeded")
	ErrOutsideAvailability           = errors.New("requested time is outside cleaner availability")
)
