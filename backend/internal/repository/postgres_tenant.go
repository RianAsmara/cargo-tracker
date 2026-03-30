package repository

import (
	"context"
	"errors"

	"github.com/RianAsmara/cargo-tracker/internal/domain"
	customErr "github.com/RianAsmara/cargo-tracker/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresTenantRepo struct {
	db *pgxpool.Pool
}

func NewPostgresTenantRepository(db *pgxpool.Pool) domain.TenantRepository {
	return &postgresTenantRepo{
		db: db,
	}
}

func (r *postgresTenantRepo) Create(ctx context.Context, t *domain.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, domain, webhook_url, delay_threshold_hours)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, t.ID, t.Name, t.Domain, t.WebhookURL, t.DelayThresholdHours).Scan(&t.CreatedAt, &t.UpdatedAt)
	return err
}

func (r *postgresTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	query := `
		SELECT id, name, domain, webhook_url, delay_threshold_hours, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	var t domain.Tenant
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Domain, &t.WebhookURL, &t.DelayThresholdHours, &t.CreatedAt, &t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customErr.ErrNotFound
		}
		return nil, err
	}

	return &t, nil
}

func (r *postgresTenantRepo) GetByDomain(ctx context.Context, d string) (*domain.Tenant, error) {
	query := `SELECT id, name, domain, webhook_url, delay_threshold_hours, created_at, updated_at FROM tenants WHERE domain = $1`

	var t domain.Tenant
	err := r.db.QueryRow(ctx, query, d).Scan(
		&t.ID,
		&t.Name,
		&t.Domain,
		&t.WebhookURL,
		&t.DelayThresholdHours,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customErr.ErrNotFound
		}
		return nil, err
	}

	return &t, nil
}
