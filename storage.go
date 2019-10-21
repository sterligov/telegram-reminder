package main

import (
	"database/sql"
)

// type Storage interface {
// 	Save(*message) error
// 	Current() ([]message, error)
// 	Delete(int64) error
// 	Completed(int64) ([]message, error)
// 	Upcoming(int64) ([]message, error)
// 	Get(int64) (*message, error)
// 	Update(int64, string) error
// }

type PostgreStorage struct {
	*sql.DB
}

func (p *PostgreStorage) Current() ([]message, error) {
	rows, err := p.Query(`
		SELECT 
			id, 
			txt, 
			TO_CHAR(reminder_at, 'YYYY-MM-DD HH24:MI') AS reminder_at,
			chat_id
		FROM telegram.reminder
		WHERE reminder_at <= NOW()
			AND reminder_at > NOW() - '1 minute'::interval
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

func (p *PostgreStorage) Save(msg *message) error {
	_, err := p.Exec(`
		INSERT INTO telegram.reminder(txt, reminder_at, chat_id)
		VALUES($1, $2, $3)
	`, msg.Txt, msg.ReminderAt, msg.ChatID)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgreStorage) Delete(id int64) error {
	_, err := p.Exec(`
		DELETE
		FROM telegram.reminder
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgreStorage) Completed(id int64) ([]message, error) {
	rows, err := p.Query(`
		SELECT id,
			   txt,
			   TO_CHAR(reminder_at, 'YYYY-MM-DD HH24:MI') AS reminder_at,
			   chat_id
		FROM telegram.reminder
		WHERE reminder_at <= NOW() - '1 minute'::interval
			AND chat_id = $1
		ORDER BY reminder_at
		LIMIT 10
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

func (p *PostgreStorage) Upcoming(id int64) ([]message, error) {
	rows, err := p.Query(`
		SELECT id,
			   txt,
			   TO_CHAR(reminder_at, 'YYYY-MM-DD HH24:MI') AS reminder_at,
			   chat_id
		FROM telegram.reminder
		WHERE reminder_at > NOW()
			AND chat_id = $1
		ORDER BY reminder_at
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

func (p *PostgreStorage) Get(id int64) (*message, error) {
	row := p.QueryRow(`
		SELECT id,
               txt,
	           TO_CHAR(reminder_at, 'YYYY-MM-DD HH24:MI') AS reminder_at,
	           chat_id
		FROM telegram.reminder
		WHERE id = $1
	`, id)

	m := &message{}
	err := row.Scan(&m.ID, &m.Txt, &m.ReminderAt, &m.ChatID)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (p *PostgreStorage) Update(id int64, date string) error {
	_, err := p.Exec(`
		UPDATE telegram.reminder
		SET reminder_at = $1
		WHERE id = $2
	`, date, id)

	return err
}

func scanMessages(rows *sql.Rows) ([]message, error) {
	messages := make([]message, 0)
	for rows.Next() {
		m := message{}
		err := rows.Scan(&m.ID, &m.Txt, &m.ReminderAt, &m.ChatID)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}
