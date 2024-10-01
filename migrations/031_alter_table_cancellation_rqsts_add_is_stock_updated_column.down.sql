ALTER TABLE cancellation_requests
DROP COLUMN IF EXISTS is_stock_updated;

DROP INDEX IF EXISTS idx_cancellation_requests_is_stock_updated;
