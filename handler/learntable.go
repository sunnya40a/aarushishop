//* learntable.go

package handler

import (
	"aarushishop/database"
	"aarushishop/globals"
	"context"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LUser struct {
	ClientID int    `json:"client_id" binding:"required"`
	UName    string `json:"uname" binding:"required"`
}

func LearnTableGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "User not found in session."})
			return
		}
		//fmt.Printf("\n\nUser Data: %+v\n\n", users)
		c.HTML(http.StatusOK, "vuelearntable.tmpl", gin.H{}) 
	}
}
func LearnEntryGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "User not found in session."})
			return
		}
		//fmt.Printf("\n\nUser Data: %+v\n\n", users)
		c.HTML(http.StatusOK, "entrytable.tmpl", gin.H{}) 
	}
}
func ListUserAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in session."})
			return
		} 

		// Connect to the database (assuming you've set up the DB connection)
		dbConn := database.GetDB()


		// Execute the SQL query to fetch data from the "learn" table
		rows, err := dbConn.QueryContext(context.Background(), "SELECT client_id, uname FROM learn")
		if err != nil {
			log.Printf("Failed to fetch data from the database: %v", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from the database"})
			return
		}
		defer rows.Close()

		var users []LUser
		for rows.Next() {
			var user LUser
			if err := rows.Scan(&user.ClientID, &user.UName); err != nil {
				log.Printf("Error scanning user data: %v", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning user data"})
				return
			}
			users = append(users, user)
		}
		c.JSON(http.StatusOK, gin.H{"user": users})
	}
}

func CreateUserAPI() gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        user := session.Get(globals.UserKey)

        if user == nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in session."})
            return
        }

        dbConn := database.GetDB()
    
		// Bind the JSON data from the request body to the newUser struct
        var newUser LUser
        if err := c.ShouldBindJSON(&newUser); err != nil {
			log.Printf("Error on JSON Binding: %v", err.Error())
            c.JSON(http.StatusBadRequest, gin.H{"error": "Error on JSON Binding"})
            return
        }
		// Execute an SQL INSERT statement to save the new user data
        _, err := dbConn.ExecContext(context.Background(), "INSERT INTO learn (client_id, uname) VALUES (?, ?)", newUser.ClientID, newUser.UName)
        if err != nil {
			log.Printf("Error inserting data: %v", err)
		
            c.JSON(http.StatusInternalServerError, gin.H{
                "status":  "error",
                "message": "Failed to save data to the database",
            })
            return
        }

		c.JSON(http.StatusOK, gin.H{
			"client_id": newUser.ClientID,
			"uname": newUser.UName,
		})
    }
}

func GetuserbyIDAPI()gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Print("Get User by ID")
	}
}

func DeleteUserAPI()gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in session."})
		return
		}
		dbConn := database.GetDB()
		
		// Execute an SQL INSERT statement to save the new user data
		_, err := dbConn.ExecContext(context.Background(), "DELETE FROM learn WHERE client_id = ?", c.Param("client_id"))
		if err != nil {
			log.Printf("Error inserting data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save data to the database",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "Deleted Susessful",
			"client_id": c.Param("client_id"),
		})		
	}
}

func EditUserAPI()gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in session."})
		return
		}

		// Get a database connection.
		dbConn := database.GetDB()
	

		// Bind the JSON data from the request body to the User struct.
		var cuser LUser   // we define here cuser and user is already used for our session.
		if err := c.ShouldBindJSON(&cuser); err != nil {
			log.Printf("Error updating data: %v", err.Error())
			rawBody, _ := c.GetRawData()
			log.Printf("Raw Request Body: %s\n", string(rawBody))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Check if the user exists with the specified client_id.
		var count int
		row := dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM learn WHERE client_id = ?", cuser.ClientID) 
        err := row.Scan(&count)


		if err != nil {
			log.Printf("Error checking user existence: %v", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
			return
		}
		if count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found to update"})
			return
		}

		// Execute an SQL UPDATE statement to modify the user's name.
		_, err = dbConn.ExecContext(context.Background(), "UPDATE learn SET uname = ? WHERE client_id = ?", cuser.UName, cuser.ClientID)
        if err != nil {
            log.Printf("Error updating data: %v", err.Error())
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user data in the database"})
            return
        }

		c.JSON(http.StatusOK, gin.H{
			"status":    "Updated Successfully",
			"client_id": cuser.ClientID,
		})
	}
}
