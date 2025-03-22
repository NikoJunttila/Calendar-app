package book

import (
	"fmt"
	"gothstack/app/db"
	"gothstack/plugins/auth"
	"time"

	"gorm.io/gorm"
)

// Book represents the books table in the database
type Book struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	OwnerID   uint   `gorm:"not null"`
	Location  string
	Pages     int
	Author    string
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relationship field
	User auth.User `gorm:"foreignKey:OwnerID"`
}

// Event name constants
const (
	BookCreatedEvent = "book.created"
	BookUpdatedEvent = "book.updated"
	BookDeletedEvent = "book.deleted"
)

// CreateBook creates a new book with the given information
func CreateBook(name string, ownerID uint, location string, pages int, author string) (Book, error) {
	book := Book{
		Name:      name,
		OwnerID:   ownerID,
		Location:  location,
		Pages:     pages,
		Author:    author,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	result := db.Get().Create(&book)
	if result.Error != nil {
		return book, fmt.Errorf("failed to create book: %w", result.Error)
	}
	return book, nil
}

// GetBook retrieves a book by its ID, ensuring it belongs to the specified owner
func GetBook(id, ownerID uint) (Book, error) {
	var book Book
	result := db.Get().Where("id = ? AND owner_id = ?", id, ownerID).First(&book)
	if result.Error != nil {
		return book, fmt.Errorf("failed to retrieve book: %w", result.Error)
	}
	return book, nil
}

// ListBooks returns all books for a specific owner
func ListBooks(ownerID uint) ([]Book, error) {
	var books []Book
	result := db.Get().Where("owner_id = ?", ownerID).Order("name asc").Find(&books)
	if result.Error != nil {
		return books, fmt.Errorf("failed to list books: %w", result.Error)
	}
	return books, nil
}

// UpdateBook updates an existing book
func UpdateBook(id uint, ownerID uint, name string, location string, pages int) (Book, error) {
	var book Book
	// First, check if the book exists and belongs to the owner
	if err := db.Get().Where("id = ? AND owner_id = ?", id, ownerID).First(&book).Error; err != nil {
		return book, fmt.Errorf("failed to retrieve book: %w", err)
	}

	// Update the book fields
	book.Name = name
	book.Location = location
	book.Pages = pages
	book.UpdatedAt = time.Now()

	// Save the updated book
	if err := db.Get().Save(&book).Error; err != nil {
		return book, fmt.Errorf("failed to update book: %w", err)
	}
	return book, nil
}

// DeleteBook soft-deletes a book (using GORM's DeletedAt)
func DeleteBook(id, ownerID uint) error {
	// We verify ownership before deletion for security
	result := db.Get().Where("id = ? AND owner_id = ?", id, ownerID).Delete(&Book{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete book: %w", result.Error)
	}
	// If no rows were affected, the book might not exist or doesn't belong to the owner
	if result.RowsAffected == 0 {
		return fmt.Errorf("book not found or doesn't belong to the specified owner")
	}
	return nil
}

// SearchBooks allows searching for books by name pattern
func SearchBooks(ownerID uint, namePattern string) ([]Book, error) {
	var books []Book
	result := db.Get().Where("owner_id = ? AND name LIKE ?", ownerID, "%"+namePattern+"%").
		Order("name asc").Find(&books)
	if result.Error != nil {
		return books, fmt.Errorf("failed to search books: %w", result.Error)
	}
	return books, nil
}

// GetBookProgress calculates reading progress based on current page
// This could be extended to include a separate BookProgress table for tracking
func GetBookProgress(bookID, ownerID, currentPage uint) (float64, error) {
	book, err := GetBook(bookID, ownerID)
	if err != nil {
		return 0, err
	}

	if book.Pages <= 0 {
		return 0, fmt.Errorf("book has no pages defined")
	}

	progress := float64(currentPage) / float64(book.Pages) * 100
	if progress > 100 {
		progress = 100
	}

	return progress, nil
}
