CREATE TABLE IF NOT EXISTS projects
(
    project character varying(80) NOT NULL,
    repository character varying(80),
    CONSTRAINT projects_pkey PRIMARY KEY (project)
);