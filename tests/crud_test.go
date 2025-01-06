package tests

import (
	"context"
	"fmt"
	"github.com/antibomberman/dblayer/query"
	"testing"
	"time"
)

func TestCrud(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	user := User{
		Username: "John Doe",
		Email:    "john.doe@example.com",
		Phone:    "123",
		Password: "password123",
	}

	newUser := User{
		Username: "new user",
		Email:    "new.user@example.com",
		Phone:    "456",
		Password: "new_password",
	}
	// Create
	_, err = dbl.Query("users").Create(user)
	if err != nil {
		fmt.Println("error create user", err)
		t.Error(err)
	}

	_, err = dbl.Query("users").Create(user)
	if err != nil {
		t.Error(err)
	}
	_, err = dbl.Query("users").CreateMap(map[string]interface{}{
		"username": "Jane Doe",
		"email":    "jane.doe@example.com",
		"phone":    "456",
		"password": "secret",
	})
	if err != nil {
		fmt.Println("error createContext user", err)
		t.Error(err)
	}
	_, err = dbl.Query("users").CreateMap(map[string]interface{}{
		"username": "Jane Doe",
		"email":    "jane.doe@example.com",
		"phone":    "456",
		"password": "secret",
	})
	if err != nil {
		t.Error(err)
	}
	//update
	err = dbl.Query("users").WhereId(1).Update(newUser)
	if err != nil {
		t.Error(err)
	}
	err = dbl.Query("users").Where("email = ?", "john.doe@example.com").Update(newUser)
	if err != nil {
		t.Error(err)
	}
	err = dbl.Query("users").WhereId(3).UpdateMap(map[string]interface{}{
		"username": "new user map",
		"email":    "new.user.map@example.com",
		"phone":    "456",
		"password": "secret",
	})
	if err != nil {
		t.Error(err)
	}
	err = dbl.Query("users").WhereId(3).UpdateMap(map[string]interface{}{
		"username": "Jane Doe",
		"email":    "jane.doe@example.com",
		"phone":    "456",
		"password": "secret",
	})

	if err != nil {
		t.Error(err)
	}
	//delete
	err = dbl.Query("users").WhereId(1).Delete()
	if err != nil {
		t.Error(err)
	}
}
func TestPaginate(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}

	var users []User
	result, err := dbl.Query("users").Where("id > ?", 1).
		Where("id > ?", 2).
		Paginate(1, 10, &users)

	if err != nil {
		t.Error(err)
	}
	fmt.Println(result)

}
func TestAgr(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	count, err := dbl.Query("users").Where("id >?", 1).Count()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Total users:", count)
	avg, err := dbl.Query("users").Where("id >?", 1).Avg("id")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Average id:", avg)
	sum, err := dbl.Query("users").Where("id >?", 1).Sum("id")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Sum id:", sum)
	maxId, err := dbl.Query("users").Where("id >?", 1).Max("id")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Max id:", maxId)

	minId, err := dbl.Query("users").Where("id >?", 1).Min("id")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Min id:", minId)

	exists, err := dbl.Query("users").WhereId("id").Exists()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("User with id 1 exists:", exists)
}

func TestTransaction(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := dbl.Begin()
	if err != nil {
		t.Fatal(err)
	}
	user := User{
		Username: "test",
		Email:    "test@example.com",
		Phone:    "1",
		Password: "password",
	}
	_, err = tx.Query("users").Create(user)
	if err != nil {
		tx.Rollback()
		t.Error(err)
	}

	tx.Commit()

	err = dbl.Transaction(func(tx *query.Transaction) error {
		_, err := tx.Query("users").Create(user)
		return err
	})
	if err != nil {
		t.Error(err)
	}

}

func TestTrancate(t *testing.T) {

}
func TestWhere(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	var users []User
	_, err = dbl.Query("users").Where("id > ?", 1).OrWhereGroup(func(builder *query.Builder) {
		builder.Where("id > ?", 2).OrWhere("id < ?", 100)
	}).Get(&users)
	if err != nil {
		t.Error(err)
	}

}
func TestWhereDates(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	count, err := dbl.Query("users").WhereDate("created_at", "=", now).Count()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Users: ", count)

	count, err = dbl.Query("users").WhereDateTime("created_at", "<=", now).Count()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Users: ", count)

	count, err = dbl.Query("users").WhereBetweenDates("created_at", now.Add(-time.Hour*24*30), now).Count()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Users: ", count)

}
func TestCtx(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	qb := dbl.Query("users")

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	count, err := qb.Context(ctx).Count()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Users: ", count)

}
