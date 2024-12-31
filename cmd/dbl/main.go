package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dbl",
	Short: "Приложение для управления миграциями",
}

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Команды для работы с миграциями",
}

var createCmd = &cobra.Command{
	Use:   "create [название_миграции]",
	Short: "Создать новую миграцию",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Необходимо указать название миграции")
			return
		}

		// Генерируем имя файла в формате: YYYYMMDDHHMMSS_название_миграции.sql
		timestamp := time.Now().Format("20060102150405")
		fileName := fmt.Sprintf("%s_%s.sql", timestamp, args[0])

		// Здесь можно добавить создание файла миграции
		fmt.Printf("Создана миграция: %s\n", fileName)
	},
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Применить все миграции",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Применение всех миграций...")
		// Здесь добавьте логику применения миграций
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Откатить последнюю миграцию",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Откат последней миграции...")
		// Здесь добавьте логику отката миграций
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Показать статус миграций",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Статус миграций:")
		// Здесь добавьте вывод статуса миграций
	},
}

func init() {
	rootCmd.AddCommand(migrationCmd)

	migrationCmd.AddCommand(createCmd)
	migrationCmd.AddCommand(upCmd)
	migrationCmd.AddCommand(downCmd)
	migrationCmd.AddCommand(statusCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
