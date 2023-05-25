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

// whiteCmd represents the white command
var whiteCmd = &cobra.Command{
	Use:   "whitelist",
	Short: "Add or remove network from whitelist",
	Long: `whitelist adds or removes network from whitelist.
	
	Networks in whitelist are always allowed to query the source.

	Examples:
	whitelist add -n 33.33.33.33/24
	whitelist add --network 33.33.33.33/24
	whitelist remove -n 33.33.33.33/24
	whitelist remove --network 33.33.33.33/24`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return ErrNotEnoughArguments
		}

		req := app.Network{
			IP: network,
		}

		jReq, err := json.Marshal(req)
		if err != nil {
			return err
		}
		rdr := bytes.NewReader(jReq)

		cfg := config.NewConfig(configFile)

		var resp *http.Response
		switch {
		case args[0] == "add":
			resp, err = http.Post("http://"+cfg.Server.Host+cfg.Server.Port+"/whitelist", "application/json", rdr)
		case args[0] == "remove":
			req, err := http.NewRequest(http.MethodDelete, "http://"+cfg.Server.Host+cfg.Server.Port+"/whitelist", rdr)
			if err != nil {
				return err
			}
			client := http.Client{}
			resp, err = client.Do(req)
		}
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("server error code: %d", resp.StatusCode)
		}

		return nil
	},
}

func init() {
	whiteCmd.Flags().StringVarP(&network, "network", "n", "", "Network to add to whitelist or remove from whitelist")
}
