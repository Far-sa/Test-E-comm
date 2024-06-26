CREATE TABLE tokens (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL,
  token TEXT NOT NULL UNIQUE, -- Unique constraint on token
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expiration TIMESTAMP NOT NULL,
  CONSTRAINT idx_jwt_tokens_expiration_user_id UNIQUE (expiration, user_id) -- Composite unique constraint
);

DROP TABLE IF EXISTS tokens;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
