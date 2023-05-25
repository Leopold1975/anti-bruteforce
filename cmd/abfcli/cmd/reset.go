package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/app"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/config"
	"github.com/spf13/cobra"
)

var ip, login string

// reset clears up buckets for the given login and ip
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "reset clears up buckets for the given login and ip",
	Long: `reset takes two argument: ip and login to reset bucket for. 
	
	Examples:
	reset -i 33.33.33.33 -l login
	reset 33.33.33.33 login
	reset --config path/to/config.toml -i 33.33.33.33 -l login
	Calling this will clear bucket for 33.33.33.33 ip and login.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 && (ip == "" || login == "") {
			return ErrNotEnoughArguments
		}

		req := app.Request{
			Login: login,
			IP:    ip,
		}
		jReq, err := json.Marshal(req)
		if err != nil {
			return err
		}
		rdr := bytes.NewReader(jReq)

		cfg := config.NewConfig(configFile)
		resp, err := http.Post("http://"+cfg.Server.Host+cfg.Server.Port+"/reset", "application/json", rdr)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("server error code: %d", resp.StatusCode)
		}
		return nil
	},
	SilenceErrors: true,
}

func init() {
	resetCmd.Flags().StringVarP(&ip, "ip", "i", "", "IP to reset buckets for")
	resetCmd.Flags().StringVarP(&login, "login", "l", "", "Login to reset buckets for")
	resetCmd.Flags().StringVarP(&configFile, "config", "c", "configs/config_cli.toml", "Path to configuration file")
}
