package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KhrisKringle/Vivacity_website-main/server/api"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestTeamHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Test POST
	teamName := "New Team"
	body, _ := json.Marshal(map[string]string{"name": teamName})

	// 1. Expect the INSERT query
	mock.ExpectExec("INSERT INTO teams").
		WithArgs(teamName).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 2. ADD THIS: Expect the SELECT query to get the new ID
	mock.ExpectQuery("SELECT id FROM teams WHERE name = \\$1").
		WithArgs(teamName).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	req, err := http.NewRequest(http.MethodPost, "/api/teams", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.TeamHandler(db))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}
