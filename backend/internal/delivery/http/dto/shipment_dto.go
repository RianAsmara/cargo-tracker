package dto

import (
	"time"

	"github.com/RianAsmara/cargo-tracker/internal/domain"
	"github.com/google/uuid"
)

// Request DTOs
type CreateShipmentReq struct {
	ResiNumber    string `json:"resi_number" validate:"required"`
	SenderName    string `json:"sender_name"`
	SenderPhone   string `json:"sender_phone"`
	ReceiverName  string `json:"receiver_name"`
	ReceiverPhone string `json:"receiver_phone"`
	Origin        string `json:"origin" validate:"required"`
	Destination   string `json:"destination" validate:"required"`
}

type AddEventReq struct {
	Status      string `json:"status" validate:"required"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

// Response DTOs
type ShipmentRes struct {
	ID            uuid.UUID `json:"id"`
	ResiNumber    string    `json:"resi_number"`
	SenderName    string    `json:"sender_name"`
	SenderPhone   string    `json:"sender_phone"`
	ReceiverName  string    `json:"receiver_name"`
	ReceiverPhone string    `json:"receiver_phone"`
	Origin        string    `json:"origin"`
	Destination   string    `json:"destination"`
	CurrentStatus string    `json:"current_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TrackingEventRes struct {
	ID          uuid.UUID `json:"id"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	EventTime   time.Time `json:"event_time"`
}

type PublicTrackingRes struct {
	Shipment ShipmentRes        `json:"shipment"`
	Events   []TrackingEventRes `json:"events"`
}

// Mapper Functions
func ToShipmentRes(s *domain.Shipment) ShipmentRes {
	return ShipmentRes{
		ID:            s.ID,
		ResiNumber:    s.ResiNumber,
		SenderName:    s.SenderName,
		SenderPhone:   s.SenderPhone,
		ReceiverName:  s.ReceiverName,
		ReceiverPhone: s.ReceiverPhone,
		Origin:        s.Origin,
		Destination:   s.Destination,
		CurrentStatus: s.CurrentStatus,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

func ToEventRes(e *domain.TrackingEvent) TrackingEventRes {
	return TrackingEventRes{
		ID:          e.ID,
		Status:      e.Status,
		Description: e.Description,
		Location:    e.Location,
		EventTime:   e.EventTime,
	}
}

func ToPublicTrackingRes(s *domain.Shipment, events []*domain.TrackingEvent) PublicTrackingRes {
	eventRes := make([]TrackingEventRes, 0, len(events))
	for _, e := range events {
		eventRes = append(eventRes, ToEventRes(e))
	}

	return PublicTrackingRes{
		Shipment: ToShipmentRes(s),
		Events:   eventRes,
	}
}
