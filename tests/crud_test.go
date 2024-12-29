package tests

import (
	"context"
	"testing"

	"github.com/antibomberman/dblayer"
)

func TestCrud(t *testing.T) {
	ctx := context.Background()
	dbl, err := dblayer.Connection(ctx, driver, dsn, maxAttempts, timeout)
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer dbl.Close()
	err = dbl.Ping()
	if err != nil {
		t.Fatalf("Ошибка подключения к БД: %v", err)
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
	_, err = dbl.Table("users").Create(user)
	if err != nil {
		t.Error(err)
	}

	_, err = dbl.Table("users").CreateContext(ctx, user)
	if err != nil {
		t.Error(err)
	}
	_, err = dbl.Table("users").CreateMap(map[string]interface{}{
		"username": "Jane Doe",
		"email":    "jane.doe@example.com",
		"phone":    "456",
		"password": "secret",
	})
	if err != nil {
		t.Error(err)
	}
	_, err = dbl.Table("users").CreateMapContext(ctx, map[string]interface{}{
		"username": "Jane Doe",
		"email":    "jane.doe@example.com",
		"phone":    "456",
		"password": "secret",
	})
	if err != nil {
		t.Error(err)
	}
	//update
	err = dbl.Table("users").WhereId(1).Update(newUser)
	if err != nil {
		t.Error(err)
	}
	err = dbl.Table("users").Where("email = ?", "john.doe@example.com").UpdateContext(ctx, newUser)
	if err != nil {
		t.Error(err)
	}
	err = dbl.Table("users").WhereId(3).UpdateMap(map[string]interface{}{
		"username": "new user map",
		"email":    "new.user.map@example.com",
		"phone":    "456",
		"password": "secret",
	})
	if err != nil {
		t.Error(err)
	}
	err = dbl.Table("users").WhereId(3).UpdateMapContext(ctx, map[string]interface{}{
		"username": "Jane Doe",
		"email":    "jane.doe@example.com",
		"phone":    "456",
		"password": "secret",
	})

	if err != nil {
		t.Error(err)
	}
}
