package dbrepo

import (
	"context"
	"database/sql"
	"fmt"
	"polling/internal/models"
	"strings"
	"time"
)

type DBRepo struct {
	DB *sql.DB
}

const dbTimeout = time.Second * 3

func (m *DBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *DBRepo) CreateUser(data models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		INSERT INTO users ( username, password, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := m.DB.ExecContext(ctx, query, data.Username, data.Password, data.FirstName, data.LastName, data.CreatedAt, data.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (m *DBRepo) GetUserByUsername(username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, username, password, first_name, last_name, created_at, updated_at
		FROM users 
		WHERE username = $1
	`

	var user models.User

	row := m.DB.QueryRowContext(ctx, query, username)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil

}

func (m *DBRepo) GetPollOptions(id int) ([]*models.PollOption, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var options []*models.PollOption

	query := `
		SELECT id, option_text
		FROM poll_options
		WHERE poll_id = $1
	`

	rows, err := m.DB.QueryContext(ctx, query, id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var opt models.PollOption
		err := rows.Scan(
			&opt.ID,
			&opt.Text,
		)

		if err != nil {
			return nil, err
		}

		options = append(options, &opt)

	}

	return options, nil
}

func (m *DBRepo) GetAllPolls() ([]*models.Poll, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var polls []*models.Poll

	query := `
		SELECT id, title, description
		FROM polls
	`

	rows, err := m.DB.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var poll models.Poll
		err := rows.Scan(
			&poll.ID,
			&poll.Title,
			&poll.Description,
		)

		if err != nil {
			return nil, err
		}

		options, err := m.GetPollOptions(poll.ID)

		if err != nil {
			return nil, err
		}

		poll.Options = options

		polls = append(polls, &poll)

	}

	return polls, nil

}

func (m *DBRepo) CreatePoll(data models.Poll) (*models.Poll, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		INSERT INTO polls (title, description, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, user_id`

	row := m.DB.QueryRowContext(ctx, query, data.Title, data.Description, data.UserID)

	var result models.Poll

	err := row.Scan(
		&result.ID,
		&result.Title,
		&result.Description,
		&result.UserID,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *DBRepo) AddPollOptions(pollId int, options []models.PollOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	query := `INSERT INTO poll_options (poll_id, option_text) VALUES `

	var args []any
	var placeholders []string

	for i, option := range options {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, pollId, option.Text)
	}

	query += strings.Join(placeholders, ", ")

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m *DBRepo) GetPollByID(id int) (*models.Poll, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, title, description
		FROM polls 
		WHERE id = $1
	`

	var poll models.Poll

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&poll.ID,
		&poll.Title,
		&poll.Description,
	)

	if err != nil {
		return nil, err
	}

	options, err := m.GetPollOptions(id)

	if err != nil {
		return nil, err
	}

	poll.Options = options

	return &poll, nil
}

func (m *DBRepo) UpdatePollByID(id int, data models.Poll) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		UPDATE polls
		SET title = $1, description = $2
		WHERE id = $3
	`

	_, err := m.DB.ExecContext(ctx, query, data.Title, data.Description, id)
	return err
}

func (m *DBRepo) DeletePollByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		DELETE FROM polls
		WHERE id = $1
	`

	_, err := m.DB.ExecContext(ctx, query, id)
	return err
}

func (m *DBRepo) UpdateOptionByID(id int, text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		UPDATE poll_options
		SET option_text = $1
		WHERE id = $2
	`

	_, err := m.DB.ExecContext(ctx, query, text, id)
	return err
}

func (m *DBRepo) DeleteOptionByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		DELETE FROM poll_options
		WHERE id = $1
	`

	_, err := m.DB.ExecContext(ctx, query, id)
	return err
}
