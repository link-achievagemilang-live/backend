-- Create urls table
CREATE TABLE IF NOT EXISTS urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(8) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,
    user_id UUID,
    click_count BIGINT NOT NULL DEFAULT 0,
    last_accessed TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_expires_at ON urls(expires_at) WHERE expires_at IS NOT NULL;

-- Create a function to clean up expired URLs (optional, can be called periodically)
CREATE OR REPLACE FUNCTION delete_expired_urls()
RETURNS void AS $$
BEGIN
    DELETE FROM urls WHERE expires_at IS NOT NULL AND expires_at < NOW();
END;
$$ LANGUAGE plpgsql;
