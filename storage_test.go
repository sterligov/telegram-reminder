package main

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func prepareStorage() (*PostgreStorage, error) {
	p := &PostgreStorage{initDatabase()}
	rows, err := p.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'telegram'
	`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var table string
		err := rows.Scan(&table)
		if err != nil {
			return nil, err
		}
		_, err = p.Exec(`
			TRUNCATE telegram.` + table + ` RESTART IDENTITY CASCADE;
			ALTER SEQUENCE telegram.seq_` + table + ` RESTART WITH 1;
		`)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}
func TestUpcoming(t *testing.T) {
	stor, err := prepareStorage()
	if err != nil {
		t.Fatal(err)
	}

	upcomingTestMessages := []message{
		message{
			Txt:        "Upcoming 1",
			ReminderAt: "2098-01-03 15:33",
			ChatID:     1,
		},
		message{
			Txt:        "Upcoming 2",
			ReminderAt: "2099-07-03 12:00",
			ChatID:     1,
		},
	}

	testMessages := []message{
		message{
			Txt:        "Comleted 1",
			ReminderAt: "2000-07-03 14:28",
			ChatID:     1,
		},
		message{
			Txt:        "Upcoming 3",
			ReminderAt: "2099-07-03 12:00",
			ChatID:     2,
		},
	}

	testMessages = append(testMessages, upcomingTestMessages...)
	for k, m := range testMessages {
		err = stor.Save(&m)
		if err != nil {
			t.Errorf("#%d: %s", k, err.Error())
			return
		}
	}

	upcomings, err := stor.Upcoming(1)
	if err != nil {
		t.Error(err)
		return
	}

	if len(upcomings) != len(upcomingTestMessages) {
		t.Errorf("Expected %d upcoming messages, got %d", len(upcomingTestMessages), len(upcomings))
		return
	}

	for k, v := range upcomingTestMessages {
		if cmp.Equal(v, upcomings[k]) {
			t.Errorf("#%d: Expected %v, got %v", k, v, upcomings[k])
		}
	}
}

func TestCompleted(t *testing.T) {
	stor, err := prepareStorage()
	if err != nil {
		t.Fatal(err)
	}

	testMessages := []message{
		message{
			Txt:        "Upcoming 1",
			ReminderAt: "2098-01-03 15:33",
			ChatID:     1,
		},
		message{
			Txt:        "Upcoming 2",
			ReminderAt: "2099-07-03 12:00",
			ChatID:     1,
		},
		message{
			Txt:        "Completed 3",
			ReminderAt: "1995-05-01 21:56",
			ChatID:     2,
		},
	}

	completedTestMessages := []message{
		message{
			Txt:        "Comleted 1",
			ReminderAt: "2000-07-03 14:28",
			ChatID:     1,
		},
		message{
			Txt:        "Completed 2",
			ReminderAt: "1999-12-30 23:28",
			ChatID:     1,
		},
	}

	testMessages = append(testMessages, completedTestMessages...)
	for k, m := range testMessages {
		err = stor.Save(&m)
		if err != nil {
			t.Errorf("#%d: %s", k, err.Error())
			return
		}
	}

	completed, err := stor.Completed(1)
	if err != nil {
		t.Error(err)
		return
	}

	if len(completed) != len(completedTestMessages) {
		t.Errorf("Expected %d completed messages, got %d", len(completedTestMessages), len(completed))
		return
	}

	for k, v := range completedTestMessages {
		if cmp.Equal(v, completed[k]) {
			t.Errorf("#%d: Expected %v, got %v", k, v, completed[k])
		}
	}
}

func TestCurrent(t *testing.T) {
	stor, err := prepareStorage()
	if err != nil {
		t.Fatal(err)
	}

	testMessages := []message{
		message{
			Txt:        "Upcoming 1",
			ReminderAt: "2098-01-03 15:33",
			ChatID:     1,
		},
		message{
			Txt:        "Upcoming 2",
			ReminderAt: "2099-07-03 12:00",
			ChatID:     1,
		},
		message{
			Txt:        "Upcoming 3",
			ReminderAt: "2099-07-03 12:00",
			ChatID:     2,
		},
		message{
			Txt:        "Upcoming 3",
			ReminderAt: "2000-07-03 12:00",
			ChatID:     2,
		},
	}

	currentTestMessages := []message{
		message{
			Txt:        "Comleted 1",
			ReminderAt: time.Now().Format(timeLayout),
			ChatID:     1,
		},
		message{
			Txt:        "Completed 2",
			ReminderAt: time.Now().Format(timeLayout),
			ChatID:     1,
		},
	}

	testMessages = append(testMessages, currentTestMessages...)
	for k, m := range testMessages {
		err = stor.Save(&m)
		if err != nil {
			t.Errorf("#%d: %s", k, err.Error())
			return
		}
	}

	current, err := stor.Current()
	if err != nil {
		t.Error(err)
		return
	}

	if len(current) != len(currentTestMessages) {
		t.Errorf("Expected %d completed messages, got %d", len(currentTestMessages), len(current))
		return
	}

	for k, v := range currentTestMessages {
		if cmp.Equal(v, current[k]) {
			t.Errorf("#%d: Expected %v, got %v", k, v, current[k])
		}
	}
}

func TestGet(t *testing.T) {
	stor, err := prepareStorage()
	if err != nil {
		t.Fatal(err)
	}
	testMessage := &message{
		ID:         1,
		Txt:        "Comleted 1",
		ReminderAt: time.Now().Format(timeLayout),
		ChatID:     1,
	}

	err = stor.Save(testMessage)
	if err != nil {
		t.Error(err)
		return
	}

	m, err := stor.Get(1)
	if err != nil {
		t.Error(err)
		return
	}
	if !cmp.Equal(m, testMessage) {
		t.Errorf("Expected %v, got %v", testMessage, m)
	}
}

func TestDelete(t *testing.T) {
	stor, err := prepareStorage()
	if err != nil {
		t.Fatal(err)
	}
	testMessage := &message{
		Txt:        "Comleted 1",
		ReminderAt: time.Now().Format(timeLayout),
		ChatID:     1,
	}

	err = stor.Save(testMessage)
	if err != nil {
		t.Error(err)
		return
	}

	err = stor.Delete(1)
	if err != nil {
		t.Error(err)
		return
	}

	m, err := stor.Get(1)
	if err == nil || m != nil {
		t.Error("Expected ErrNoRows")
	}
}

func TestUpdate(t *testing.T) {
	stor, err := prepareStorage()
	if err != nil {
		t.Fatal(err)
	}
	testMessage := &message{
		Txt:        "Comleted 1",
		ReminderAt: time.Now().Format(timeLayout),
		ChatID:     1,
	}

	err = stor.Save(testMessage)
	if err != nil {
		t.Error(err)
		return
	}

	newDate := "2099-12-31 23:59"
	err = stor.Update(1, newDate)
	if err != nil {
		t.Error(err)
		return
	}

	m, err := stor.Get(1)
	if err != nil {
		t.Error(err)
	}
	if m.ReminderAt != newDate {
		t.Errorf("Expected date %s, got %s", newDate, m.ReminderAt)
	}
}
