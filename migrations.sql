-- This script only contains the table creation statements and does not fully represent the table in the database. Do not use it as a backup.
-- Sequence and defined type

CREATE SEQUENCE IF NOT EXISTS entries_id;

-- Table Definition
CREATE TABLE IF NOT EXISTS "public"."entries" (
	"id" int4 NOT NULL DEFAULT nextval('entries_id'::regclass),
	"user_id" int4 NOT NULL,
	"amount" float NOT NULL,
	"entry_date" date NOT NULL,
	"timestamp" timestamp NOT NULL,
	PRIMARY KEY ("id", "user_id")
);

-- Indices
CREATE UNIQUE INDEX entries ON public.entries
USING btree (user_id);

CREATE UNIQUE INDEX entries ON public.entries
USING btree (user_id, entry_date);


-- Insert Dummy Records:
-- Insert Dummy Records
INSERT INTO "public"."entries" (id, user_id, amount, entry_date, timestamp) VALUES 
(1, 1, 100.0, '2023-01-01', '2023-01-01 10:00:00'),
(2, 2, 150.5, '2023-01-02', '2023-01-02 11:00:00'),
(3, 1, 200.0, '2023-01-03', '2023-01-03 12:00:00'),
(4, 3, 250.75, '2023-01-04', '2023-01-04 13:00:00'),
(5, 2, 300.0, '2023-01-05', '2023-01-05 14:00:00'),
(6, 4, 350.25, '2023-01-06', '2023-01-06 15:00:00'),
(7, 1, 400.0, '2023-01-07', '2023-01-07 16:00:00'),
(8, 3, 450.5, '2023-01-08', '2023-01-08 17:00:00'),
(9, 2, 500.0, '2023-01-09', '2023-01-09 18:00:00'),
(10, 4, 550.75, '2023-01-10', '2023-01-10 19:00:00'),
(11, 1, 600.0, '2023-01-11', '2023-01-11 20:00:00'),
(12, 3, 650.5, '2023-01-12', '2023-01-12 21:00:00'),
(13, 2, 700.0, '2023-01-13', '2023-01-13 22:00:00'),
(14, 4, 750.25, '2023-01-14', '2023-01-14 23:00:00'),
(15, 1, 800.0, '2023-01-15', '2023-01-15 24:00:00');