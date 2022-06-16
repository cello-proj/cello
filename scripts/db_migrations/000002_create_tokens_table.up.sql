ALTER TABLE projects ADD COLUMN IF NOT EXISTS token_ids text[];

CREATE TABLE IF NOT EXISTS tokens
(
    token_id VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ,
    CONSTRAINT tokens_pkey PRIMARY KEY (token_id)
);
GRANT ALL PRIVILEGES ON tokens TO cello;
