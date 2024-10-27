package cleaning

import (
	"database/sql"
	"time"

	"github.com/martinlehoux/home_software/utils"
	"github.com/martinlehoux/kagamigo/kcore"
)

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

func AllRoutines(database *sql.DB) []Routine {
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

func AllRecordsByRoutine(database *sql.DB) map[int][]Record {
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

func ExpectedRoutines(routines []Routine, recordsByRoutine map[int][]Record) []ExpectedRoutine {
	endOfWeek := utils.EndOfWeek(time.Now())
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

func MatchingRoutineIDs(db *sql.DB, routineSearch string) []int {
	rows, err := db.Query("select id from routine where title like ?", routineSearch)
	kcore.Expect(err, "failed to query database")
	routineIDs := []int{}
	defer rows.Close()
	for rows.Next() {
		var routineID int
		kcore.Expect(rows.Scan(&routineID), "failed to scan row")
		routineIDs = append(routineIDs, routineID)
	}
	return routineIDs
}
