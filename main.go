package main

import (
	"bufio"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	"github.com/martinlehoux/home_software/cleaning"
	"github.com/martinlehoux/home_software/utils"
	"github.com/martinlehoux/kagamigo/kcore"
	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var databaseName string
var db *sql.DB

var cleaningDisplayCmd = &cobra.Command{
	Use:   "display",
	Short: "Display cleaning routines",
	Run: func(cmd *cobra.Command, args []string) {
		routinesRooms := cleaning.RoutinesRooms(db)
		for _, room := range routinesRooms {
			fmt.Printf("%s (%d/%d):\n", room.Name, room.DoneCount(), len(room.Routines))
			for _, routine := range room.Routines {
				title := strings.Split(routine.Title, "/")[1]
				println("  ", routine.LastRecordedAtString(), title)
			}
		}
	},
}

var cleaningRecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record a routine",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: record <routine>")
			os.Exit(1)
		}
		var err error
		routineIDs := cleaning.MatchingRoutineIDs(db, args[0])
		log.Printf("recording %d routines", len(routineIDs))
		for _, routineID := range routineIDs {
			_, err = db.Exec("insert into record (routine_id, recorded_at) values (?, ?)", routineID, time.Now().Format(time.DateOnly))
			kcore.Expect(err, "failed to insert record")
		}
	},
}

var recipesRegisterCmd = &cobra.Command{
	Use: "register",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Recipe title? > ")
		r := bufio.NewReader(os.Stdin)
		title, err := r.ReadString('\n')
		kcore.Expect(err, "failed to read input")
		// TODO: check not already registered (or close)
		_, err = db.Exec("insert into recipes (title, notes) values (?, '')", title)
		kcore.Expect(err, "failed to insert recipe")
		log.Printf("recipe registered")
	},
}

func getAlreadySuggestedRecipes(now time.Time) []Recipe {
	endOfLastWeek := utils.EndOfWeek(now).Add(-time.Hour * 24 * 7)
	rows, err := db.Query("select recipes.id, recipes.title from recipe_suggestions left join recipes on recipe_suggestions.recipe_id = recipes.id  where suggested_at >= ?", endOfLastWeek.Format(time.DateOnly))
	kcore.Expect(err, "failed to query database")
	defer rows.Close()

	alreadySuggested := []Recipe{}
	for rows.Next() {
		var recipe Recipe
		kcore.Expect(rows.Scan(&recipe.ID, &recipe.Title), "failed to scan row")
		alreadySuggested = append(alreadySuggested, recipe)
	}
	return alreadySuggested
}

type Recipe struct {
	ID    int
	Title string
}

var recipesSuggestCmd = &cobra.Command{
	Use: "suggest",
	Run: func(cmd *cobra.Command, args []string) {
		alreadySuggested := getAlreadySuggestedRecipes(time.Now())
		if len(alreadySuggested) > 0 {
			fmt.Println("Already suggested:")
			for _, recipe := range alreadySuggested {
				fmt.Println("  ", recipe.Title)
			}
			return
		}
		maxRecipes := 10
		rows, err := db.Query("select id, title from recipes")
		kcore.Expect(err, "failed to query database")
		defer rows.Close()
		recipes := []Recipe{}
		for rows.Next() {
			var recipe Recipe
			kcore.Expect(rows.Scan(&recipe.ID, &recipe.Title), "failed to scan row")
			recipes = append(recipes, recipe)
		}
		recipes = lo.Shuffle(recipes)[:min(maxRecipes, len(recipes))]
		tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
		kcore.Expect(err, "failed to begin transaction")
		for _, recipe := range recipes {
			_, err = tx.Exec("insert into recipe_suggestions (recipe_id, suggested_at) values (?, ?)", recipe.ID, time.Now().Format(time.DateOnly))
			kcore.Expect(err, "failed to insert recipe suggestion")
		}
		kcore.Expect(tx.Commit(), "failed to commit transaction")
		for _, recipe := range recipes {
			fmt.Printf("Suggested: %s\n", recipe.Title)
		}
	},
}

var serverPort string
var serverCmd = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		http.HandleFunc("GET /cleaning/", func(res http.ResponseWriter, req *http.Request) {
			routinesRooms := cleaning.RoutinesRooms(db)
			kcore.RenderPage(ctx, cleaning.CleaningPage(routinesRooms), res)
		})
		http.HandleFunc("POST /cleaning/record", func(res http.ResponseWriter, req *http.Request) {
			if err := req.ParseForm(); err != nil {
				http.Error(res, "Failed to parse form", http.StatusBadRequest)
				return
			}
			routineIDs := []int{}
			for key, values := range req.Form {
				// TODO: assert key prefix
				kcore.Assert(len(values) == 1, "expected 1 value")
				kcore.Assert(values[0] == "on", "expected 'on'")
				routineID, err := strconv.Atoi(key)
				kcore.Expect(err, "failed to convert routine ID to integer")
				routineIDs = append(routineIDs, routineID)
			}
			tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
			kcore.Expect(err, "failed to begin transaction")
			for _, routineID := range routineIDs {
				_, err = tx.Exec("insert into record (routine_id, recorded_at) values (?, ?)", routineID, time.Now().Format(time.DateOnly))
				kcore.Expect(err, "failed to insert record")
			}
			kcore.Expect(tx.Commit(), "failed to commit transaction")
			http.Redirect(res, req, "/cleaning/", http.StatusSeeOther)
		})
		log.Printf("listening on http://localhost:%s", serverPort)
		err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil)
		kcore.Expect(err, "failed to listen and serve")
		// TODO: graceful shutdown
	},
}

//go:embed migrations/*.sql
var migrationsFs embed.FS

var migrateCmd = &cobra.Command{
	Use: "migrate",
	Run: func(cmd *cobra.Command, args []string) {
		migrator := dbmate.New(&url.URL{Scheme: "sqlite", Opaque: databaseName})
		migrator.FS = migrationsFs
		migrator.MigrationsDir = []string{"migrations"}
		migrator.SchemaFile = "schema.sql"
		log.Println("migrating database")
		kcore.Expect(migrator.Migrate(), "failed to migrate")
	},
}

func main() {
	var err error
	cmd := &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			db, err = sql.Open("sqlite3", databaseName)
			kcore.Expect(err, "failed to open database")
		},
	}
	defer func() {
		kcore.Expect(db.Close(), "failed to close database")
	}()
	cleaningCmd := &cobra.Command{
		Use: "cleaning",
	}
	cleaningCmd.AddCommand(cleaningRecordCmd)
	cleaningCmd.AddCommand(cleaningDisplayCmd)
	cmd.AddCommand(cleaningCmd)
	recipesCmd := &cobra.Command{
		Use: "recipes",
	}
	recipesCmd.AddCommand(recipesRegisterCmd)
	recipesCmd.AddCommand(recipesSuggestCmd)
	cmd.AddCommand(recipesCmd)
	serverCmd.Flags().StringVar(&serverPort, "port", "8081", "port to listen on")
	serverCmd.MarkFlagRequired("port")
	cmd.AddCommand(serverCmd)
	cmd.AddCommand(migrateCmd)

	cmd.PersistentFlags().StringVar(&databaseName, "database", "database.db", "database name")
	cmd.MarkPersistentFlagRequired("database")
	cmd.Execute()
}
