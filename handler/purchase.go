package handler

import (
	"aarushishop/database"
	"aarushishop/helpers"
	"aarushishop/model"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mrz1836/go-sanitize"
)

func CustomDateValidator(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// AddPurchaseAPI is a Gin handler function that handles the API endpoint for adding a purchase
func AddPurchaseAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve database connection
		dbConn := database.GetDB()

		// Define struct for purchase data with validation tags
		type AddPurchase struct {
			Po          int     `json:"PO" validate:"required, min=20000101001,max=20001231999"` // Purchase order number
			Pdate       string  `json:"Pdate" validate:"required,date"`                          // Purchase date
			Item_List   string  `json:"item_list" validate:"required,len=9,max=9"`               // Item list
			Description string  `json:"description" validate:"required,len=1,max=255"`           // Description of the purchase
			Qty         int     `json:"qty" validate:"required, min=1,max=999"`                  // Quantity
			Category    string  `json:"category" validate:"required,len=1,max=255"`              // Category of the purchase
			Price       float64 `json:"Price" validate:"required, min=0,max=99999.99"`           // Price
			User        string  `json:"User" validate:"required,len=8,max=15"`                   // User who made the purchase
		}

		// Initialize the validator
		validate := validator.New()
		// Register custom date validation function
		validate.RegisterValidation("date", CustomDateValidator)

		var Purchase AddPurchase

		// Bind JSON data to the AddPurchase struct
		if err := c.ShouldBindJSON(&Purchase); err != nil {
			log.Printf("Error on JSON Binding: %v", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error on JSON Binding"})
			return
		}

		// Validate the Purchase struct
		if err := validate.Struct(Purchase); err != nil {
			log.Printf("Validation error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Parse the string date into a time.Time variable
		parsedDate, err := time.Parse("2006-01-02", Purchase.Pdate)
		if err != nil {
			log.Printf("Error parsing date: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing date. Date format should be 'yyyy-mm-dd'"})
			return
		}

		// Sanitize user input to prevent XSS attacks
		Purchase.Item_List = sanitize.HTML(helpers.SanitizeData(Purchase.Item_List).(string))
		Purchase.Description = sanitize.HTML(helpers.SanitizeData(Purchase.Description).(string))
		Purchase.Category = sanitize.HTML(helpers.SanitizeData(Purchase.Category).(string))
		Purchase.User = sanitize.HTML(helpers.SanitizeData(Purchase.User).(string))
		Purchase.Pdate = sanitize.HTML(helpers.SanitizeData(Purchase.Pdate).(string))

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
		checkInventoryQuery := squirrel.Select("qty").From("inventory").Where(squirrel.Eq{"item_list": Purchase.Item_List}).Limit(1).RunWith(tx)
		err = checkInventoryQuery.Scan(&existingQty)

		if err == nil {
			// Item exists in the inventory, update the quantity
			updateInventoryQuery := squirrel.Update("inventory").Set("qty", squirrel.Expr("qty + ?", Purchase.Qty)).Where(squirrel.Eq{"item_list": Purchase.Item_List}).RunWith(tx)
			_, err := updateInventoryQuery.Exec()
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
			insertInventoryQuery := squirrel.Insert("inventory").Columns("item_list", "description", "qty", "category").Values(Purchase.Item_List, Purchase.Description, Purchase.Qty, Purchase.Category).RunWith(tx)
			_, err := insertInventoryQuery.Exec()
			if err != nil {
				log.Printf("Error inserting data into inventory: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "Failed to save data to the database",
				})
				return
			}
		}

		// Now use parsedDate to format the date as needed
		formattedDate := parsedDate.Format("2006-01-02") // Format the date as needed

		// Execute an SQL to Insert Purchase Entry within the transaction
		insertPurchaseQuery := squirrel.Insert("purchaseHistory").
			Columns("PO", "Pdate", "item_list", "description", "qty", "category", "Price", "User").
			Values(Purchase.Po, formattedDate, Purchase.Item_List, Purchase.Description, Purchase.Qty, Purchase.Category, Purchase.Price, Purchase.User).
			RunWith(tx)
		_, err = insertPurchaseQuery.Exec()
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

		// Respond with JSON containing the added purchase details
		c.JSON(http.StatusOK, gin.H{
			"PO":          Purchase.Po,
			"Pdate":       Purchase.Pdate,
			"item_list":   Purchase.Item_List,
			"description": Purchase.Description,
			"qty":         Purchase.Qty,
			"category":    Purchase.Category,
			"Price":       Purchase.Price,
			"User":        Purchase.User,
		})
	}
}

