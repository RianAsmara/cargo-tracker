package repository

import (
	"context"

	"github.com/RianAsmara/cargo-tracker/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresDelayWorkerRepo struct {
	db *pgxpool.Pool
}

func NewPostgresDelayWorkerRepository(db *pgxpool.Pool) domain.DelayWorkerRepository {
	return &postgresDelayWorkerRepo{
		db: db,
	}
}

func (r *postgresDelayWorkerRepo) FindStagnantShipments(ctx context.Context) ([]*domain.Shipment, error) {
	// Dynamically compares shipment updated_at against the specific tenant's defined hourly threshold.
	query := `
		SELECT 
			s.id, s.tenant_id, s.resi_number, s.sender_name, s.sender_phone, 
			s.receiver_name, s.receiver_phone, s.origin, s.destination, 
			s.current_status, s.is_delayed, s.created_at, s.updated_at
		FROM shipments s
		JOIN tenants t ON s.tenant_id = t.id
		WHERE s.is_delayed = false
		  AND s.current_status NOT IN ('DELIVERED', 'CANCELED')
		  AND s.updated_at < (CURRENT_TIMESTAMP - (INTERVAL '1 hour' * COALESCE(t.delay_threshold_hours, 24)))
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shipments []*domain.Shipment
	for rows.Next() {
		var s domain.Shipment
		err := rows.Scan(
			&s.ID, &s.TenantID, &s.ResiNumber, &s.SenderName, &s.SenderPhone,
			&s.ReceiverName, &s.ReceiverPhone, &s.Origin, &s.Destination,
			&s.CurrentStatus, &s.IsDelayed, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		shipments = append(shipments, &s)
	}

	return shipments, rows.Err()
}

func (r *postgresDelayWorkerRepo) MarkAsDelayed(ctx context.Context, shipmentID uuid.UUID, tenantID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update shipment flag
	updateQuery := `
		UPDATE shipments 
		SET is_delayed = true, current_status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND tenant_id = $3
	`
	_, err = tx.Exec(ctx, updateQuery, domain.StatusDelayed, shipmentID, tenantID)
	if err != nil {
		return err
	}

	// Append formal tracking event referencing the delay to inform the user timeline
	eventQuery := `
		INSERT INTO tracking_events (id, tenant_id, shipment_id, status, description, location)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.Exec(ctx, eventQuery, 
		uuid.New(), tenantID, shipmentID, 
		domain.StatusDelayed, 
		"System has flagged this shipment as delayed due to lack of recent scanning activity.",
		"System Worker",
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
