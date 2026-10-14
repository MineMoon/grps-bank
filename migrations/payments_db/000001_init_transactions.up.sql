CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    idempotency_key VARCHAR(64) NOT NULL UNIQUE,
    from_account_id UUID NOT NULL,
    to_account_id UUID NOT NULL,
    amount NUMERIC(18, 4) NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_positive_amount CHECK (amount > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_idempotency ON transactions(idempotency_key);