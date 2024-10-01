ALTER TABLE cancellation_requests
ADD COLUMN is_stock_updated BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_cancellation_requests_is_stock_updated ON cancellation_requests(is_stock_updated);
