/*
Copyright © 2025 Ronmi Ren <ronmi.ren@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/Ronmi/yugi/web"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
		defer stop()

		if err := config.Init(ctx); err != nil {
			fmt.Println("failed to initialize configurations:", err)
			return
		}

		if err := actions.Migrate(config.DB); err != nil {
			fmt.Println("failed to migrate database:", err)
			return
		}

		srv, err := web.New(config.DB)
		if err != nil {
			fmt.Println("failed to start web server:", err)
			return
		}

		log.Info().Msg("starting web server")
		srv.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

}
