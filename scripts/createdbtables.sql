CREATE USER cello WITH PASSWORD '1234';
CREATE TABLE IF NOT EXISTS projects
(
    project character varying(80) NOT NULL,
    repository character varying(200),
    CONSTRAINT projects_pkey PRIMARY KEY (project)
);
GRANT ALL PRIVILEGES ON projects TO cello;
