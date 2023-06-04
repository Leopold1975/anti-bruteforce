//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/app"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/config"
	"github.com/stretchr/testify/suite"
)

type ABFSuite struct {
	suite.Suite
	addr   string
	client *http.Client
}

func (a *ABFSuite) SetupSuite() {
	cfg := config.NewConfig("../../configs/config.toml")
	a.addr = "http://" + cfg.Server.Host + cfg.Server.Port

	if err := os.Chdir("../../"); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("make", "run")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	c := http.DefaultClient
	c.Timeout = time.Second * 5
	a.client = c
}

func (a *ABFSuite) TearDownSuite() {
	cmd := exec.Command("make", "down")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *ABFSuite) TestTryAuthAnd404() {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.addr+"/", nil)
	a.Require().NoError(err)
	resp, err := a.client.Do(req)

	a.Require().NoError(err)
	defer resp.Body.Close()
	a.Require().Equal(http.StatusNotFound, resp.StatusCode)

	r := app.Request{
		Login:    "login",
		Password: "password",
		IP:       "33.33.34.33",
	}
	jReq, err := json.Marshal(r)
	a.Require().NoError(err)
	for i := 0; i < 10; i++ {
		rdr := bytes.NewReader(jReq)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
		a.Require().NoError(err)

		resp, err = a.client.Do(req)
		a.Require().NoError(err)
		a.Require().Equal(http.StatusOK, resp.StatusCode, r)
		resp.Body.Close()
	}

	rdr := bytes.NewReader(jReq)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
	a.Require().NoError(err)

	resp, err = a.client.Do(req)
	a.Require().NoError(err)
	a.Require().Equal(http.StatusTooManyRequests, resp.StatusCode)
	resp.Body.Close()

	for i := 0; i < 100; i++ {
		r := app.Request{
			Login:    "login" + strconv.Itoa(i),
			Password: "password1",
			IP:       "33.33.34." + strconv.Itoa(i),
		}
		jReq, err := json.Marshal(r)
		a.Require().NoError(err)

		rdr := bytes.NewReader(jReq)
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
		a.Require().NoError(err)

		resp, err = a.client.Do(req)
		a.Require().NoError(err)
		a.Require().Equal(http.StatusOK, resp.StatusCode, r)
		resp.Body.Close()
	}

	r = app.Request{
		Login:    "login",
		Password: "password1",
		IP:       "33.33.34.35",
	}
	jReq, err = json.Marshal(r)
	a.Require().NoError(err)

	rdr = bytes.NewReader(jReq)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
	a.Require().NoError(err)
	resp, err = a.client.Do(req)
	a.Require().NoError(err)
	a.Require().Equal(http.StatusTooManyRequests, resp.StatusCode)
	resp.Body.Close()

	for i := 0; i < 100; i++ {
		r := app.Request{
			Login:    "new_login" + strconv.Itoa(i),
			Password: "new_password" + strconv.Itoa(i),
			IP:       "44.44.44.36",
		}
		jReq, err := json.Marshal(r)
		a.Require().NoError(err)

		rdr := bytes.NewReader(jReq)
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
		a.Require().NoError(err)
		resp, err = a.client.Do(req)
		a.Require().NoError(err)
		a.Require().Equal(http.StatusOK, resp.StatusCode, r)
		resp.Body.Close()
	}

	r = app.Request{
		Login:    "new_login",
		Password: "new_password",
		IP:       "44.44.44.36",
	}
	jReq, err = json.Marshal(r)
	a.Require().NoError(err)

	rdr = bytes.NewReader(jReq)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
	a.Require().NoError(err)
	resp, err = a.client.Do(req)
	a.Require().NoError(err)
	a.Require().NoError(err)
	a.Require().Equal(http.StatusTooManyRequests, resp.StatusCode)
	resp.Body.Close()
}

