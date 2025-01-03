## Run

- `templ generate`
- `go run main.go server`

## Production

```sh
set PKG_CONFIG_LIBDIR /usr/lib/aarch64-linux-gnu/pkgconfig
set CC "zig cc -target aarch64-linux-gnu -isystem /usr/include -L/usr/lib/aarch64-linux-gnu"
set CXX "zig c++ -target aarch64-linux-gnu -isystem /usr/include -L/usr/lib/aarch64-linux-gnu"
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build
scp home_software admin@framboise.local

# https://askubuntu.com/questions/8653/how-to-keep-processes-running-after-ending-ssh-session
tmux attach-session -t home_software
```

## TODO

- sqlx seems good abstraction here
- use tailwind
- no cgo (dbmate sqlite driver, sqlite) to simplify compilation (https://github.com/amacneil/dbmate/blob/v2.24.2/pkg/driver/sqlite/sqlite.go)

- **Cleaning**

  - interactive filter of chores to record
  - possibilité de cocher tous les aspirateurs
  - score, score over time
  - navheader
  - global score
  - lowest % complete on top
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
