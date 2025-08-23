CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
  id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email       TEXT UNIQUE NOT NULL,
  pass_hash   TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- tabella dei refresh token emessi
CREATE TABLE IF NOT EXISTS refresh_tokens (
  jti           UUID PRIMARY KEY,               -- ID univoco del refresh
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash    TEXT NOT NULL,                  -- hash del refresh (mai salvare il token in chiaro)
  issued_at     TIMESTAMPTZ NOT NULL,
  expires_at    TIMESTAMPTZ NOT NULL,
  revoked_at    TIMESTAMPTZ,                    -- se valorizzato, token revocato
  replaced_by   UUID,                           -- jti del nuovo token (per rotazione)
  user_agent    TEXT,
  ip            INET
);

CREATE INDEX IF NOT EXISTS idx_refresh_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_valid ON refresh_tokens(expires_at);
