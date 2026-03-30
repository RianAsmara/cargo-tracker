package worker

import (
	"context"
	"log"
	"time"

	"github.com/RianAsmara/cargo-tracker/internal/domain"
)

type DelayDetector struct {
	repo     domain.DelayWorkerRepository
	interval time.Duration
}

func NewDelayDetector(repo domain.DelayWorkerRepository, interval time.Duration) *DelayDetector {
	return &DelayDetector{
		repo:     repo,
		interval: interval,
	}
}

// Start initiates the infinite loop matching delay logic safely.
// Pass a context that will be cancelled upon application server shutdown.
func (d *DelayDetector) Start(ctx context.Context) {
	log.Printf("Starting Delay Detector Worker with interval: %v", d.interval)
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Delay Detector shutting down gracefully...")
			return
		case <-ticker.C:
			d.processStagnantShipments(ctx)
		}
	}
}

func (d *DelayDetector) processStagnantShipments(ctx context.Context) {
	shipments, err := d.repo.FindStagnantShipments(ctx)
	if err != nil {
		log.Printf("Worker Error finding stagnant shipments: %v", err)
		return
	}

	if len(shipments) == 0 {
		return // No stagnant shipments found efficiently skipping writes.
	}

	log.Printf("Delay Worker flag cycle: Encountered %d stagnant shipment(s)", len(shipments))

	for _, s := range shipments {
		err := d.repo.MarkAsDelayed(ctx, s.ID, s.TenantID)
		if err != nil {
			log.Printf("Worker Error marking shipment %v as delayed: %v", s.ID, err)
			continue
		}
		log.Printf("Shipment %v (Tenant %v) successfully flagged as delayed", s.ResiNumber, s.TenantID)
		
		// Optional: If we had a robust Message Queue here like Redis/RabbitMQ,
		// we could push a notification event allowing the async infrastructure dispatcher 
		// to trigger a webhook notifying the Tenant's third-party CRM that the shipment is delayed.
	}
}
