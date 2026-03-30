package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/RianAsmara/cargo-tracker/internal/delivery/http/dto"
	"github.com/RianAsmara/cargo-tracker/internal/domain"
	customErr "github.com/RianAsmara/cargo-tracker/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type DeliveryHandlers struct {
	shipmentUseCase domain.ShipmentUseCase
	jwtSecret       string
}

func NewHandlers(su domain.ShipmentUseCase, jwtSecret string) *DeliveryHandlers {
	return &DeliveryHandlers{
		shipmentUseCase: su,
		jwtSecret:       jwtSecret,
	}
}

// ============== PUBLIC ENDPOINTS ============== 
func (h *DeliveryHandlers) GetTracking(c echo.Context) error {
	resiNo := c.Param("resi")
	if resiNo == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"resi required"})
	}

	shipment, events, err := h.shipmentUseCase.GetShipmentByResi(c.Request().Context(), resiNo)
	if err != nil {
		if err == customErr.ErrNotFound {
			return c.JSON(http.StatusNotFound, ErrorResponse{"shipment not found"})
		}
		log.Printf("GetTracking error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{"internal server error"})
	}

	return c.JSON(http.StatusOK, dto.ToPublicTrackingRes(shipment, events))
}

// ============== ADMIN ENDPOINTS ============== 

// POST /auth/login
func (h *DeliveryHandlers) AdminLogin(c echo.Context, userRepo domain.UserRepository) error {
	var req dto.LoginReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"invalid request body"})
	}

	// Validate (assuming standard validator exists, but simple check here)
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"email and password are required"})
	}

	user, err := userRepo.GetByEmail(c.Request().Context(), req.Email)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{"invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{"invalid credentials"})
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       user.ID.String(),
		"tenant_id": user.TenantID.String(),
		"role":      user.Role,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(h.jwtSecret))

	return c.JSON(http.StatusOK, dto.LoginRes{Token: tokenString})
}

// POST /shipments
func (h *DeliveryHandlers) CreateShipment(c echo.Context) error {
	var req dto.CreateShipmentReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"invalid request body"})
	}

	// Map DTO to Entity
	newShipment := &domain.Shipment{
		ResiNumber:    req.ResiNumber,
		SenderName:    req.SenderName,
		SenderPhone:   req.SenderPhone,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Origin:        req.Origin,
		Destination:   req.Destination,
	}

	createdShipment, err := h.shipmentUseCase.CreateShipment(c.Request().Context(), newShipment)
	if err != nil {
		log.Printf("CreateShipment error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{"failed to create shipment"})
	}

	return c.JSON(http.StatusCreated, dto.ToShipmentRes(createdShipment))
}

// GET /shipments/:resi
func (h *DeliveryHandlers) GetShipment(c echo.Context) error {
	resiNo := c.Param("resi")
	if resiNo == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"resi required"})
	}

	shipment, _, err := h.shipmentUseCase.GetShipmentByResi(c.Request().Context(), resiNo)
	if err != nil {
		if err == customErr.ErrNotFound {
			return c.JSON(http.StatusNotFound, ErrorResponse{"shipment not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{"internal server error"})
	}

	// For admin view, maybe we just return shipment details 
	// Or we can return full public tracking format if admin needs events too.
	// We'll just return shipment dto.
	return c.JSON(http.StatusOK, dto.ToShipmentRes(shipment))
}

// POST /shipments/:id/events
func (h *DeliveryHandlers) AddEvent(c echo.Context) error {
	idParam := c.Param("id")
	shipmentID, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"invalid shipment ID format"})
	}

	var req dto.AddEventReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"invalid request body"})
	}

	if req.Status == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"status is required"})
	}

	eventEntity := &domain.TrackingEvent{
		Status:      req.Status,
		Location:    req.Location,
		Description: req.Description,
	}

	createdEvent, err := h.shipmentUseCase.AddEvent(c.Request().Context(), shipmentID, eventEntity)
	if err != nil {
		log.Printf("AddEvent error: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{"failed to add event"})
	}

	return c.JSON(http.StatusCreated, dto.ToEventRes(createdEvent))
}
