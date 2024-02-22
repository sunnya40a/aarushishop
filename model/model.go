// * model/model.go
package model

import "github.com/golang-jwt/jwt/v5"

// This is for login
type LoginUser struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

// User represents a user in the database
type User struct {
	UserID       int    `json:"UserID"`       // user_id integer
	Username     string `json:"Username"`     // username character varying(20)
	Email        string `json:"Email"`        // email character varying
	PasswordHash string `json:"PasswordHash"` // password_hash character varying
	Comment      string `json:"Comment"`      // comment character varying
}

// This is for Material Category
type Category struct {
	CategoryCode string `json:"CategoryCode"`
	Description  string `json:"Description"`
}

// This is for Inventory Material
type Inventory struct {
	Item_List   string `json:"Item_List"`
	Description string `json:"Description"`
	Qty         int    `json:"Qty"`
	Category    string `json:"Category"`
}

// This is for Purchase.
type Purchase struct {
	Po          int     `json:"PO"`
	Pdate       string  `json:"Pdate"`
	Item_List   string  `json:"Item_list"`
	Description string  `json:"Description"`
	Qty         int     `json:"Qty"`
	Category    string  `json:"Category"`
	Price       float64 `json:"Price"`
	User        string  `json:"User"`
}

type TestUser struct {
	ClientID int    `json:"client_id"`
	Uname    string `json:"Uname"`
	DOB      string `json:"DOB"`
}

// CustomClaims defines the custom claims for the JWT token.
type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
}