func (a *ABFSuite) TestBlacklist() {
	ctx := context.Background()
	rq := app.Network{
		IP: "55.55.55.0/24",
	}
	jReq, err := json.Marshal(rq)
	a.Require().NoError(err)

	rdr := bytes.NewReader(jReq)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/blacklist", rdr)
	a.Require().NoError(err)

	resp, err := a.client.Do(req)
	a.Require().NoError(err)
	a.Require().Equal(http.StatusOK, resp.StatusCode, rq)
	resp.Body.Close()

	r := app.Request{
		Login:    "blacklist",
		Password: "black",
		IP:       "55.55.55.33",
	}
	jReq, err = json.Marshal(r)
	a.Require().NoError(err)
	for i := 0; i < 10; i++ {
		rdr := bytes.NewReader(jReq)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
		a.Require().NoError(err)

		resp, err = a.client.Do(req)
		a.Require().NoError(err)
		a.Require().Equal(http.StatusTooManyRequests, resp.StatusCode, r)
		resp.Body.Close()
	}
	jReq, err = json.Marshal(rq)
	a.Require().NoError(err)
	rdr = bytes.NewReader(jReq)
	req, err = http.NewRequestWithContext(ctx, http.MethodDelete, a.addr+"/blacklist", rdr)
	a.Require().NoError(err)

	resp, err = a.client.Do(req)
	a.Require().NoError(err)
	a.Require().Equal(http.StatusOK, resp.StatusCode, rq)
	resp.Body.Close()

	r = app.Request{
		Login:    "blacklist2",
		Password: "black2",
		IP:       "55.55.55.2",
	}
	jReq, err = json.Marshal(r)
	a.Require().NoError(err)
	for i := 0; i < 10; i++ {
		rdr = bytes.NewReader(jReq)
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
		a.Require().NoError(err)

		resp, err = a.client.Do(req)
		a.Require().NoError(err)

		a.Require().Equal(http.StatusOK, resp.StatusCode, r)
		resp.Body.Close()
	}
}

func (a *ABFSuite) TestWhitelist() {
	ctx := context.Background()
	rq := app.Network{
		IP: "66.66.66.0/24",
	}
	jReq, err := json.Marshal(rq)
	a.Require().NoError(err)

	rdr := bytes.NewReader(jReq)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/whitelist", rdr)
	a.Require().NoError(err)

	resp, err := a.client.Do(req)
	a.Require().NoError(err)
	a.Require().Equal(http.StatusOK, resp.StatusCode, rq)
	resp.Body.Close()

	r := app.Request{
		Login:    "whitelist",
		Password: "white",
		IP:       "66.66.66.66",
	}
	jReq, err = json.Marshal(r)
	a.Require().NoError(err)
	for i := 0; i < 100; i++ {
		rdr := bytes.NewReader(jReq)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
		a.Require().NoError(err)

		resp, err = a.client.Do(req)
		a.Require().NoError(err)
		a.Require().Equal(http.StatusOK, resp.StatusCode, r)
		resp.Body.Close()
	}
	jReq, err = json.Marshal(rq)
	a.Require().NoError(err)
	rdr = bytes.NewReader(jReq)
	req, err = http.NewRequestWithContext(ctx, http.MethodDelete, a.addr+"/whitelist", rdr)
	a.Require().NoError(err)

	resp, err = a.client.Do(req)
	a.Require().NoError(err)
	a.Require().Equal(http.StatusOK, resp.StatusCode, rq)
	resp.Body.Close()

	r = app.Request{
		Login:    "whitelist1",
		Password: "white1",
		IP:       "66.66.66.66",
	}
	jReq, err = json.Marshal(r)
	a.Require().NoError(err)
	for i := 0; i < 10; i++ {
		rdr = bytes.NewReader(jReq)
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
		a.Require().NoError(err)

		resp, err = a.client.Do(req)
		a.Require().NoError(err)

		a.Require().Equal(http.StatusOK, resp.StatusCode, r)
		resp.Body.Close()
	}
	rdr = bytes.NewReader(jReq)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
	a.Require().NoError(err)

	resp, err = a.client.Do(req)
	a.Require().NoError(err)

	a.Require().Equal(http.StatusTooManyRequests, resp.StatusCode, r)
	resp.Body.Close()
}

func (a *ABFSuite) TestReset() {
	ctx := context.Background()
	r := app.Request{
		Login:    "reset_test",
		Password: "reset",
		IP:       "77.77.77.33",
	}
	jReq, err := json.Marshal(r)
	a.Require().NoError(err)
	for i := 0; i < 10; i++ {
		rdr := bytes.NewReader(jReq)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
		a.Require().NoError(err)

		resp, err := a.client.Do(req)
		a.Require().NoError(err)
		a.Require().Equal(http.StatusOK, resp.StatusCode, r)
		resp.Body.Close()
	}

	rdr := bytes.NewReader(jReq)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
	a.Require().NoError(err)

	resp, err := a.client.Do(req)
	a.Require().NoError(err)
	a.Require().Equal(http.StatusTooManyRequests, resp.StatusCode, r)
	resp.Body.Close()

	rdr = bytes.NewReader(jReq)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/reset", rdr)
	a.Require().NoError(err)

	resp, err = a.client.Do(req)
	a.Require().NoError(err)
	a.Require().Equal(http.StatusOK, resp.StatusCode, r)
	resp.Body.Close()

	rdr = bytes.NewReader(jReq)
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, a.addr+"/try", rdr)
	a.Require().NoError(err)

	resp, err = a.client.Do(req)
	a.Require().NoError(err)
	a.Require().Equal(http.StatusOK, resp.StatusCode, r)
	resp.Body.Close()
}

func TestABFSuite(t *testing.T) {
	suite.Run(t, new(ABFSuite))
}
