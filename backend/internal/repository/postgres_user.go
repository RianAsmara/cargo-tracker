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

type postgresUserRepo struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &postgresUserRepo{
		db: db,
	}
}

func getTenantID(ctx context.Context) (uuid.UUID, error) {
	tenantID, ok := ctx.Value(domain.TenantContextKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, customErr.ErrNoTenantCtx
	}
	return tenantID, nil
}

func (r *postgresUserRepo) Create(ctx context.Context, u *domain.User) error {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return err
	}

	u.TenantID = tenantID
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		u.ID,
		u.TenantID,
		u.Email,
		u.PasswordHash,
		u.Role,
	).Scan(&u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return err
	}
	return nil
}

func (r *postgresUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1 AND tenant_id = $2
	`

	var u domain.User
	err = r.db.QueryRow(ctx, query, email, tenantID).Scan(
		&u.ID,
		&u.TenantID,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customErr.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}
