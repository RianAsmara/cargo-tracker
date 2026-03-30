ALTER TABLE tenants ADD COLUMN delay_threshold_hours INTEGER DEFAULT 24;
ALTER TABLE shipments ADD COLUMN is_delayed BOOLEAN DEFAULT FALSE;

-- Optimization for the background worker scanner
CREATE INDEX idx_shipments_delay_check ON shipments(is_delayed, current_status, updated_at);
