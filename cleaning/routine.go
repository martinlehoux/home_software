package cleaning

import (
	"database/sql"
	"slices"
	"strings"
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

func allRoutines(database *sql.DB) []Routine {
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

func allRecordsByRoutine(database *sql.DB) map[int][]Record {
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
	ID             int
	Title          string
	lastRecordedAt time.Time
}

func (er *ExpectedRoutine) LastRecordedAt() string {
	if er.lastRecordedAt.IsZero() {
		return "never"
	}

	return er.lastRecordedAt.Format(time.DateOnly)
}

func IsExpectedRoutine(routine Routine, records []Record) (ExpectedRoutine, bool) {
	lastRecordedAt := time.Time{}
	for _, record := range records {
		if record.RecordedAt.After(lastRecordedAt) {
			lastRecordedAt = record.RecordedAt
		}
	}
	if utils.EndOfWeek(time.Now()).Sub(lastRecordedAt).Hours() > float64(routine.FrequencyWeeks*7*24) {
		return ExpectedRoutine{ID: routine.ID, Title: routine.Title, lastRecordedAt: lastRecordedAt}, true
	}
	return ExpectedRoutine{}, false
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

type Room struct {
	Name             string
	ExpectedRoutines []ExpectedRoutine
	Routines         []Routine
}

func RoutinesRooms(db *sql.DB) []Room {
	roomsByName := map[string]Room{}
	routines := allRoutines(db)
	recordsByRoutine := allRecordsByRoutine(db)
	for _, routine := range routines {
		parts := strings.Split(routine.Title, "/")
		kcore.Assert(len(parts) == 2, "expected 2 parts")
		roomName := parts[0]
		room, ok := roomsByName[roomName]
		if !ok {
			room = Room{Name: roomName}
		}
		room.Routines = append(room.Routines, routine)
		if expected, ok := IsExpectedRoutine(routine, recordsByRoutine[routine.ID]); ok {
			room.ExpectedRoutines = append(room.ExpectedRoutines, expected)
		}
		roomsByName[roomName] = room
	}
	rooms := []Room{}
	for _, room := range roomsByName {
		slices.SortFunc(room.ExpectedRoutines, func(a, b ExpectedRoutine) int { return int(a.lastRecordedAt.Sub(b.lastRecordedAt).Nanoseconds()) })
		rooms = append(rooms, room)
	}
	slices.SortFunc(rooms, func(a, b Room) int { return len(a.ExpectedRoutines) - len(b.ExpectedRoutines) })
	return rooms
}
