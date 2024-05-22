package main

import (
	"errors"
	"net/http"
	"strconv"
	"database/sql" 
	_ "modernc.org/sqlite"

	"github.com/gin-gonic/gin"
	"example/go-project/db"
)

type book struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	ISBN    int    `json:"isbn"`
	Author  string `json:"author"`
	Release int    `json:"release"`
}

func getBooks(c *gin.Context) {
	var books []book
	rows, err := db.GetDb().Query("SELECT id, title, isbn, author, release FROM books")
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error."})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var b book
		if err := rows.Scan(&b.ID, &b.Title, &b.ISBN, &b.Author, &b.Release); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error."})
			return
		}
		books = append(books, b)
	}
	c.IndentedJSON(http.StatusOK, books)
}

func bookById(c *gin.Context) {
	id := c.Param("id")
	book, err := getBookById(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Book not found."})
		return
	}

	c.IndentedJSON(http.StatusOK, book)
}

func getBookById(id string) (*book, error) {
	bookID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var b book
	err = db.GetDb().QueryRow("SELECT id, title, isbn, author, release FROM books WHERE id = ?", bookID).Scan(&b.ID, &b.Title, &b.ISBN, &b.Author, &b.Release)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("book not found")
		}
		return nil, err
	}
	return &b, nil
}

func createBook(c *gin.Context) {
	var newBook book

	if err := c.BindJSON(&newBook); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request."})
		return
	}

	res, err := db.GetDb().Exec("INSERT INTO books (title, isbn, author, release) VALUES (?, ?, ?, ?)", newBook.Title, newBook.ISBN, newBook.Author, newBook.Release)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error."})
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error."})
		return
	}
	newBook.ID = id
	c.IndentedJSON(http.StatusCreated, newBook)
}

func updateBookById(c *gin.Context) {
	id := c.Param("id")
	bookID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid book ID."})
		return
	}

	var updatedBook book
	if err := c.BindJSON(&updatedBook); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request."})
		return
	}

	updatedBook.ID = bookID 

	_, err = db.GetDb().Exec("UPDATE books SET title = ?, isbn = ?, author = ?, release = ? WHERE id = ?", updatedBook.Title, updatedBook.ISBN, updatedBook.Author, updatedBook.Release, bookID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error."})
		return
	}
	c.IndentedJSON(http.StatusOK, updatedBook)
}

func updateBook(c *gin.Context) {
	var updatedBook book

	if err := c.BindJSON(&updatedBook); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request."})
		return
	}

	bookID := updatedBook.ID

	_, err := db.GetDb().Exec("UPDATE books SET title = ?, isbn = ?, author = ?, release = ? WHERE id = ?", updatedBook.Title, updatedBook.ISBN, updatedBook.Author, updatedBook.Release, bookID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error."})
		return
	}
	c.IndentedJSON(http.StatusOK, updatedBook)
}

func deleteBook(c *gin.Context) {
	id := c.Param("id")
	bookID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid book ID."})
		return
	}

	_, err = db.GetDb().Exec("DELETE FROM books WHERE id = ?", bookID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Database error."})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Book deleted."})
}

func main() {
	db.Init()
	router := gin.Default()
	router.GET("/books", getBooks)
	router.GET("/books/:id", bookById)
	router.POST("/books", createBook)
	router.PUT("/books/:id", updateBookById)
	router.PUT("/books", updateBook)
	router.DELETE("/books/:id", deleteBook)
	router.Run("localhost:8080")
}
