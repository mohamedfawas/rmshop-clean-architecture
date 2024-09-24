ALTER TABLE cancellation_requests DROP CONSTRAINT IF EXISTS fk_order;
ALTER TABLE cancellation_requests DROP CONSTRAINT IF EXISTS fk_user;
ALTER TABLE cancellation_requests DROP CONSTRAINT IF EXISTS uq_order_cancellation;

DROP TABLE IF EXISTS cancellation_requests;