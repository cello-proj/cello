CREATE USER cello WITH PASSWORD '1234';
CREATE TABLE IF NOT EXISTS projects
(
    project character varying(80) NOT NULL,
    repository character varying(200),
    CONSTRAINT projects_pkey PRIMARY KEY (project)
);
CREATE TABLE IF NOT EXISTS tokens
(
    token_id VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    project VARCHAR(80) NOT NULL,
    CONSTRAINT tokens_pkey PRIMARY KEY (token_id),
    FOREIGN KEY (project) REFERENCES projects(project) on delete cascade on update cascade
);
GRANT ALL PRIVILEGES ON tokens TO cello;
GRANT ALL PRIVILEGES ON projects TO cello;
