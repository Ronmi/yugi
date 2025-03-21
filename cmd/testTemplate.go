/*
Copyright © 2025 Ronmi Ren <ronmi.ren@gmail.com
*/
package cmd

import (
	"github.com/Ronmi/yugi/config"
	"github.com/Ronmi/yugi/web"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// testTemplateCmd represents the testTemplate command
var testTemplateCmd = &cobra.Command{
	Use:   "testTemplate",
	Short: "Test web templates with predefined test data",
	Run: func(cmd *cobra.Command, args []string) {
		err := config.RunSome(config.VarsConfig, config.URLConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init config")
		}

		err = web.LoadTmpl()
		if err != nil {
			log.Error().Err(err).Msg("failed to load templates")
		}

		log.Info().Msg("done")
	},
}

func init() {
	rootCmd.AddCommand(testTemplateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testTemplateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testTemplateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
