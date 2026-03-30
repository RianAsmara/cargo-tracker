DROP INDEX IF EXISTS idx_shipments_delay_check;
ALTER TABLE shipments DROP COLUMN is_delayed;
ALTER TABLE tenants DROP COLUMN delay_threshold_hours;
