package main

import (
	"fmt"
	"github.com/antibomberman/dbl/cmd/dbl/internal"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "dbl",
	Short: "Приложение для управления миграциями",
}
var migrationCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Команды для работы с миграциями",
}

var createCmd = &cobra.Command{
	Use:   "create [название_миграции] ",
	Short: "Создать новую миграцию",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Необходимо указать название миграции")
			return
		}
		err := internal.Create(args[0])
		if err != nil {
			log.Fatal(err)
			return
		}

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

func main() {
	rootCmd.AddCommand(migrationCmd)
	migrationCmd.AddCommand(createCmd)
	migrationCmd.AddCommand(upCmd)
	migrationCmd.AddCommand(downCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
