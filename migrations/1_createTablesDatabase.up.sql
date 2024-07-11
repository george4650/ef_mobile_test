
CREATE TABLE IF NOT EXISTS projects
(
    id serial,
    project_uuid character varying,
    name character varying NOT NULL,
    PRIMARY KEY (id, project_uuid)
);

CREATE TABLE IF NOT EXISTS videos
(
    id  serial PRIMARY KEY,
    name  character varying NOT NULL,
    login character varying NOT NULL,
    session_id character varying,
    created_at timestamp NOT NULL,
    fullpath character varying NOT NULL,
    mac_addr character varying NOT NULL,
    ip_addr character varying NOT NULL,
    get_args character varying,
    filename character varying,
    project_id character varying NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects (project_uuid)
);

CREATE UNIQUE INDEX project_project_uuid ON projects (project_uuid);