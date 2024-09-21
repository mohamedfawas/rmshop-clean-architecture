ALTER TABLE cancellation_requests DROP CONSTRAINT IF EXISTS fk_cancellation_requests_order;
ALTER TABLE cancellation_requests DROP CONSTRAINT IF EXISTS fk_cancellation_requests_user;
DROP INDEX IF EXISTS idx_cancellation_requests_order_id;
DROP INDEX IF EXISTS idx_cancellation_requests_user_id;
DROP INDEX IF EXISTS idx_cancellation_requests_status;
DROP TABLE IF EXISTS cancellation_requests;