// beloc PurchasePaginationAPI API is for pagination and for sample use.
func PurchasePaginationAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		dbConn := database.GetDB()
		pageStr := c.DefaultQuery("page", "0")
		limitStr := c.DefaultQuery("limit", "0")
		sortBy := c.DefaultQuery("sortBy", "PO")
		sortOrder := c.DefaultQuery("sortOrder", "asc")
		searchTerm := c.Query("search")
		datef := c.Query("datef")
		datee := c.Query("datee")
		datecriteria := ""
		searchcriteria := ""
		sqlbasestring := "SELECT PO, Pdate, item_list, description, qty, category, Price, User FROM purchaseHistory"

		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			if page != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page parameter"})
				return
			}
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			if limit != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid limit parameter (should be between 1 and 100)"})
				return
			}
		}

		offset := (page - 1) * limit

		var sortOrderSQL string
		if sortOrder == "desc" {
			sortOrderSQL = "DESC"
		} else {
			sortOrderSQL = "ASC"
		}

		if datef != "" && datee != "" {
			datecriteria = fmt.Sprintf(" WHERE (Pdate BETWEEN '%s' AND '%s')", datef, datee)
		}
		if searchTerm != "" {
			searchcriteria = fmt.Sprintf("(PO LIKE '%%%s%%' OR Pdate LIKE '%%%s%%' OR item_list LIKE '%%%s%%' OR description LIKE '%%%s%%' OR category LIKE '%%%s%%' OR Price LIKE '%%%s%%' OR User LIKE '%%%s%%')", searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
			if datecriteria != "" {
				searchcriteria = fmt.Sprintf(" AND %s", searchcriteria)
			} else {
				searchcriteria = fmt.Sprintf(" WHERE %s", searchcriteria)
			}
		}
		sqlstring := ""
		if page > 0 && limit > 0 {
			sqlstring = fmt.Sprintf("%s%s%s ORDER BY %s %s LIMIT %d OFFSET %d", sqlbasestring, datecriteria, searchcriteria, sortBy, sortOrderSQL, limit, offset)
		} else {
			sqlstring = fmt.Sprintf("%s%s%s ORDER BY %s %s", sqlbasestring, datecriteria, searchcriteria, sortBy, sortOrderSQL)
		}
		countstring := fmt.Sprintf("SELECT COUNT(*) FROM purchaseHistory %s%s ", datecriteria, searchcriteria)
		log.Printf("\n%s\n", countstring)
		log.Printf("\n%s\n", sqlstring)

		var totalCount int
		err = dbConn.QueryRowContext(context.Background(), countstring).Scan(&totalCount)

		if err != nil {
			log.Println("Failed to fetch total count from the database:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch total count from the database"})
			return
		}

		rows, err := dbConn.QueryContext(context.Background(), sqlstring)
		if err != nil {
			log.Println("Failed to fetch paginated data from the database:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch paginated data from the database"})
			return
		}
		defer rows.Close()

		var purchases []model.Purchase

		for rows.Next() {
			var purchase model.Purchase
			if err := rows.Scan(&purchase.Po, &purchase.Pdate, &purchase.Item_List, &purchase.Description, &purchase.Qty, &purchase.Category, &purchase.Price, &purchase.User); err != nil {
				log.Println("Error scanning purchase data:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Error scanning purchase data"})
				return
			}
			purchases = append(purchases, purchase)
		}

		c.JSON(http.StatusOK, gin.H{
			"TotalRecords": totalCount,
			"Data":         purchases,
		})
	}
}

// ListCategoryAPI is a Gin handler function that returns category list as JSON
func ListCategoryAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Connect to the MySQL database
		dbConn := database.GetDB()

		// Build the SQL query using Squirrel
		query := squirrel.Select("category_code", "description").From("category_list")

		// Execute the SQL query
		rows, err := query.RunWith(dbConn).QueryContext(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch data from the database"})
			return
		}
		defer rows.Close()

		// Create a slice to store the category data
		var categories []model.Category

		// Iterate through the query results and append them to the slice
		for rows.Next() {
			var category model.Category
			if err := rows.Scan(&category.CategoryCode, &category.Description); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Error scanning category data"})
				return
			}
			categories = append(categories, category)
		}

		// Return the JSON response
		c.JSON(http.StatusOK, gin.H{"category": categories})
	}
}

