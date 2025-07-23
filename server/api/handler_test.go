package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Vivacity_website/server/api"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestTeamHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Mock data for teams
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Valiant").
		AddRow(2, "Voidwalkers")

	// Mock the query for getting all teams
	mock.ExpectQuery("SELECT id, name FROM teams").WillReturnRows(rows)

	// Create a new request to get all teams
	req, err := http.NewRequest("GET", "/api/teams", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(api.TeamHandler(db))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	expected := `[{"id":1,"name":"Valiant"},{"id":2,"name":"Voidwalkers"}]`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// Test creating a new team
	newTeam := map[string]string{"name": "Esoteric"}
	body, _ := json.Marshal(newTeam)

	mock.ExpectExec("INSERT INTO teams").
		WithArgs("Esoteric").
		WillReturnResult(sqlmock.NewResult(3, 1))

	req, err = http.NewRequest("POST", "/api/teams", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
}
