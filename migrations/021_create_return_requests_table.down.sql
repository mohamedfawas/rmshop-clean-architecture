ALTER TABLE return_requests DROP CONSTRAINT IF EXISTS fk_return_requests_order;
ALTER TABLE return_requests DROP CONSTRAINT IF EXISTS fk_return_requests_user;

DROP INDEX IF EXISTS idx_return_requests_order_id;
DROP INDEX IF EXISTS idx_return_requests_user_id;
DROP INDEX IF EXISTS idx_return_requests_is_approved;

DROP TABLE IF EXISTS return_requests;
