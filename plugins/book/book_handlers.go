package book

import (
	"fmt"
	"gothstack/plugins/auth"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
	"github.com/go-chi/chi/v5"
)

const (
	// Define where uploaded files will be stored
	uploadDir     = "./uploads/books"
	maxUploadSize = 50 << 20 // 50MB
)

// Validation schema for book creation
var bookSchema = v.Schema{
	//"name": v.Rules(v.Min(1), v.Max(200)),
}

// BookPageData holds data for the book pages
type BookPageData struct {
	FormValues BookFormValues
	FormErrors v.Errors
	Books      []Book
}

// BookFormValues holds form data for book creation
type BookFormValues struct {
	Name           string `form:"name"`
	Pages          int    `form:"pages"`
	SuccessMessage string
}

// Initialize the upload directory
func init() {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}
}

// HandleBookList handles the book list page
func HandleBookList(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID
	books, err := ListBooks(userID)
	if err != nil {
		return err
	}
	return kit.Render(BookList(BookPageData{Books: books}))
}

// HandleBookCreate handles the book creation form page
func HandleBookCreate(kit *kit.Kit) error {
	return kit.Render(BookCreate(BookPageData{}))
}

// HandleBookCreatePost handles the form submission for creating a book
func HandleBookCreatePost(kit *kit.Kit) error {
	// Parse multipart form with specified max memory
	if err := kit.Request.ParseMultipartForm(maxUploadSize); err != nil {
		return fmt.Errorf("error parsing form: %w", err)
	}

	// Extract form values
	var values BookFormValues
	errors, ok := v.Request(kit.Request, &values, bookSchema)
	if !ok {
		return kit.Render(BookForm(values, errors))
	}

	// Extract user ID from authentication
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID

	// Handle file upload
	file, header, err := kit.Request.FormFile("pdf_file")
	if err != nil {
		errors.Add("pdf_file", "PDF file is required")
		return kit.Render(BookForm(values, errors))
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		errors.Add("pdf_file", "Only PDF files are allowed")
		return kit.Render(BookForm(values, errors))
	}

	// Create a unique filename to prevent collisions
	fileExt := filepath.Ext(header.Filename)
	timestamp := time.Now().Unix()
	newFilename := fmt.Sprintf("%d_%d%s", userID, timestamp, fileExt)
	filePath := filepath.Join(uploadDir, newFilename)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	// Create book record in the database
	book, err := CreateBook(values.Name, userID, filePath, values.Pages, "Author")
	if err != nil {
		// Try to clean up the file if database insertion fails
		os.Remove(filePath)
		return fmt.Errorf("error creating book record: %w", err)
	}

	// Set success message
	values.SuccessMessage = fmt.Sprintf("New book '%s' uploaded successfully with ID %d", book.Name, book.ID)
	return kit.Render(BookForm(BookFormValues{SuccessMessage: values.SuccessMessage}, errors))
}

// HandleBookView handles viewing a specific book
func HandleBookView(kit *kit.Kit) error {
	idStr := chi.URLParam(kit.Request, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid book ID: %w", err)
	}

	userID := kit.Auth().(auth.Auth).UserID
	book, err := GetBook(uint(id), userID)
	if err != nil {
		return err
	}

	return kit.Render(BookView(book))
}

// HandleBookRead handles displaying the book for reading
func HandleBookRead(kit *kit.Kit) error {
	idStr := chi.URLParam(kit.Request, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid book ID: %w", err)
	}

	userID := kit.Auth().(auth.Auth).UserID
	book, err := GetBook(uint(id), userID)
	if err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(book.Location); os.IsNotExist(err) {
		return fmt.Errorf("PDF file not found: %w", err)
	}

	// Set content type for PDF
	kit.Response.Header().Set("Content-Type", "application/pdf")
	// Optional: Set content disposition if you want to force download
	// kit.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(book.Location)))

	// Serve the PDF file
	http.ServeFile(kit.Response, kit.Request, book.Location)
	return nil
}

// HandleBookDelete handles the deletion of a book
func HandleBookDelete(kit *kit.Kit) error {
	idStr := chi.URLParam(kit.Request, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid book ID: %w", err)
	}

	userID := kit.Auth().(auth.Auth).UserID

	// Get the book first to have access to its file location
	book, err := GetBook(uint(id), userID)
	if err != nil {
		return err
	}

	// Delete the book record
	if err := DeleteBook(uint(id), userID); err != nil {
		return err
	}
	fmt.Println(book)
	// Note: We're not deleting the actual file here as it's a soft delete
	// If you want to delete the file, uncomment the following:
	// if err := os.Remove(book.Location); err != nil {
	//     return fmt.Errorf("failed to delete file: %w", err)
	// }

	// Redirect back to the book list
	return kit.Redirect(http.StatusSeeOther, "/books")
}
