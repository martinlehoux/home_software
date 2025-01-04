## Run

- `templ generate`
- `go run main.go server`

## Production

```sh
make home_software
scp home_software admin@framboise.local:

# https://askubuntu.com/questions/8653/how-to-keep-processes-running-after-ending-ssh-session
tmux attach-session -t home_software
home_software server --port=8081 --database=home_software.db
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