// ListPurchaseAPI is a Gin handler function that returns Purchase list as JSON
func ListPurchaseAPI() gin.HandlerFunc {
		return func(c *gin.Context) {

		type ListPurchases struct {
			Page      int    `form:"page" validate:"omitempty,min=1,max=20"`     // Page number for pagination (optional), range: 1-20
			Limit     int    `form:"limit" validate:"omitempty,min=1,max=100"`   // Number of records per page (optional), range: 1-100
			SortBy    string `form:"sortBy" validate:"omitempty,max=15,ascii"`   // Field to sort by (optional), max length: 15, ASCII characters only
			SortOrder string `form:"sortOrder" validate:"omitempty,max=5,ascii"` // Sort order (optional), max length: 5, ASCII characters only
			Search    string `form:"search" validate:"omitempty,max=20,ascii"`   // Search query (optional), max length: 20, ASCII characters only
			DateFrom  string `form:"datef" validate:"omitempty,date"`            // Start date for filtering (optional), must be a valid date
			DateTo    string `form:"datee" validate:"omitempty,date"`            // End date for filtering (optional), must be a valid date
		}
		// Retrieve database connection
		dbConn := database.GetDB()

		// Bind query parameters to ListPurchase struct
		var ListPurchase ListPurchases
		if err := c.BindQuery(&ListPurchase); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid query parameters"})
			return
		}

		// Set default values for SortBy and SortOrder if not provided
		if ListPurchase.SortBy == "" {
			ListPurchase.SortBy = "PO"
		}
		if ListPurchase.SortOrder == "" {
			ListPurchase.SortOrder = "asc"
		}

		// Validate query parameters
		validate := validator.New()
		validate.RegisterValidation("date", CustomDateValidator)
		if err := validate.Struct(ListPurchase); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid query parameters", "errors": err})
			return
		}

		// Convert page and limit to integers
		pageStr := strconv.Itoa(ListPurchase.Page)
		limitStr := strconv.Itoa(ListPurchase.Limit)
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			if page != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page parameter"})
				return
			}
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			if limit != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid limit parameter (should be between 1 and 100)"})
				return
			}
		}
		offset := (page - 1) * limit

		// Determine SQL sort order
		var sortOrderSQL string
		if ListPurchase.SortOrder == "desc" {
			sortOrderSQL = "DESC"
		} else {
			sortOrderSQL = "ASC"
		}

		// Initialize SQL query builder
		var queryBuilder squirrel.SelectBuilder
		queryBuilder = squirrel.Select("PO, Pdate, item_list, description, qty, category, Price, User").From("purchaseHistory")
		countBuilder := squirrel.Select("COUNT(*)").From("purchaseHistory")

		// Add date filter if provided
		if ListPurchase.DateFrom != "" && ListPurchase.DateTo != "" {
			if _, err := time.Parse("2006-01-02", ListPurchase.DateFrom); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid datef parameter format. Should be YYYY-MM-DD"})
				return
			}
			if _, err := time.Parse("2006-01-02", ListPurchase.DateTo); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid datee parameter format. Should be YYYY-MM-DD"})
				return
			}
			queryBuilder = queryBuilder.Where(squirrel.Expr("Pdate BETWEEN ? AND ?", ListPurchase.DateFrom, ListPurchase.DateTo))
			countBuilder = countBuilder.Where(squirrel.Expr("Pdate BETWEEN ? AND ?", ListPurchase.DateFrom, ListPurchase.DateTo))
		}

		// Add search criteria if provided
		if ListPurchase.Search != "" {
			//sanitize serch field
			searchTerm := "%" + sanitize.HTML(sanitize.Scripts(helpers.SanitizeData(ListPurchase.Search).(string))) + "%"
			searchCriteria := squirrel.Or{
				squirrel.Like{"PO": searchTerm},
				squirrel.Like{"Pdate": searchTerm},
				squirrel.Like{"item_list": searchTerm},
				squirrel.Like{"description": searchTerm},
				squirrel.Like{"category": searchTerm},
				squirrel.Like{"Price": searchTerm},
				squirrel.Like{"User": searchTerm},
			}
			queryBuilder = queryBuilder.Where(searchCriteria)
			countBuilder = countBuilder.Where(searchCriteria)
		}

		// Add pagination and sorting
		if page > 0 && limit > 0 {
			queryBuilder = queryBuilder.OrderBy(fmt.Sprintf("%s %s", ListPurchase.SortBy, sortOrderSQL)).Limit(uint64(limit)).Offset(uint64(offset))
		} else {
			queryBuilder = queryBuilder.OrderBy(fmt.Sprintf("%s %s", ListPurchase.SortBy, sortOrderSQL))
		}

		// Generate SQL query string
		sqlstring, args, err := queryBuilder.ToSql()
		if err != nil {
			log.Println("Failed to build SQL query:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to build SQL query"})
			return
		}
		
		// Generate SQL query string for counting total records
		countString, countArgs, err := countBuilder.ToSql()
		if err != nil {
			log.Println("Failed to build count query:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to build count query"})
			return
		}

		// Execute count query to get total record count
		var totalCount int
		err = dbConn.QueryRowContext(context.Background(), countString, countArgs...).Scan(&totalCount)
		if err != nil {
			log.Println("Failed to fetch total count from the database:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch total count from the database"})
			return
		}

		// Execute main query to fetch paginated data
		rows, err := dbConn.QueryContext(context.Background(), sqlstring, args...)
		if err != nil {
			log.Println("Failed to fetch paginated data from the database:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch paginated data from the database"})
			return
		}
		defer rows.Close()

		// Scan rows into model.Purchase structs
		var purchases []model.Purchase
		for rows.Next() {
			var purchase model.Purchase
			if err := rows.Scan(&purchase.Po, &purchase.Pdate, &purchase.Item_List, &purchase.Description, &purchase.Qty, &purchase.Category, &purchase.Price, &purchase.User); err != nil {
				log.Println("Error scanning purchase data:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Error scanning purchase data"})
				return
			}
			purchases = append(purchases, purchase)
		}

		// Return paginated data and total record count as JSON response
		c.JSON(http.StatusOK, gin.H{
			"Data":         purchases,
			"TotalRecords": totalCount,
		})
	}
}


