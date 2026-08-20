CREATE TABLE projects (id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, name TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL);
CREATE TABLE environments (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, name TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL);
CREATE TABLE flags (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, env_id TEXT NOT NULL, flag_key TEXT NOT NULL, value_type TEXT NOT NULL, default_value JSONB NOT NULL, version BIGINT NOT NULL, state TEXT NOT NULL);
CREATE TABLE flag_versions (flag_id TEXT NOT NULL, version BIGINT NOT NULL, snapshot JSONB NOT NULL, actor TEXT, reason TEXT, created_at TIMESTAMPTZ NOT NULL, PRIMARY KEY(flag_id,version));
