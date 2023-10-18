// * model/model.go
package model

// User represents a user in the database
type User struct {
	UserID       int    `json:"UserID"`        // user_id integer
	Username     string `json:"Username"`      // username character varying(20)
	Email        string `json:"Email"`         // email character varying
	PasswordHash string `json:"PasswordHash"`  // password_hash character varying
	Comment      string `json:"Comment"`       // comment character varying
}