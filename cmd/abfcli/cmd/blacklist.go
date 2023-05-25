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

var network string

// blackCmd represents the black command
var blackCmd = &cobra.Command{
	Use:   "blacklist",
	Short: "Add or remove network from blacklist",
	Long: `blacklist adds or removes network from blacklist.
	
	Networks in blacklist are always forbidden to query the source.

	Examples:
	blacklist add -n 33.33.33.33/24
	blacklist add --network 33.33.33.33/24
	blacklist remove -n 33.33.33.33/24
	blacklist remove --network 33.33.33.33/24`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("help") || cmd.Flags().Changed("h") {
			return nil
		}
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
			resp, err = http.Post("http://"+cfg.Server.Host+cfg.Server.Port+"/blacklist", "application/json", rdr)
		case args[0] == "remove":
			req, err := http.NewRequest(http.MethodDelete, "http://"+cfg.Server.Host+cfg.Server.Port+"/blacklist", rdr)
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
	blackCmd.Flags().StringVarP(&network, "network", "n", "", "Network to add to blacklist or remove from blacklist")
}
