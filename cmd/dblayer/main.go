package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "dblayer",
	Short: "привер",
	Long: `Более подробное описание вашего приложения
           и его возможностей.`,
}

var exampleCmd = &cobra.Command{
	Use:   "migration [параметр]",
	Short: "Пример команды",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Printf("Выполняется команда с параметром: %s\n", args[0])
		} else {
			fmt.Println("Требуется указать параметр")
		}
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)

	exampleCmd.Flags().StringP("flag", "f", "", "Пример флага")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
