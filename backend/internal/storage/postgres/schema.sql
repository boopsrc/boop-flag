CREATE TABLE IF NOT EXISTS users (
    id             UUID PRIMARY KEY,
    google_sub     TEXT NOT NULL UNIQUE,
    email          TEXT NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    display_name   TEXT NOT NULL DEFAULT '',
    avatar_url     TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    onboarded_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_google_sub ON users (google_sub);
