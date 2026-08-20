CREATE TABLE change_events (id TEXT PRIMARY KEY, tenant_id TEXT, flag_key TEXT, version BIGINT, action TEXT, payload JSONB, created_at TIMESTAMPTZ NOT NULL);
CREATE TABLE client_cursors (client_id TEXT PRIMARY KEY, tenant_id TEXT, cursor BIGINT NOT NULL, updated_at TIMESTAMPTZ NOT NULL);
