CREATE TABLE IF NOT EXISTS argocloudops.projects
(
    project character varying(80) NOT NULL,
    repository character varying(80),
    CONSTRAINT projects_pkey PRIMARY KEY (project)
);