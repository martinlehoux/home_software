-- migrate:up
CREATE TABLE "recipe_suggestions" (
    id integer PRIMARY KEY,
    recipe_id integer NOT NULL,
    suggested_at text NOT NULL,

    FOREIGN KEY (recipe_id) REFERENCES recipes (id)
);

-- migrate:down
DROP TABLE "recipe_suggestions";
