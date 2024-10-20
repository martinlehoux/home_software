package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/martinlehoux/kagamigo/kcore"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var db *sql.DB

type Routine struct {
	ID             int
	Title          string
	FrequencyWeeks int
}

type Record struct {
	ID         int
	RoutineID  int
	RecordedAt time.Time
}

func getAllRoutines(database *sql.DB) []Routine {
	rows, err := database.Query("select id, title, frequency_weeks from routine")
	kcore.Expect(err, "failed to query database")
	defer func() {
		kcore.Expect(rows.Close(), "failed to close rows")
	}()
	routines := []Routine{}
	for rows.Next() {
		var routine Routine
		kcore.Expect(rows.Scan(&routine.ID, &routine.Title, &routine.FrequencyWeeks), "failed to scan row")
		routines = append(routines, routine)
	}
	return routines
}

func getAllRecordsByRoutine(database *sql.DB) map[int][]Record {
	rows, err := database.Query("select id, routine_id, recorded_at from record")
	kcore.Expect(err, "failed to query database")
	defer func() {
		kcore.Expect(rows.Close(), "failed to close rows")
	}()
	recordsByRoutine := map[int][]Record{}
	for rows.Next() {
		var record Record
		var recordedAt string
		kcore.Expect(rows.Scan(&record.ID, &record.RoutineID, &recordedAt), "failed to scan row")
		record.RecordedAt, err = time.Parse(time.DateOnly, recordedAt)
		kcore.Expect(err, "failed to parse date")
		if _, ok := recordsByRoutine[record.RoutineID]; !ok {
			recordsByRoutine[record.RoutineID] = []Record{}
		}
		recordsByRoutine[record.RoutineID] = append(recordsByRoutine[record.RoutineID], record)
	}

	return recordsByRoutine
}

type ExpectedRoutine struct {
	Title          string
	LastRecordedAt time.Time
}

func getExpectedRoutines(routines []Routine, recordsByRoutine map[int][]Record) []ExpectedRoutine {
	now := time.Now()
	endOfWeek := now.AddDate(0, 0, (7-int(now.Weekday()))%7)
	expectedRoutines := []ExpectedRoutine{}
	for _, routine := range routines {
		records, ok := recordsByRoutine[routine.ID]
		if !ok {
			expectedRoutines = append(expectedRoutines, ExpectedRoutine{Title: routine.Title})
		} else {
			lastRecordedAt := time.Time{}
			for _, record := range records {
				if record.RecordedAt.After(lastRecordedAt) {
					lastRecordedAt = record.RecordedAt
				}
			}
			if endOfWeek.Sub(lastRecordedAt).Hours() > float64(routine.FrequencyWeeks*7*24) {
				expectedRoutines = append(expectedRoutines, ExpectedRoutine{Title: routine.Title, LastRecordedAt: lastRecordedAt})
			}
		}
	}
	return expectedRoutines
}

var displayCmd = &cobra.Command{
	Use:   "display",
	Short: "Display routines",
	Run: func(cmd *cobra.Command, args []string) {
		routines := getAllRoutines(db)
		recordsByRoutine := getAllRecordsByRoutine(db)
		routinesByRoom := map[string][]Routine{}
		for _, routine := range routines {
			parts := strings.Split(routine.Title, "/")
			kcore.Assert(len(parts) == 2, "expected 2 parts")
			room := parts[0]
			if _, ok := routinesByRoom[room]; !ok {
				routinesByRoom[room] = []Routine{}
			}
			routinesByRoom[room] = append(routinesByRoom[room], routine)
		}
		expectedRoutines := getExpectedRoutines(routines, recordsByRoutine)
		expectedRoutinesByRoom := map[string][]ExpectedRoutine{}
		for _, routine := range expectedRoutines {
			parts := strings.Split(routine.Title, "/")
			kcore.Assert(len(parts) == 2, "expected 2 parts")
			room := parts[0]
			if _, ok := expectedRoutinesByRoom[room]; !ok {
				expectedRoutinesByRoom[room] = []ExpectedRoutine{}
			}
			expectedRoutinesByRoom[room] = append(expectedRoutinesByRoom[room], routine)
		}
		for room, expectedRoutines := range expectedRoutinesByRoom {
			fmt.Printf("%s (%d/%d):\n", room, len(routinesByRoom[room])-len(expectedRoutines), len(routinesByRoom[room]))
			for _, routine := range expectedRoutines {
				lastRecordedAt := "never" + strings.Repeat(" ", 7)
				if !routine.LastRecordedAt.IsZero() {
					lastRecordedAt = routine.LastRecordedAt.Format(time.DateOnly) + "  "
				}
				title := strings.Split(routine.Title, "/")[1]
				println("  ", lastRecordedAt, title)
			}
		}
	},
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record a routine",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: record <routine>")
			os.Exit(1)
		}
		routineSearch := args[0]
		rows, err := db.Query("select id from routine where title like ?", routineSearch)
		kcore.Expect(err, "failed to query database")
		routineIDs := []int{}
		for rows.Next() {
			var routineID int
			kcore.Expect(rows.Scan(&routineID), "failed to scan row")
			routineIDs = append(routineIDs, routineID)
		}
		rows.Close()
		log.Printf("recording %d routines", len(routineIDs))
		for _, routineID := range routineIDs {
			_, err = db.Exec("insert into record (routine_id, recorded_at) values (?, ?)", routineID, time.Now().Format(time.DateOnly))
			kcore.Expect(err, "failed to insert record")
		}
	},
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "database.db")
	kcore.Expect(err, "failed to open database")
	defer func() {
		kcore.Expect(db.Close(), "failed to close database")
	}()
	var cmd = &cobra.Command{}
	var cleaningCmd = &cobra.Command{
		Use: "cleaning",
	}
	cleaningCmd.AddCommand(recordCmd)
	cleaningCmd.AddCommand(displayCmd)
	cmd.AddCommand(cleaningCmd)
	cmd.Execute()
}
