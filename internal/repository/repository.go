package repository

import (
	"database/sql"
	"polling/internal/models"
)

type Repository interface {
	Connection() *sql.DB
	CreateUser(data models.User) error
	GetUserByUsername(username string) (*models.User, error)
	CreatePoll(data models.Poll) (*models.Poll, error)
	GetAllPolls() ([]*models.Poll, error)
	GetPollOptions(id int) ([]*models.PollOption, error)
	AddPollOptions(pollId int, options []models.PollOption) error
	GetPollByID(id int) (*models.Poll, error)
	UpdatePollByID(id int, data models.Poll) error
	DeletePollByID(id int) error
	UpdateOptionByID(id int, text string) error
	DeleteOptionByID(id int) error
}
