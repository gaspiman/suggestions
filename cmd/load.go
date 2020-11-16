package cmd

import (
	"fmt"
	"log"
	"suggestions/load"

	"github.com/spf13/cobra"
)

var freqFlag string
var swFlag string
var titlesFlag string

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("load called")

		if err := load.Load(swFlag, freqFlag, titlesFlag); err != nil {
			log.Fatalf("load returned an error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	loadCmd.Flags().StringVarP(&freqFlag, "freq", "", "", "Frequency list file path")
	loadCmd.Flags().StringVarP(&titlesFlag, "titles", "", "", "Wikipedia Titles list file path")
	loadCmd.Flags().StringVarP(&swFlag, "sw", "", "", "Stopwords list file path")
}
