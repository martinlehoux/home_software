CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(255) primary key);
CREATE TABLE IF NOT EXISTS "routine" (
  id integer PRIMARY KEY,
  title text NOT NULL,
  frequency_weeks integer NOT NULL
);
CREATE TABLE IF NOT EXISTS "record" (
  id integer PRIMARY KEY,
  routine_id integer NOT NULL,
  recorded_at text NOT NULL,
  FOREIGN KEY (routine_id) REFERENCES ROUTINE (id)
);
CREATE TABLE IF NOT EXISTS "recipes" (
    id integer PRIMARY KEY,
    title text NOT NULL,
    notes text NOT NULL
);
CREATE TABLE IF NOT EXISTS "recipe_suggestions" (
    id integer PRIMARY KEY,
    recipe_id integer NOT NULL,
    suggested_at text NOT NULL,

    FOREIGN KEY (recipe_id) REFERENCES recipes (id)
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20240922073559'),
  ('20241020171027'),
  ('20241020173804');
