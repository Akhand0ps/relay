CREATE TABLE IF NOT EXISTS endpoints (
    id VARCHAR(64) PRIMARY KEY,
    url TEXT NOT NULL,
    event_types TEXT[] NOT NULL,
    secret VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL DEFAULT 'default_tenant',
    type VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    idempotency_key VARCHAR(128),
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_tenant_idempotency UNIQUE (tenant_id,idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_event_type ON events(type);
CREATE INDEX IF NOT EXISTS idx_event_idempotency ON events(tenant_id,idempotency_key);



-- deliveries tb

CREATE TABLE IF NOT EXISTS deliveries (

    id VARCHAR(64) PRIMARY KEY,
    event_id VARCHAR(64) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    endpoint_id VARCHAR(64) NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    attempt_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
    
);


--AUDIT LOG
CREATE TABLE IF NOT EXISTS delivery_attempts(

    id VARCHAR(64) PRIMARY KEY,
    delivery_id VARCHAR(64) NOT NULL REFERENCES deliveries(id) ON DELETE CASCADE,
    status_code INT,
    error_message TEXT,
    attempted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);


