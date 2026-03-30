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

type postgresShipmentRepo struct {
	db *pgxpool.Pool
}

func NewPostgresShipmentRepository(db *pgxpool.Pool) domain.ShipmentRepository {
	return &postgresShipmentRepo{
		db: db,
	}
}

func (r *postgresShipmentRepo) Create(ctx context.Context, s *domain.Shipment) error {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return err
	}

	s.TenantID = tenantID
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.CurrentStatus == "" {
		s.CurrentStatus = string(domain.StatusPending)
	}

	query := `
		INSERT INTO shipments (id, tenant_id, resi_number, sender_name, sender_phone, receiver_name, receiver_phone, origin, destination, current_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		s.ID, s.TenantID, s.ResiNumber, s.SenderName, s.SenderPhone,
		s.ReceiverName, s.ReceiverPhone, s.Origin, s.Destination, s.CurrentStatus,
	).Scan(&s.CreatedAt, &s.UpdatedAt)

	return err
}

func (r *postgresShipmentRepo) GetByResi(ctx context.Context, resiNumber string) (*domain.Shipment, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, resi_number, sender_name, sender_phone, receiver_name, receiver_phone, origin, destination, current_status, created_at, updated_at
		FROM shipments
		WHERE resi_number = $1 AND tenant_id = $2
	`

	var s domain.Shipment
	err = r.db.QueryRow(ctx, query, resiNumber, tenantID).Scan(
		&s.ID, &s.TenantID, &s.ResiNumber, &s.SenderName, &s.SenderPhone,
		&s.ReceiverName, &s.ReceiverPhone, &s.Origin, &s.Destination, &s.CurrentStatus,
		&s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customErr.ErrNotFound
		}
		return nil, err
	}

	return &s, nil
}

func (r *postgresShipmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shipment, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, resi_number, sender_name, sender_phone, receiver_name, receiver_phone, origin, destination, current_status, created_at, updated_at
		FROM shipments
		WHERE id = $1 AND tenant_id = $2
	`

	var s domain.Shipment
	err = r.db.QueryRow(ctx, query, id, tenantID).Scan(
		&s.ID, &s.TenantID, &s.ResiNumber, &s.SenderName, &s.SenderPhone,
		&s.ReceiverName, &s.ReceiverPhone, &s.Origin, &s.Destination, &s.CurrentStatus,
		&s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customErr.ErrNotFound
		}
		return nil, err
	}

	return &s, nil
}

func (r *postgresShipmentRepo) List(ctx context.Context, page, limit int) ([]*domain.Shipment, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	query := `
		SELECT id, tenant_id, resi_number, sender_name, sender_phone, receiver_name, receiver_phone, origin, destination, current_status, created_at, updated_at
		FROM shipments
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shipments []*domain.Shipment
	for rows.Next() {
		var s domain.Shipment
		err := rows.Scan(
			&s.ID, &s.TenantID, &s.ResiNumber, &s.SenderName, &s.SenderPhone,
			&s.ReceiverName, &s.ReceiverPhone, &s.Origin, &s.Destination, &s.CurrentStatus,
			&s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		shipments = append(shipments, &s)
	}

	return shipments, rows.Err()
}

func (r *postgresShipmentRepo) AddEvent(ctx context.Context, e *domain.TrackingEvent) error {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return err
	}

	e.TenantID = tenantID
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}

	query := `
		INSERT INTO tracking_events (id, tenant_id, shipment_id, status, description, location)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING event_time, created_at
	`

	err = r.db.QueryRow(ctx, query, e.ID, e.TenantID, e.ShipmentID, e.Status, e.Description, e.Location).Scan(&e.EventTime, &e.CreatedAt)
	if err != nil {
		return err
	}

	// Update current status on shipment
	updateQuery := `UPDATE shipments SET current_status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND tenant_id = $3`
	_, updateErr := r.db.Exec(ctx, updateQuery, e.Status, e.ShipmentID, tenantID)

	return updateErr
}

func (r *postgresShipmentRepo) GetEvents(ctx context.Context, shipmentID uuid.UUID) ([]*domain.TrackingEvent, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, shipment_id, status, description, location, event_time, created_at
		FROM tracking_events
		WHERE shipment_id = $1 AND tenant_id = $2
		ORDER BY event_time DESC
	`

	rows, err := r.db.Query(ctx, query, shipmentID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.TrackingEvent
	for rows.Next() {
		var e domain.TrackingEvent
		err := rows.Scan(
			&e.ID, &e.TenantID, &e.ShipmentID, &e.Status, &e.Description, &e.Location, &e.EventTime, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}

	return events, rows.Err()
}
