package usecase

import (
	"context"

	"github.com/RianAsmara/cargo-tracker/internal/domain"
	"github.com/RianAsmara/cargo-tracker/internal/infrastructure"
	"github.com/google/uuid"
)

type shipmentUseCase struct {
	shipmentRepo domain.ShipmentRepository
	tenantRepo   domain.TenantRepository
	webhook      infrastructure.WebhookDispatcher
}

func NewShipmentUseCase(sr domain.ShipmentRepository, tr domain.TenantRepository, w infrastructure.WebhookDispatcher) domain.ShipmentUseCase {
	return &shipmentUseCase{
		shipmentRepo: sr,
		tenantRepo:   tr,
		webhook:      w,
	}
}

func (u *shipmentUseCase) CreateShipment(ctx context.Context, shipment *domain.Shipment) (*domain.Shipment, error) {
	if err := u.shipmentRepo.Create(ctx, shipment); err != nil {
		return nil, err
	}
	// Fetch complete newly created entity
	return u.shipmentRepo.GetByResi(ctx, shipment.ResiNumber)
}

func (u *shipmentUseCase) GetShipmentByResi(ctx context.Context, resiNumber string) (*domain.Shipment, []*domain.TrackingEvent, error) {
	shipment, err := u.shipmentRepo.GetByResi(ctx, resiNumber)
	if err != nil {
		return nil, nil, err
	}

	events, err := u.shipmentRepo.GetEvents(ctx, shipment.ID)
	if err != nil {
		return nil, nil, err
	}

	return shipment, events, nil
}

func (u *shipmentUseCase) ListShipments(ctx context.Context) ([]*domain.Shipment, error) {
	// Let's hardcode pagination simply for example or accept in params
	return u.shipmentRepo.List(ctx, 1, 50)
}

func (u *shipmentUseCase) AddEvent(ctx context.Context, shipmentID uuid.UUID, event *domain.TrackingEvent) (*domain.TrackingEvent, error) {
	event.ShipmentID = shipmentID
	if err := u.shipmentRepo.AddEvent(ctx, event); err != nil {
		return nil, err
	}

	// Trigger webhook async
	tenantID, ok := ctx.Value(domain.TenantContextKey).(uuid.UUID)
	if ok {
		go func(id uuid.UUID, evt *domain.TrackingEvent) {
			_ = context.Background()
			// Mock webhook dispatcher call.
			// Error handling swallowed for async webhook mock.
			_ = id
			_ = evt
		}(tenantID, event)
	}

	return event, nil
}
