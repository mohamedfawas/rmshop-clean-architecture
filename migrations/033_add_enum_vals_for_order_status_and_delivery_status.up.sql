ALTER TABLE orders DROP COLUMN IF EXISTS order_status;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_status;

-- Order Status: Overall state of the order
ALTER TABLE orders
ADD COLUMN order_status VARCHAR(20) CHECK (order_status IN (
    'pending_payment',
    'processing',
    'shipped',
    'completed',
    'cancelled',
    'refunded'
)) DEFAULT 'pending_payment';

-- Delivery Status: State of the delivery process
ALTER TABLE orders
ADD COLUMN delivery_status VARCHAR(20) CHECK (delivery_status IN (
    'pending',
    'in_transit',
    'out_for_delivery',
    'delivered',
    'failed_attempt',
    'returned_to_sender'
)) DEFAULT 'pending';

-- Create indexes for faster queries
CREATE INDEX idx_orders_order_status ON orders(order_status);
CREATE INDEX idx_orders_delivery_status ON orders(delivery_status);