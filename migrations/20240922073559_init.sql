-- migrate:up
CREATE TABLE "routine" (
  id integer PRIMARY KEY,
  title text NOT NULL,
  frequency_weeks integer NOT NULL
);

CREATE TABLE "record" (
  id integer PRIMARY KEY,
  routine_id integer NOT NULL,
  recorded_at text NOT NULL,
  FOREIGN KEY (routine_id) REFERENCES ROUTINE (id)
);

-- migrate:down
DROP TABLE "routine";

DROP TABLE "record";

