package cleaning

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/martinlehoux/home_software/utils"
	"github.com/martinlehoux/kagamigo/kcore"
)

type Record struct {
	ID         int
	RoutineID  int
	RecordedAt time.Time
}

type Routine struct {
	ID             int
	Title          string
	FrequencyWeeks int
	records        []Record
}

func (routine Routine) LastRecordedAt() time.Time {
	lastRecordedAt := time.Time{}
	for _, record := range routine.records {
		if record.RecordedAt.After(lastRecordedAt) {
			lastRecordedAt = record.RecordedAt
		}
	}
	return lastRecordedAt
}

func (routine Routine) LastRecordedAtString() string {
	lastRecordedAt := routine.LastRecordedAt()
	if lastRecordedAt.IsZero() {
		return "never"
	}
	return lastRecordedAt.Format(time.DateOnly)
}

func (routine Routine) DueThisWeek() bool {
	return utils.EndOfWeek(time.Now()).Sub(routine.LastRecordedAt()).Hours() > float64(routine.FrequencyWeeks*7*24)
}

func allRoutines(database *sql.DB) []Routine {
	recordsByRoutine := allRecordsByRoutine(database)
	rows, err := database.Query("select id, title, frequency_weeks from routine")
	kcore.Expect(err, "failed to query database")
	defer func() {
		kcore.Expect(rows.Close(), "failed to close rows")
	}()
	routines := []Routine{}
	for rows.Next() {
		var routine Routine
		routine.records = []Record{}
		kcore.Expect(rows.Scan(&routine.ID, &routine.Title, &routine.FrequencyWeeks), "failed to scan row")
		if records, ok := recordsByRoutine[routine.ID]; ok {
			routine.records = records
		}
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
	Name     string
	Routines []Routine
}

func (room Room) DoneCount() int {
	doneCount := 0
	for _, routine := range room.Routines {
		if !routine.DueThisWeek() {
			doneCount++
		}
	}
	return doneCount
}

func (room Room) DueCount() int {
	return len(room.Routines) - room.DoneCount()
}

func (room Room) Title() string {
	doneCount := room.DoneCount()
	title := fmt.Sprintf("%s (%d/%d)", room.Name, doneCount, len(room.Routines))
	if doneCount == len(room.Routines) {
		title = "✅ " + title
	} else {
		title = "⏳ " + title
	}

	return title
}

func routineCmp(a, b Routine) int {
	if b.DueThisWeek() && !a.DueThisWeek() {
		return 1
	} else if !b.DueThisWeek() && a.DueThisWeek() {
		return -1
	}
	return int(a.LastRecordedAt().Sub(b.LastRecordedAt()).Nanoseconds())
}

func RoutinesRooms(db *sql.DB) []Room {
	roomsByName := map[string]Room{}
	routines := allRoutines(db)
	for _, routine := range routines {
		parts := strings.Split(routine.Title, "/")
		kcore.Assert(len(parts) == 2, "expected 2 parts")
		roomName := parts[0]
		room, ok := roomsByName[roomName]
		if !ok {
			room = Room{Name: roomName}
		}
		room.Routines = append(room.Routines, routine)
		roomsByName[roomName] = room
	}
	rooms := []Room{}
	for _, room := range roomsByName {
		slices.SortFunc(room.Routines, routineCmp)
		rooms = append(rooms, room)
	}
	slices.SortFunc(rooms, func(a, b Room) int { return a.DueCount() - b.DueCount() })
	return rooms
}
