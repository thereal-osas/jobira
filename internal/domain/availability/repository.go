package availability

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, availability *CleanerAvailability) error
	GetByID(ctx context.Context, id uint) (*CleanerAvailability, error)
	ListByCleanerID(ctx context.Context, cleanerID uint) ([]CleanerAvailability, error)
	ListByCleanerIDRange(ctx context.Context, cleanerID uint, fromDate string, toDate string) ([]CleanerAvailability, error)
	Update(ctx context.Context, availability *CleanerAvailability) error
	Delete(ctx context.Context, id uint) error
	HasConflict(ctx context.Context, cleanrID uint, availableDate string, startTime string, endTime string, excludeID uint) (bool, error)
	CreateBlock(ctx context.Context, block *AvailabilityBlock) error
	ListBlocksByCleanerID(ctx context.Context, cleanerID uint) ([]AvailabilityBlock, error)
	ListBlocksByCleanerIDRange(ctx context.Context, cleanerID uint, startAt time.Time, endAt time.Time) ([]AvailabilityBlock, error)
	DeleteBlock(ctx context.Context, blockID uint, cleanerID uint) error
	CreateRecurring(ctx context.Context, recurring *RecurringAvailability) error
	ListRecurringByCleanerID(ctx context.Context, cleanerID uint) ([]RecurringAvailability, error)
	DeleteRecurring(ctx context.Context, recurringID uint, cleanerID uint) error
	UpsertSettings(ctx context.Context, settings *AvailabilitySettings) error
	GetSettings(ctx context.Context, cleanerID uint) (*AvailabilitySettings, error)
	CreateOverride(ctx context.Context, override *AvailabilityOverride) error
	ListOverridesByCleanerIDRange(ctx context.Context, cleanerID uint, fromDate string, toDate string) ([]AvailabilityOverride, error)
	DeleteOverride(ctx context.Context, overrideID uint, cleanerID uint) error
}
