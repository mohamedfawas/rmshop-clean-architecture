-- Add order_status column
ALTER TABLE orders
ADD COLUMN order_status VARCHAR(50) NOT NULL DEFAULT 'pending';

-- Add refund_status column
ALTER TABLE orders
ADD COLUMN refund_status VARCHAR(50);

-- Add updated_at column if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='orders' AND column_name='updated_at') THEN
        ALTER TABLE orders
        ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- Update existing rows to set updated_at to created_at
UPDATE orders
SET updated_at = created_at
WHERE updated_at IS NULL;

-- Make updated_at NOT NULL and set default
ALTER TABLE orders
ALTER COLUMN updated_at SET NOT NULL,
ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

-- Add trigger to automatically update updated_at
CREATE OR REPLACE FUNCTION update_orders_updated_at()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_orders_updated_at
BEFORE UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION update_orders_updated_at();