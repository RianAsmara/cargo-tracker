package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID                  uuid.UUID
	Name                string
	Domain              string
	WebhookURL          string
	DelayThresholdHours int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type contextKey string

const TenantContextKey contextKey = "tenant_id"

type TenantRepository interface {
	Create(ctx context.Context, t *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*Tenant, error)
}
