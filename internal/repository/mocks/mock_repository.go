package mocks

import (
	"database/sql"
	"errors"
	"polling/internal/models"
)

// MockDBRepo implements the repository.Repository interface for testing
type MockDBRepo struct {
	ShouldFail bool
	MockUser   *models.User
	MockError  error
}

func (m *MockDBRepo) Connection() *sql.DB {
	return nil
}

func (m *MockDBRepo) CreateUser(data models.User) error {
	if m.ShouldFail {
		return errors.New("database error")
	}
	if data.Username == "existing_user" {
		return errors.New("user already exists")
	}
	return nil
}

func (m *MockDBRepo) GetUserByUsername(username string) (*models.User, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockUser, nil
}

func (m *MockDBRepo) CreatePoll(data models.Poll) (*models.Poll, error) {
	if m.ShouldFail {
		return nil, errors.New("database error")
	}
	// Return the poll with an ID to simulate successful creation
	data.ID = 1
	return &data, nil
}

func (m *MockDBRepo) GetAllPolls() ([]*models.Poll, error) {
	return nil, nil
}

func (m *MockDBRepo) GetPollByID(id int) (*models.Poll, error) {
	return nil, nil
}

func (m *MockDBRepo) GetPollOptions(id int) ([]*models.PollOption, error) {
	return nil, nil
}

func (m *MockDBRepo) UpdatePollByID(id int, data models.Poll) error {
	return nil
}

func (m *MockDBRepo) DeletePollByID(id int) error {
	if id == 1 {
		return nil
	}

	return errors.New("database error")
}

func (m *MockDBRepo) AddPollOptions(pollId int, options []models.PollOption) error {
	if pollId == 2 {
		return errors.New("database error")
	}
	return nil
}

func (m *MockDBRepo) UpdateOptionByID(id int, text string) error {
	return nil
}

func (m *MockDBRepo) DeleteOptionByID(id int) error {
	return nil
}

func (m *MockDBRepo) Vote(pollID int, optionID int, userID int) error {
	return nil
}

func (m *MockDBRepo) Unvote(optionID int, userID int) error {
	return nil
}

func (m *MockDBRepo) GetOptionVotes(optionID int) ([]*models.Vote, error) {
	return nil, nil
}

func (m *MockDBRepo) IsPollOwner(pollID int, userID int) bool {
	if pollID == 1 && userID == 1 || pollID == 2 && userID == 1 {
		return true
	}
	return false
}
