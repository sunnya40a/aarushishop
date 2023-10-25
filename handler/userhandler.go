//* userhandler.go

package handler

import (
	"aarushishop/database"
	"aarushishop/globals"
	"aarushishop/model"
	"context"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func TableGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "User not found in session."})
			return
		}

		// Connect to the database (assuming you've set up the DB connection)
		dbConn, err := database.GetDBConnection()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Database connection error"})
			return
		}
		defer dbConn.Release()

		// Execute the SQL query to fetch data from the "users" table
		rows, err := dbConn.Query(context.Background(), "SELECT user_id, username, email, password_hash, comment FROM users")
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Failed to fetch data from the database"})
			return
		}
		defer rows.Close()
		// Create a slice to store the user data
		var users []model.User // Replace "model.User" with the struct type that matches your user data
		// Iterate through the query results and append them to the slice
		for rows.Next() {
			var user model.User // Replace "model.User" with the struct type that matches your user data
			if err := rows.Scan(&user.UserID, &user.Username, &user.Email, &user.PasswordHash, &user.Comment); err != nil {
				c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error scanning user data"})
				return
			}
			users = append(users, user)
		}

		c.HTML(http.StatusOK, "templateusers.tmpl", gin.H{
			"user": users,
		})
	}
}

func ListUserGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "User not found in session."})
			return
		}

	 		//fmt.Printf("\n\nUser Data: %+v\n\n", users)
		c.HTML(http.StatusOK, "vueuserlist.tmpl", gin.H{}) 
	}
}

func ListMyUserAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"content": "User not found in session."})
			return
		}
		// Connect to the database (assuming you've set up the DB connection)
		dbConn, err := database.GetDBConnection()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database connection error"})
			return
		}
		defer dbConn.Release()
		// Execute the SQL query to fetch data from the "users" table
		rows, err := dbConn.Query(context.Background(), "SELECT user_id, username, email, password_hash, comment FROM users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch data from the database"})
			return
		}
		defer rows.Close()
		// Create a slice to store the user data
		var users []model.User // Replace "model.User" with the struct type that matches your user data
		// Iterate through the query results and append them to the slice
		for rows.Next() {
			var user model.User // Replace "model.User" with the struct type that matches your user data
			if err := rows.Scan(&user.UserID, &user.Username, &user.Email, &user.PasswordHash, &user.Comment); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Error scanning user data"})
				return
			}
			users = append(users, user)
		}
		c.JSON(http.StatusOK, gin.H{"user": users})
	}
}