// CPublicTestAPI is a Gin handler function that handles the API endpoint for adding a purchase
func PublicTestAPI() gin.HandlerFunc {
	return func(c *gin.Context) {

		type ListPurchases struct {
			Page      int    `form:"page" validate:"omitempty,min=1,max=20"`     // Page number for pagination (optional), range: 1-20
			Limit     int    `form:"limit" validate:"omitempty,min=1,max=100"`   // Number of records per page (optional), range: 1-100
			SortBy    string `form:"sortBy" validate:"omitempty,max=15,ascii"`   // Field to sort by (optional), max length: 15, ASCII characters only
			SortOrder string `form:"sortOrder" validate:"omitempty,max=5,ascii"` // Sort order (optional), max length: 5, ASCII characters only
			Search    string `form:"search" validate:"omitempty,max=20,ascii"`   // Search query (optional), max length: 20, ASCII characters only
			DateFrom  string `form:"datef" validate:"omitempty,date"`            // Start date for filtering (optional), must be a valid date
			DateTo    string `form:"datee" validate:"omitempty,date"`            // End date for filtering (optional), must be a valid date
		}
		// Retrieve database connection
		dbConn := database.GetDB()

		// Bind query parameters to ListPurchase struct
		var ListPurchase ListPurchases
		if err := c.BindQuery(&ListPurchase); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid query parameters"})
			return
		}

		// Set default values for SortBy and SortOrder if not provided
		if ListPurchase.SortBy == "" {
			ListPurchase.SortBy = "PO"
		}
		if ListPurchase.SortOrder == "" {
			ListPurchase.SortOrder = "asc"
		}

		// Validate query parameters
		validate := validator.New()
		validate.RegisterValidation("date", CustomDateValidator)
		if err := validate.Struct(ListPurchase); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid query parameters", "errors": err})
			return
		}

		// Convert page and limit to integers
		pageStr := strconv.Itoa(ListPurchase.Page)
		limitStr := strconv.Itoa(ListPurchase.Limit)
		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			if page != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page parameter"})
				return
			}
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			if limit != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid limit parameter (should be between 1 and 100)"})
				return
			}
		}
		offset := (page - 1) * limit

		// Determine SQL sort order
		var sortOrderSQL string
		if ListPurchase.SortOrder == "desc" {
			sortOrderSQL = "DESC"
		} else {
			sortOrderSQL = "ASC"
		}

		// Initialize SQL query builder
		var queryBuilder squirrel.SelectBuilder
		queryBuilder = squirrel.Select("PO, Pdate, item_list, description, qty, category, Price, User").From("purchaseHistory")
		countBuilder := squirrel.Select("COUNT(*)").From("purchaseHistory")

		// Add date filter if provided
		if ListPurchase.DateFrom != "" && ListPurchase.DateTo != "" {
			if _, err := time.Parse("2006-01-02", ListPurchase.DateFrom); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid datef parameter format. Should be YYYY-MM-DD"})
				return
			}
			if _, err := time.Parse("2006-01-02", ListPurchase.DateTo); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid datee parameter format. Should be YYYY-MM-DD"})
				return
			}
			queryBuilder = queryBuilder.Where(squirrel.Expr("Pdate BETWEEN ? AND ?", ListPurchase.DateFrom, ListPurchase.DateTo))
			countBuilder = countBuilder.Where(squirrel.Expr("Pdate BETWEEN ? AND ?", ListPurchase.DateFrom, ListPurchase.DateTo))
		}

		// Add search criteria if provided
		if ListPurchase.Search != "" {
			//sanitize serch field
			searchTerm := "%" + sanitize.HTML(sanitize.Scripts(helpers.SanitizeData(ListPurchase.Search).(string))) + "%"
			searchCriteria := squirrel.Or{
				squirrel.Like{"PO": searchTerm},
				squirrel.Like{"Pdate": searchTerm},
				squirrel.Like{"item_list": searchTerm},
				squirrel.Like{"description": searchTerm},
				squirrel.Like{"category": searchTerm},
				squirrel.Like{"Price": searchTerm},
				squirrel.Like{"User": searchTerm},
			}
			queryBuilder = queryBuilder.Where(searchCriteria)
			countBuilder = countBuilder.Where(searchCriteria)
		}

		// Add pagination and sorting
		if page > 0 && limit > 0 {
			queryBuilder = queryBuilder.OrderBy(fmt.Sprintf("%s %s", ListPurchase.SortBy, sortOrderSQL)).Limit(uint64(limit)).Offset(uint64(offset))
		} else {
			queryBuilder = queryBuilder.OrderBy(fmt.Sprintf("%s %s", ListPurchase.SortBy, sortOrderSQL))
		}

		// Generate SQL query string
		sqlstring, args, err := queryBuilder.ToSql()
		if err != nil {
			log.Println("Failed to build SQL query:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to build SQL query"})
			return
		}
		
		// Generate SQL query string for counting total records
		countString, countArgs, err := countBuilder.ToSql()
		if err != nil {
			log.Println("Failed to build count query:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to build count query"})
			return
		}

		// Execute count query to get total record count
		var totalCount int
		err = dbConn.QueryRowContext(context.Background(), countString, countArgs...).Scan(&totalCount)
		if err != nil {
			log.Println("Failed to fetch total count from the database:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch total count from the database"})
			return
		}

		// Execute main query to fetch paginated data
		rows, err := dbConn.QueryContext(context.Background(), sqlstring, args...)
		if err != nil {
			log.Println("Failed to fetch paginated data from the database:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch paginated data from the database"})
			return
		}
		defer rows.Close()

		// Scan rows into model.Purchase structs
		var purchases []model.Purchase
		for rows.Next() {
			var purchase model.Purchase
			if err := rows.Scan(&purchase.Po, &purchase.Pdate, &purchase.Item_List, &purchase.Description, &purchase.Qty, &purchase.Category, &purchase.Price, &purchase.User); err != nil {
				log.Println("Error scanning purchase data:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Error scanning purchase data"})
				return
			}
			purchases = append(purchases, purchase)
		}

		// Return paginated data and total record count as JSON response
		c.JSON(http.StatusOK, gin.H{
			"Data":         purchases,
			"TotalRecords": totalCount,
		})
	}
}
