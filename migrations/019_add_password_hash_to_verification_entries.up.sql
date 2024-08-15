-- Step 1: Add the column as nullable
ALTER TABLE verification_entries ADD COLUMN password_hash TEXT;

-- Step 2: Update existing rows (replace with a secure default if needed)
UPDATE verification_entries SET password_hash = '' WHERE password_hash IS NULL;

-- Step 3: Set the column to NOT NULL
ALTER TABLE verification_entries ALTER COLUMN password_hash SET NOT NULL;