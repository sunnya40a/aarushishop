package model

// User represents a user in the database
type User struct {
	UserID       int    // user_id integer
	Username     string // username character varying(20)
	Email        string // email character varying
	PasswordHash string // password_hash character varying
	Comment      string // comment character varying
}