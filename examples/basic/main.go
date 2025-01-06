package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/dblayer"
	QB "github.com/antibomberman/dblayer/query"
	t "github.com/antibomberman/dblayer/table"
	_ "modernc.org/sqlite"
)

// Data structures for working with database
type User struct {
	ID        int64     `db:"id"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	Age       int       `db:"age"`
	CreatedAt time.Time `db:"created_at"`
	Version   int       `db:"version"`
}
type Post struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}
type Comment struct {
	ID        int64     `db:"id"`
	PostID    int64     `db:"post_id"`
	UserID    int64     `db:"user_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

func main() {
	// Connect to database
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/testdb?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Initialize DBL
	dbl := dblayer.New("mysql", db)
	// Create tables
	if err := createTables(dbl); err != nil {
		log.Fatal(err)
	}
	// Usage examples
	if err := examples(dbl); err != nil {
		log.Fatal(err)
	}

	count, err := dbl.Query("users").WhereAge("age", ">", 3).WhereDateIsNotNull("created_at").Count()
	if err != nil {
		fmt.Printf("Error count table")
	}
	fmt.Printf("Number of users older than 3: %d\n", count)

	//example 2
	var post Post

	fount, err := dbl.Query("posts").First(&post)
	if err != nil {
		fmt.Printf("Error")
	}
	if !fount {
		fmt.Printf("No post")
	}
	fmt.Printf("Title: %s, Content: %s\n", post.Title, post.Content)

}

func createTables(dbl *dblayer.DBLayer) error {
	// Create users table
	_, err := dbl.Table("users").CreateIfNotExists(func(schema *t.Builder) {
		schema.ID()
		schema.String("email", 255).Unique()
		schema.String("name", 100)
		schema.Integer("age")
		schema.Timestamp("created_at")
		schema.Integer("version").Default(1)
	})
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}
	// Create posts table
	_, err = dbl.Table("posts").CreateIfNotExists(func(schema *t.Builder) {
		schema.Integer("id").Primary().AutoIncrement()
		schema.Integer("user_id")
		schema.String("title", 200)
		schema.Text("content")
		schema.Timestamp("created_at")
	})
	if err != nil {
		return fmt.Errorf("error creating posts table: %w", err)
	}
	// Create comments table
	_, err = dbl.Table("comments").CreateIfNotExists(func(schema *t.Builder) {
		schema.Integer("id").Primary().AutoIncrement()
		schema.Integer("post_id")
		schema.Integer("user_id")
		schema.Text("content")
		schema.Timestamp("created_at")
	})
	if err != nil {
		return fmt.Errorf("error creating comments table: %w", err)
	}
	return nil
}
func examples(dbl *dblayer.DBLayer) error {
	// 1. Create user with transaction and audit
	tx, err := dbl.Begin()
	if err != nil {
		return err
	}
	user := &User{
		Email:     "agabek309@gmail.com",
		Name:      "Agabek Backend Developer",
		Age:       25,
		CreatedAt: time.Now(),
		Version:   1,
	}
	userID, err := tx.Query("users").
		WithAudit(1). // Admin ID
		Create(user)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 2. Create post for user
	post := &Post{
		UserID:    userID,
		Title:     "My first post",
		Content:   "Hello world!",
		CreatedAt: time.Now(),
	}
	postID, err := tx.Query("posts").Create(post)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 3. Add comment
	comment := &Comment{
		PostID:    postID,
		UserID:    userID,
		Content:   "First comment",
		CreatedAt: time.Now(),
	}
	_, err = tx.Query("comments").Create(comment)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}
	// 4. Complex query with JOIN and caching
	var results []struct {
		UserName     string    `db:"name"`
		PostTitle    string    `db:"title"`
		CommentCount int       `db:"comment_count"`
		LastComment  time.Time `db:"last_comment"`
	}
	_, err = dbl.Query("users").
		Select("users.name, posts.title, COUNT(comments.id) as comment_count, MAX(comments.created_at) as last_comment").
		LeftJoin("posts", "posts.user_id = users.id").
		LeftJoin("comments", "comments.post_id = posts.id").
		GroupBy("users.id, posts.id").
		Having("comment_count > 0").
		OrderBy("last_comment", "DESC").
		Remember("user_posts_stats", 5*time.Minute).
		Get(&results)
	if err != nil {
		return err
	}
	// 5. Batch update with optimistic locking
	users := []map[string]interface{}{
		{"id": 1, "age": 26, "version": 2},
		{"id": 2, "age": 31, "version": 2},
	}
	err = dbl.Query("users").BatchUpdate(users, "id", 100)
	if err != nil {
		return err
	}
	// 6. Full-text search
	var searchResults []Post
	_, err = dbl.Query("posts").
		Search([]string{"title", "content"}, "hello").
		OrderBy("search_rank", "DESC").
		Limit(10).
		Get(&searchResults)
	if err != nil {
		return err
	}
	// 7. Pagination with metrics
	var pagedUsers []User
	collector := QB.NewMetricsCollector()
	result, err := dbl.Query("users").
		WithMetrics(collector).
		OrderBy("created_at", "DESC").
		Paginate(1, 10, &pagedUsers)
	if err != nil {
		return err
	}
	fmt.Printf("Total users: %d, Pages: %d\n", result.Total, result.LastPage)
	return nil
}
