package datab

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	connStr := "user=vivacity password=vivacityOrg dbname=vivacity_website sslmode=disable host=localhost port=5432"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	fmt.Println("Successfully connected to PostgreSQL: Vivacity_website!")
	return db, nil
}

// SetupDB initializes the database and creates tables
func SetupDB(db *sql.DB) error {
	// Create teams table
	createTeamsTable := `
	CREATE TABLE IF NOT EXISTS teams (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createTeamsTable)
	if err != nil {
		return fmt.Errorf("error creating teams table: %v", err)
	}

	// Make sure the teams table is committed before creating users table
	// Create users table with a foreign key reference to teams
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) NOT NULL,
		user_id INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createUsersTable)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	// Create team_members table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS team_members (
        user_id INT NOT NULL,
        team_id INT NOT NULL,
        role VARCHAR(255) NOT NULL,
        PRIMARY KEY (user_id, team_id),
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
    );
	`)
	if err != nil {
		return fmt.Errorf("error creating team_members table: %v", err)
	}

	// Create time_slots table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS time_slots (
			slot_id SERIAL PRIMARY KEY,
			weekday VARCHAR(9) NOT NULL,
			time TIME NOT NULL,
			UNIQUE (weekday, time)
		);
	`)
	if err != nil {
		return fmt.Errorf("error creating time_slots table: %v", err)
	}

	// Create availability table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS availability (
			user_id INT NOT NULL,
			team_id INT NOT NULL,
			slot_id INT NOT NULL,
			available BOOLEAN NOT NULL DEFAULT FALSE,
			PRIMARY KEY (user_id, slot_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (slot_id) REFERENCES time_slots(slot_id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		return fmt.Errorf("error creating availability table: %v", err)
	}

	// Populate time_slots
	_, err = db.Exec(`
		INSERT INTO time_slots (weekday, time)
		VALUES
			('Monday', '19:00:00'), ('Monday', '21:00:00'),
			('Tuesday', '19:00:00'), ('Tuesday', '21:00:00'),
			('Wednesday', '19:00:00'), ('Wednesday', '21:00:00'),
			('Thursday', '19:00:00'), ('Thursday', '21:00:00'),
			('Friday', '19:00:00'), ('Friday', '21:00:00'),
			('Saturday', '19:00:00'), ('Saturday', '21:00:00'),
			('Sunday', '19:00:00'), ('Sunday', '21:00:00')
		ON CONFLICT (weekday, time) DO NOTHING;
	`)
	if err != nil {
		return fmt.Errorf("error populating time_slots: %v", err)
	}

	return nil
}
