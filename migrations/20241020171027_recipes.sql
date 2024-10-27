-- migrate:up
CREATE TABLE "recipes" (
    id integer PRIMARY KEY,
    title text NOT NULL,
    notes text NOT NULL
);

-- migrate:down
DROP TABLE "recipes";
