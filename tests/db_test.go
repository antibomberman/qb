package tests

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"testing"
)

func TestAudit(t *testing.T) {
	dbl, err := ConnectDB()
	if err != nil {
		t.Fatal(err)
	}
	//err = dbl.AuditTableCreate()
	if err != nil {
		t.Fatalf("Ошибка создания таблицы аудита: %v", err)
	}
	user := User{
		Username: "tes3t",
		Email:    "t2est@example.com",
		Phone:    "1",
		Password: "password",
	}
	_, err = dbl.Table("users").WithAudit(1).Create(user)
	if err != nil {
		t.Fatalf("Ошибка создания записи в таблице: %v", err)
	}
	err = dbl.Table("users").OrderBy("id", "desc").Limit(1).Where("id > ?", 1).UpdateMap(map[string]interface{}{
		"username": "new name",
	})
	if err != nil {
		t.Fatalf("Ошибка %v", err)
	}
	err = dbl.Table("users").OrderBy("id", "desc").Limit(1).Update(user)
	if err != nil {
		t.Fatalf("Ошибка обновления записи в таблице: %v", err)
	}

}
