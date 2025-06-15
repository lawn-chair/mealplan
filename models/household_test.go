package models

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	clerk "github.com/clerk/clerk-sdk-go/v2"
	"github.com/jmoiron/sqlx"
)

func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock db: %v", err)
	}
	sdb := sqlx.NewDb(db, "postgres")
	return sdb, mock
}

func mockClerkUser(id, email, lastName string) *clerk.User {
	primaryEmailID := "email_1"
	return &clerk.User{
		ID:                    id,
		LastName:              &lastName,
		PrimaryEmailAddressID: &primaryEmailID,
		EmailAddresses: []*clerk.EmailAddress{{
			ID:           primaryEmailID,
			EmailAddress: email,
		}},
	}
}

func TestGenerateJoinCode(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()
	mock.ExpectExec("INSERT INTO household_join_codes").WillReturnResult(sqlmock.NewResult(1, 1))
	code, err := GenerateJoinCode(db, 1, time.Hour)
	if err != nil || len(code) != 8 {
		t.Errorf("unexpected error or code length: %v, %s", err, code)
	}
}

func TestJoinHouseholdByCode(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()
	user := mockClerkUser("user1", "user1@email.com", "Smith")
	mock.ExpectQuery("SELECT household_id FROM household_join_codes").WithArgs("CODE1234").WillReturnRows(sqlmock.NewRows([]string{"household_id"}).AddRow(2))
	mock.ExpectExec("DELETE FROM household_members WHERE user_id").WithArgs(user.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO household_members").WillReturnResult(sqlmock.NewResult(1, 1))
	err := JoinHouseholdByCode(db, user, "CODE1234")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLeaveHousehold(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()
	user := mockClerkUser("user1", "user1@email.com", "Smith")
	mock.ExpectExec("DELETE FROM household_members WHERE user_id").WithArgs(user.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("INSERT INTO households").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec("INSERT INTO household_members").WillReturnResult(sqlmock.NewResult(1, 1))
	err := LeaveHousehold(db, user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestListHouseholdMembers(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()
	mock.ExpectQuery("SELECT user_id FROM household_members").WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user1").AddRow("user2"))
	members, err := ListHouseholdMembers(db, 1)
	if err != nil || len(members) != 2 {
		t.Errorf("unexpected error or wrong member count: %v, %v", err, members)
	}
	if members[0] != "user1" || members[1] != "user2" {
		t.Errorf("unexpected member values: %v", members)
	}
}
