## Run

- `templ generate`
- `go run main.go server --port=8081 --database=home_software.db`

## Production

```sh
make home_software
scp home_software admin@framboise.local:

systemd-run --user /home/admin/home_software server --database=home_software.db --port=8081
```

## TODO

- sqlx seems good abstraction here
- use tailwind
- more config (db name, port)

- **Cleaning**

  - interactive filter of chores to record
  - possibilité de cocher tous les aspirateurs
  - score, score over time
  - navheader
  - global score
  - choose any date to submit
  - Add new chore

- **Recipes**

  - simple web interface
  - Recipes
    - Title
    - Notes
  - Week suggestions
    - Week (Monday date)
    - Recipe ID
  - Cooking
    - Recipe ID
    - Date
  - Random
  - Last made
  - Last suggested
  - Via telegram / email
  - rarest ingredients of the recipe
