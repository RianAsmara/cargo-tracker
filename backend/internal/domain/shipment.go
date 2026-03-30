package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ShipmentStatus string

const (
	StatusPending   ShipmentStatus = "PENDING"
	StatusInTransit ShipmentStatus = "IN_TRANSIT"
	StatusDelivered ShipmentStatus = "DELIVERED"
	StatusDelayed   ShipmentStatus = "DELAYED"
	StatusCanceled  ShipmentStatus = "CANCELED"
)

type Shipment struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	ResiNumber    string
	SenderName    string
	SenderPhone   string
	ReceiverName  string
	ReceiverPhone string
	Origin        string
	Destination   string
	CurrentStatus string
	IsDelayed     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TrackingEvent struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	ShipmentID  uuid.UUID
	Status      string
	Description string
	Location    string
	EventTime   time.Time
	CreatedAt   time.Time
}

type ShipmentRepository interface {
	Create(ctx context.Context, shipment *Shipment) error
	GetByResi(ctx context.Context, resiNumber string) (*Shipment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Shipment, error)
	List(ctx context.Context, limit, offset int) ([]*Shipment, error)
	AddEvent(ctx context.Context, event *TrackingEvent) error
	GetEvents(ctx context.Context, shipmentID uuid.UUID) ([]*TrackingEvent, error)
}

// Usecase accepts and returns DTOs (wait, clean architecture typically means usecase returns domains or DTOs,
// but the interface should be defined in Usecase without DTO imports to prevent cycle if not careful,
// or we just use defined DTO structs inside the domain package).
// Since the user asked for "DTO + entity separation", we will define input/output structures for the Usecase in the domain or dto package.
// Actually, standard Go approach is Usecase returns Entity, and Delivery layer converts Entity to Response DTO,
// and Delivery converts Request DTO to Entity before calling Usecase.
// Let's stick strictly to that: Usecase takes/returns Entity.

type ShipmentUseCase interface {
	CreateShipment(ctx context.Context, shipment *Shipment) (*Shipment, error)
	GetShipmentByResi(ctx context.Context, resiNumber string) (*Shipment, []*TrackingEvent, error)
	ListShipments(ctx context.Context) ([]*Shipment, error)
	AddEvent(ctx context.Context, shipmentID uuid.UUID, event *TrackingEvent) (*TrackingEvent, error)
}

// Background Worker Repository isolated from generic user requests
type DelayWorkerRepository interface {
	FindStagnantShipments(ctx context.Context) ([]*Shipment, error)
	MarkAsDelayed(ctx context.Context, shipmentID uuid.UUID, tenantID uuid.UUID) error
}
