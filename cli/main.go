package main

import (
	"fmt"
	"github.com/antibomberman/dblayer/cli/internal"
	"github.com/antibomberman/dblayer/migrate"
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

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Откатить последнюю миграцию",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Откат последней миграции...")
		// Здесь добавьте логику отката миграций
	},
}
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "generate default files",
	Run: func(cmd *cobra.Command, args []string) {
		migrate.InitDir()
		internal.GenerateEnv()
	},
}
var CreateCmd = &cobra.Command{
	Use:   "create [название_миграции] ",
	Short: "Создать новую миграцию",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Необходимо указать название миграции")
			return
		}
		migrate.InitDir()
		internal.GenerateEnv()

		path, err := migrate.Create(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Файл миграции создан: ", path)

	},
}
var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Применить все миграции",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Применение миграций...")
		migrate.Up()
	},
}

func main() {
	rootCmd.AddCommand(migrationCmd)
	migrationCmd.AddCommand(CreateCmd)
	migrationCmd.AddCommand(UpCmd)
	migrationCmd.AddCommand(DownCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
