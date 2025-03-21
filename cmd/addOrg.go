/*
Copyright © 2025 Ronmi Ren <ronmi.ren@gmail.com
*/
package cmd

import (
	"fmt"

	"github.com/Ronmi/yugi/actions"
	"github.com/Ronmi/yugi/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// addOrgCmd represents the addOrg command
var addOrgCmd = &cobra.Command{
	Use:   "addOrg",
	Short: "新增一個罷免團隊及初始成員",
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())

		name := viper.GetString("name")
		area := viper.GetStringSlice("area")
		target := viper.GetString("target")
		manager := viper.GetString("manager")

		if name == "" {
			fmt.Println("請指定團隊名稱")
			return
		}
		if len(area) < 2 {
			fmt.Println("請指定至少兩個罷免區域 (範例: -a 中五選區 -a 北區 -a 北屯區")
			return
		}
		if target == "" {
			fmt.Println("請指定罷免目標")
			return
		}
		if manager == "" {
			fmt.Println("請指定初始幹部")
			return
		}

		err := config.RunSome(config.DBConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to init db")
		}

		log.Info().Str("name", name).Strs("area", area).Str("target", target).Str("manager", manager).Msg("團隊資訊")

		err = actions.AddOrg(config.DB, actions.Org{
			Name:   name,
			Area:   area,
			Target: target,
		}, manager)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to add org")
		}

		log.Info().Msg("團隊新增成功")
	},
}

func init() {
	rootCmd.AddCommand(addOrgCmd)

	f := addOrgCmd.Flags()
	f.StringP("name", "n", "", "團隊名稱")
	f.StringSliceP("area", "a", nil, "罷免區域")
	f.StringP("target", "t", "", "罷免目標")
	f.StringP("manager", "m", "", "幹部名稱 (四字中文短語)")
}
