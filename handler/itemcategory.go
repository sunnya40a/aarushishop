package handler

import (
	"aarushishop/database"
	//	"aarushishop/globals"
	"aarushishop/model"
	"context"
	"net/http"

	//	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ListCategoryAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		//session := sessions.Default(c)
		//user := session.Get(globals.UserKey)

		//if user == nil {
		//	c.HTML(http.StatusUnauthorized, LoginTemplate, gin.H{"content": "User not found in session."})
		//	return
		//}

		// Connect to the MySQL database
		dbConn:=database.GetDB()
		// Execute the SQL query to fetch data from the "users" table
		rows, err := dbConn.QueryContext(context.Background(), "select category_code, description  from category_list")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch data from the database"})
			return
		}
		defer rows.Close()

		// Create a slice to store the category data
		var categories []model.Category // Replace "model.Category" with the struct type that matches your category data

		// Iterate through the query results and append them to the slice
		for rows.Next() {
			var category model.Category // Replace "model.User" with the struct type that matches your user data
			if err := rows.Scan(&category.CategoryCode, &category.Description); err != nil {
				c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"message": "Error scanning user data"})
				return
			}
			categories = append(categories, category)
		}

		c.JSON(http.StatusOK, gin.H{
			"category": categories,
		})
	} 
}
