package cmd

import (
	"auto-tester-cli/executor"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Définition de la commande principale
var rootCmd = &cobra.Command{
	Use:   "auto-tester-cli",
	Short: "Auto-tester-cli exécute et teste votre code sans effort.",
	Long:  "Auto-tester-cli permet d'exécuter du code dans différents langages via des conteneurs Docker.",
	Run: func(cmd *cobra.Command, args []string) {
		lang, _ := cmd.Flags().GetString("lang")
		file, _ := cmd.Flags().GetString("file")
		if lang == "" || file == "" {
			fmt.Println("Langage et fichier requis. Utilisez --lang et --file.")
			os.Exit(1)
		}

		executor.ExecuteCode(lang, file)
	},
}

// Configuration des flags (options)
func init() {
	rootCmd.Flags().StringP("lang", "l", "", "Langage du code (ex: python, java)")
	rootCmd.Flags().StringP("file", "f", "", "Chemin du fichier source à exécuter")
}

// Execute lance la commande CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
