package handler

import (
	"aarushishop/database"
	"aarushishop/globals"
	"aarushishop/helpers"
	"aarushishop/model"

	//	"context"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// LoginAPI handles user login
func LoginAPI() gin.HandlerFunc {
	return func(c *gin.Context) {

		var user model.LoginUser
		if err := c.ShouldBindJSON(&user); err != nil {
			log.Printf("Error on JSON Binding: %v", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"content": "Invalid JSON format"})
			return
		}

		// Check if username or password is empty
		log.Printf("Username :  %s -- Password: %s", user.Username, user.Password)
		if helpers.EmptyUserPass(user.Username, user.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"content": "Parameters can't be empty."})
			return
		}
		log.Printf("%s %s", user.Username, user.Password)
		// Check user credentials
		if !helpers.CheckUserPass(user.Username, user.Password) {
			// Use constant for status code
			c.JSON(http.StatusUnauthorized, gin.H{"content": "Incorrect username or password."})
			return
		}

		// Create a session for the authenticated user with custom options
		session := sessions.Default(c)
		session.Options(sessions.Options{
			Path:     "/",
			MaxAge:   900, // 15 minutes in seconds
			HttpOnly: true,
			Secure:   true, // Set to true if your application uses HTTPS
			SameSite: http.SameSiteStrictMode,
		})

		// Set the authenticated user in the session
		session.Set(globals.UserKey, user.Username)

		// Save the session (set the session cookie)
		if err := session.Save(); err != nil {
			// Log the error for debugging purposes
			log.Printf("Error saving session: %v", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Log successful login
		log.Printf("User %s logged in successfully", user.Username)

		// Optionally, you may send a success response
		c.JSON(http.StatusOK, gin.H{"content": "Login successful..."})
	}
}

// LogoutAPI handles user logout.
func LogoutAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the user from the session
		session := sessions.Default(c)
		user := session.Get(globals.UserKey)

		if user == nil {
			// Handle the case where the session is invalid or user is not logged in
			c.JSON(http.StatusUnauthorized, gin.H{"content": "Invalid logout request."})
			return
		}

		// Delete the user from the session
		session.Delete(globals.UserKey)

		// Save the session to remove the user's session cookie
		if err := session.Save(); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"content": "Logout successful..."})
	}
}

// Sample testing api
func TestAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Axios is working nicely",
		})
	}
}

// Purchase Entry API
func AddPurchaseAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// // Retrieve the user from the session
		// session := sessions.Default(c)
		// user := session.Get(globals.UserKey)

		// if user == nil {
		// 	// Handle the case where the session is invalid or user is not logged in
		// 	c.JSON(http.StatusUnauthorized, gin.H{"content": "Invalid logout request."})
		// 	return
		// }
		dbConn := database.GetDB()

		var Purchase model.Purchase
		//var Inventory model.Inventory

		if err := c.ShouldBindJSON(&Purchase); err != nil {
			log.Printf("Error on JSON Binding: %v", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error on JSON Binding"})
			return
		}

		// Start a database transaction
		tx, err := dbConn.Begin()
		if err != nil {
			log.Printf("Error starting transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to start a database transaction",
			})
			return
		}

		// Defer the rollback in case of an error
		defer func() {
			if err := tx.Rollback(); err != nil {
				log.Printf("Error rolling back transaction: %v", err)
			}
		}()

		// Check if the item exists in the inventory
		var existingQty int
		checkInventoryQuery := "SELECT qty FROM inventory WHERE item_list = ?"
		err = tx.QueryRow(checkInventoryQuery, Purchase.Item_List).Scan(&existingQty)

		if err == nil {
			// Item exists in the inventory, update the quantity
			updateInventoryQuery := "UPDATE inventory SET qty = qty + ? WHERE item_list = ?"
			_, err := tx.Exec(updateInventoryQuery, Purchase.Qty, Purchase.Item_List)
			if err != nil {
				log.Printf("Error updating inventory: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "Failed to update inventory",
				})
				return
			}
		} else {
			// Item doesn't exist in the inventory, insert a new record
			insertInventoryQuery := "INSERT INTO inventory (item_list, description, qty, category) VALUES (?, ?, ?, ?)"
			_, err := tx.Exec(insertInventoryQuery, Purchase.Item_List, Purchase.Description, Purchase.Qty, Purchase.Category)
			if err != nil {
				log.Printf("Error inserting data into inventory: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "Failed to save data to the database",
				})
				return
			}
		}

		// Execute an SQL to Insert Purchase Entry within the transaction
		insertPurchaseQuery := "INSERT INTO purchaseHistory (PO, item_list, description, qty, category, Price, User) VALUES (?, ?, ?, ?, ?, ?, ?)"
		_, err = tx.Exec(insertPurchaseQuery, Purchase.Po, Purchase.Item_List, Purchase.Description, Purchase.Qty, Purchase.Category, Purchase.Price, Purchase.User)
		if err != nil {
			log.Printf("Error inserting data into purchaseHistory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to save data to the database",
			})
			return
		}

		// Commit the transaction if both insert queries succeed
		if err := tx.Commit(); err != nil {
			log.Printf("Error committing transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to commit the transaction",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"PO":          Purchase.Po,
			"item_list":   Purchase.Item_List,
			"description": Purchase.Description,
			"qty":         Purchase.Qty,
			"category":    Purchase.Category,
			"Price":       Purchase.Price,
			"User":        Purchase.User,
		})
	}
}

// ListPurchaseAPI is a Gin handler function that retrieves a list of purchases from the database and returns them as JSON.
func ListPurchaseAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Uncomment the following code if user authentication is required
		// // Retrieve the user from the session
		// session := sessions.Default(c)
		// user := session.Get(globals.UserKey)

		// // Check if the user is not logged in
		// if user == nil {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"content": "Invalid logout request."})
		// 	return
		// }

		// Connect to the MySQL database
		dbConn := database.GetDB()

		// Perform the database query to fetch purchase history
		rows, err := dbConn.Query("SELECT PO, item_list, description, qty, category, Price, User FROM purchaseHistory ORDER BY PO")
		if err != nil {
			log.Printf("Error executing query: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from the database"})
			return
		}
		defer rows.Close()

		// Iterate through the rows and build a slice of Purchase structs
		var purchases []model.Purchase
		for rows.Next() {
			var purchase model.Purchase

			// Scan the database row into the Purchase struct
			if err := rows.Scan(&purchase.Po, &purchase.Item_List, &purchase.Description, &purchase.Qty, &purchase.Category, &purchase.Price, &purchase.User); err != nil {
				log.Printf("Error scanning row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from the database"})
				return
			}

			// Append the Purchase struct to the slice
			purchases = append(purchases, purchase)
		}

		// Check for errors during iteration
		if err := rows.Err(); err != nil {
			log.Printf("Error during iteration: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from the database"})
			return
		}

		// Return the fetched data as JSON
		c.JSON(http.StatusOK, purchases)
	}
}
