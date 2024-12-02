package models

import (
	"database/sql"
	"time"
)

// Snippet Define a snippet to hold the data for an individual.
// Notice how the fields of the struct correspond to the fields of the struct correspond to the fields in our MySQL snippets
// table?
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// SnippetModel Define a SnippetModel type which wraps a sql.DB connection pool.
// This will also include the below methods to interact with the data.
type SnippetModel struct {
	DB *sql.DB
}

// Insert This will insert a new snippet into the database.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	// Writes the SQL statement we want to execute.
	// The placeholder parameter syntax differs depending on your database. MySQL, SQL server and SQLite use the ? notation
	// But the PostgresSQL uses the $N notation. Example: INSERT INTO ... VALUES($1, $2, $3...)
	stmt := `INSERT INTO snippets (title, content, created, expires) VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// Use the Exec() method on the embedded connection pool to execute the statement.
	// The first parameter is the SQL statement, followed by the method returns a sql.Result type, which contains some basic
	// information about what happened when the statement was executed.
	// Behind the scenes, the DB.Exec() method works in three steps:
	// - It creates a new prepared statement on the database using the provided SQL statement.
	// - Exec() passes the parameter values to the database. The database then executes the prepared statement.
	// - It then closes (or deallocates) the prepared statement on the database.
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// Use the LastInsertId() method on the result to get the ID of our newly inserted record in the snippets table.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// The ID returned has the type int64, so we convert it to an int type before returning
	return int(id), nil
}

// Get This will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (*Snippet, error) {
	return nil, nil
}

// Latest This will return the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]*Snippet, error) {
	return nil, nil
}
