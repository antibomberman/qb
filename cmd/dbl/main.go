package main

import (
	"fmt"
	"github.com/antibomberman/dbl/cmd/dbl/internal"
	"github.com/spf13/cobra"
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
	migrationCmd.AddCommand(internal.CreateCmd)
	migrationCmd.AddCommand(internal.UpCmd)
	migrationCmd.AddCommand(downCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
