package tests

import (
	"context"
	"github.com/antibomberman/dblayer"
	"testing"
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
	// Create
	dbl.Table("users").Create(user)
	dbl.Table("users").CreateContext(ctx, user)
	dbl.Table("users").CreateMap(map[string]interface{}{
		"username": "Jane Doe",
		"email":    "jane.doe@example.com",
		"phone":    "456",
		"password": "secret",
	})
	dbl.Table("users").CreateMapContext(ctx, map[string]interface{}{
		"username": "Jane Doe",
		"email":    "jane.doe@example.com",
		"phone":    "456",
		"password": "secret",
	})
	//update
	dbl.Table("users").WhereId(1).Create(user)
	dbl.Table("users").Where("email = ?", "john.doe@example.com").CreateContext(ctx, user)
	dbl.Table("users").WhereId(3).CreateMap(map[string]interface{}{
		"username": "Jane Doe",
		"email":    "jane.doe@example.com",
		"phone":    "456",
		"password": "secret",
	})
	dbl.Table("users").WhereId(3).CreateMapContext(ctx, map[string]interface{}{
		"username": "Jane Doe",
		"email":    "jane.doe@example.com",
		"phone":    "456",
		"password": "secret",
	})

}
