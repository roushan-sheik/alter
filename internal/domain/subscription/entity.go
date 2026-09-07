package subscription

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlanCode is the known subscription plan identifier.
type PlanCode string

// BillingInterval defines how often the plan is billed.
type BillingInterval string

// Store is the payment store (Apple or Google).
type Store string

// SubscriptionStatus is the current state of a subscription.
type SubscriptionStatus string

const (
	PlanCodeBiannual PlanCode = "BIANNUAL"
	PlanCodeAnnual   PlanCode = "ANNUAL"
	PlanCodeFamily   PlanCode = "FAMILY"

	IntervalMonth BillingInterval = "MONTH"
	IntervalYear  BillingInterval = "YEAR"

	StoreApple  Store = "APPLE"
	StoreGoogle Store = "GOOGLE"

	StatusActive   SubscriptionStatus = "ACTIVE"
	StatusCanceled SubscriptionStatus = "CANCELED"
	StatusPastDue  SubscriptionStatus = "PAST_DUE"
	StatusExpired  SubscriptionStatus = "EXPIRED"
)

// SubscriptionPlan represents the SUBSCRIPTION_PLANS table.
type SubscriptionPlan struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Code            PlanCode        `gorm:"type:varchar(50);not null;uniqueIndex"`
	Name            string          `gorm:"type:varchar(255);not null"`
	PriceAmount     float64         `gorm:"type:numeric(10,2);not null"`
	Currency        string          `gorm:"type:varchar(10);not null;default:'USD'"`
	BillingInterval BillingInterval `gorm:"type:varchar(20);not null"`
	IsActive        bool            `gorm:"not null;default:true"`
	CreatedAt       time.Time
}

// BeforeCreate assigns a UUID before inserting a new SubscriptionPlan row.
func (p *SubscriptionPlan) BeforeCreate(_ *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}

// Subscription represents the SUBSCRIPTIONS table.
type Subscription struct {
	ID                    uuid.UUID          `gorm:"type:uuid;primaryKey"`
	UserID                uuid.UUID          `gorm:"type:uuid;not null;index"`
	PlanID                uuid.UUID          `gorm:"type:uuid;not null"`
	Store                 Store              `gorm:"type:varchar(20);not null"`
	Status                SubscriptionStatus `gorm:"type:varchar(20);not null;index"`
	ExternalTransactionID string             `gorm:"type:varchar(500);uniqueIndex"`
	StartDate             time.Time          `gorm:"not null"`
	ExpiresAt             time.Time          `gorm:"not null"`
	CreatedAt             time.Time
	UpdatedAt             time.Time

	// Preload-able relation
	Plan SubscriptionPlan `gorm:"foreignKey:PlanID"`
}

// BeforeCreate assigns a UUID before inserting a new Subscription row.
func (s *Subscription) BeforeCreate(_ *gorm.DB) error {
	s.ID = uuid.New()
	return nil
}
