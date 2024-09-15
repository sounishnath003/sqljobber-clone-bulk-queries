-- This script only contains the table creation statements and does not fully represent the table in the database. Do not use it as a backup.

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS t_jobs;

-- Table Definition
CREATE TABLE "public"."t_jobs" (
    "id" int4 NOT NULL DEFAULT nextval('t_jobs'::regclass),
    "job_name" varchar NOT NULL,
    "ttl" int4 NOT NULL,
    "status" varchar NOT NULL,
    PRIMARY KEY ("id")
);


-- Indices
CREATE UNIQUE INDEX t_jobs ON public.t_jobs USING btree (id);