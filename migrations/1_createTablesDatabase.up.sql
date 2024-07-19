CREATE SCHEMA IF NOT EXISTS humanresource;

CREATE SCHEMA IF NOT EXISTS tasks;

CREATE TABLE IF NOT EXISTS humanresource.users
(
    user_id INTEGER NOT NULL,
    passport_serie SMALLINT NOT NULL,
    passport_number SMALLINT NOT NULL,
    surname VARCHAR(32) NOT NULL,
    name VARCHAR(32) NOT NULL,
    patronymic VARCHAR(32) NOT NULL,
    address VARCHAR(32) NOT NULL,
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT uq_users_passport_serie_passport_number PRIMARY KEY (passport_serie,passport_number)
    );

CREATE TABLE IF NOT EXISTS tasks.tasks
(
    tasks_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    CONSTRAINT pk_tasks_tasks_id_user_id UNIQUE (tasks_id,user_id)
    );

CREATE SEQUENCE IF NOT EXISTS humanresource.user_sq AS INTEGER START WITH 1;

CREATE SEQUENCE IF NOT EXISTS tasks.task_sq AS INTEGER START WITH 1;

