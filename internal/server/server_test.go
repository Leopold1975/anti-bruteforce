//nolint:dupl
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/app"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/config"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/redislimiter"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/pkg/logger"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/require"
)

//nolint:funlen
func TestServer(t *testing.T) {
	rdb, mock := redismock.NewClientMock()

	defer rdb.Close()

	testCfg := config.Config{
		Logger: config.LoggerConf{
			Level: "INFO",
		},
		Limiter: config.LimiterConf{
			N: 10,
			M: 90,
			K: 100,
		},
	}

	logger := logger.New(testCfg)
	tService := redislimiter.NewServiceMock(rdb, testCfg)

	serv := New(tService, testCfg, logger)
	t.Run("not found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("test login bruteforce", func(t *testing.T) { //nolint:funlen
		r := app.Request{
			Login:    "login",
			Password: "password",
			IP:       "33.33.34.33",
		}
		jReq, err := json.Marshal(r)
		require.NoError(t, err)

		for i := 1; i < 10; i++ {
			mock.ExpectSMembers("blacklist").SetVal([]string{})
			mock.ExpectSMembers("whitelist").SetVal([]string{})
			mock.ExpectGet(r.Login).SetVal(strconv.Itoa(i))
			mock.ExpectGet(r.Password).SetVal(strconv.Itoa(i))
			mock.ExpectGet(r.IP).SetVal(strconv.Itoa(i))

			mock.ExpectTxPipeline()
			mock.ExpectIncr(r.Login).SetVal(1)
			mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
			mock.ExpectIncr(r.Password).SetVal(1)
			mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
			mock.ExpectIncr(r.IP).SetVal(1)
			mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
			mock.ExpectTxPipelineExec()

			rdr := bytes.NewReader(jReq)
			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodPost, "/try", rdr)
			require.NoError(t, err)

			serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

			require.Equal(t, http.StatusOK, rr.Code)
		}

		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("11")
		mock.ExpectGet(r.Password).SetVal("11")
		mock.ExpectGet(r.IP).SetVal("11")

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		rdr := bytes.NewReader(jReq)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusTooManyRequests, rr.Code)
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("test password bruteforce", func(t *testing.T) {
		var i int
		for i = 1; i < 90; i++ {
			r := app.Request{
				Login:    "login" + strconv.Itoa(i),
				Password: "password",
				IP:       "33.33." + strconv.Itoa(i) + "." + strconv.Itoa(i),
			}
			mock.ExpectSMembers("blacklist").SetVal([]string{})
			mock.ExpectSMembers("whitelist").SetVal([]string{})
			mock.ExpectGet(r.Login).SetVal("1")
			mock.ExpectGet(r.Password).SetVal(strconv.Itoa(i))
			mock.ExpectGet(r.IP).SetVal(strconv.Itoa(i))

			mock.ExpectTxPipeline()
			mock.ExpectIncr(r.Login).SetVal(1)
			mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
			mock.ExpectIncr(r.Password).SetVal(1)
			mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
			mock.ExpectIncr(r.IP).SetVal(1)
			mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
			mock.ExpectTxPipelineExec()

			jReq, err := json.Marshal(r)
			require.NoError(t, err)
			rdr := bytes.NewReader(jReq)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodPost, "/try", rdr)
			require.NoError(t, err)

			serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

			require.Equal(t, http.StatusOK, rr.Code)
		}

		r := app.Request{
			Login:    "login",
			Password: "password",
			IP:       "33.33.35.33",
		}

		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("1")
		mock.ExpectGet(r.Password).SetVal(strconv.Itoa(i))
		mock.ExpectGet(r.IP).SetVal(strconv.Itoa(i))

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		jReq, err := json.Marshal(r)
		require.NoError(t, err)
		rdr := bytes.NewReader(jReq)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusTooManyRequests, rr.Code)
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("test ip bruteforce", func(t *testing.T) {
		var i int
		for i = 1; i < 100; i++ {
			r := app.Request{
				Login:    "login" + strconv.Itoa(i),
				Password: "password" + strconv.Itoa(i),
				IP:       "33.33.33.33",
			}
			mock.ExpectSMembers("blacklist").SetVal([]string{})
			mock.ExpectSMembers("whitelist").SetVal([]string{})
			mock.ExpectGet(r.Login).SetVal("1")
			mock.ExpectGet(r.Password).SetVal("1")
			mock.ExpectGet(r.IP).SetVal(strconv.Itoa(i))

			mock.ExpectTxPipeline()
			mock.ExpectIncr(r.Login).SetVal(1)
			mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
			mock.ExpectIncr(r.Password).SetVal(1)
			mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
			mock.ExpectIncr(r.IP).SetVal(1)
			mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
			mock.ExpectTxPipelineExec()

			jReq, err := json.Marshal(r)
			require.NoError(t, err)
			rdr := bytes.NewReader(jReq)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodPost, "/try", rdr)
			require.NoError(t, err)

			serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

			require.Equal(t, http.StatusOK, rr.Code)
		}

		r := app.Request{
			Login:    "login",
			Password: "password",
			IP:       "33.33.33.33",
		}

		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("1")
		mock.ExpectGet(r.Password).SetVal("1")
		mock.ExpectGet(r.IP).SetVal(strconv.Itoa(i))

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		jReq, err := json.Marshal(r)
		require.NoError(t, err)
		rdr := bytes.NewReader(jReq)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusTooManyRequests, rr.Code)
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}

//nolint:funlen
func TestBlacklist(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	testCfg := config.Config{
		Logger: config.LoggerConf{
			Level: "INFO",
		},
		Limiter: config.LimiterConf{
			N: 10,
			M: 90,
			K: 100,
		},
	}

	logger := logger.New(testCfg)
	tService := redislimiter.NewServiceMock(rdb, testCfg)

	serv := New(tService, testCfg, logger)
	t.Run("add to blacklist", func(t *testing.T) {
		// send request to add network to blacklist
		rq := app.Network{
			IP: "33.33.33.0/24",
		}
		// expect query to redisDB
		mock.ExpectSAdd("blacklist", rq.IP).SetVal(1)

		jReq, err := json.Marshal(rq)
		require.NoError(t, err)

		rdr := bytes.NewReader(jReq)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/blacklist", rdr)
		require.NoError(t, err)

		// execute query and expect statusOK
		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		// send request to check if the network was added to BL and
		// the following request returns statusTooManyRequests
		r := app.Request{
			Login:    "login",
			Password: "password",
			IP:       "33.33.33.33",
		}

		// expect query to redisDB
		mock.ExpectSMembers("blacklist").SetVal([]string{"33.33.33.0/24"})

		jReq, err = json.Marshal(r)
		require.NoError(t, err)
		rdr = bytes.NewReader(jReq)
		rr = httptest.NewRecorder()

		req, err = http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		// execute query and expect statusTooManyRequests
		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusTooManyRequests, rr.Code)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("remove from blacklist", func(t *testing.T) {
		rq := app.Network{
			IP: "33.33.33.0/24",
		}
		mock.ExpectSAdd("blacklist", rq.IP).SetVal(1)

		jReq, err := json.Marshal(rq)
		require.NoError(t, err)

		rdr := bytes.NewReader(jReq)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/blacklist", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		r := app.Request{
			Login:    "login",
			Password: "password",
			IP:       "33.33.33.33",
		}

		mock.ExpectSMembers("blacklist").SetVal([]string{"33.33.33.0/24"})

		jReq, err = json.Marshal(r)
		require.NoError(t, err)
		rdr = bytes.NewReader(jReq)
		rr = httptest.NewRecorder()

		req, err = http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusTooManyRequests, rr.Code)

		mock.ExpectSRem("blacklist", rq.IP).SetVal(1)

		jReq, err = json.Marshal(rq)
		require.NoError(t, err)

		rdr = bytes.NewReader(jReq)
		rr = httptest.NewRecorder()

		req, err = http.NewRequest(http.MethodDelete, "/blacklist", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("1")
		mock.ExpectGet(r.Password).SetVal("1")
		mock.ExpectGet(r.IP).SetVal("1")

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		jReq, err = json.Marshal(r)
		require.NoError(t, err)
		rdr = bytes.NewReader(jReq)
		rr = httptest.NewRecorder()

		req, err = http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}

//nolint:funlen
func TestWhitelist(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	testCfg := config.Config{
		Logger: config.LoggerConf{
			Level: "INFO",
		},
		Limiter: config.LimiterConf{
			N: 10,
			M: 90,
			K: 100,
		},
	}

	logger := logger.New(testCfg)
	tService := redislimiter.NewServiceMock(rdb, testCfg)

	serv := New(tService, testCfg, logger)

	t.Run("add to whitelist", func(t *testing.T) {
		// send request to add network to whitelist
		rq := app.Network{
			IP: "33.33.33.0/24",
		}
		// expect query to redisDB
		mock.ExpectSAdd("whitelist", rq.IP).SetVal(1)

		jReq, err := json.Marshal(rq)
		require.NoError(t, err)

		rdr := bytes.NewReader(jReq)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/whitelist", rdr)
		require.NoError(t, err)

		// execute query and expect statusOK
		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		// send request to check if the network was added to WL and
		// no matter how much there will be requests server returns statusOK
		r := app.Request{
			Login:    "login",
			Password: "password",
			IP:       "33.33.33.33",
		}

		jReq, err = json.Marshal(r)
		require.NoError(t, err)

		for i := 1; i < 100; i++ {
			mock.ExpectSMembers("blacklist").SetVal([]string{})
			mock.ExpectSMembers("whitelist").SetVal([]string{"33.33.33.0/24"})

			rdr := bytes.NewReader(jReq)
			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodPost, "/try", rdr)
			require.NoError(t, err)

			serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

			require.Equal(t, http.StatusOK, rr.Code)
		}

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("remove from whitelist", func(t *testing.T) {
		// send request to add network to whitelist
		rq := app.Network{
			IP: "33.33.33.0/24",
		}
		// expect query to redisDB
		mock.ExpectSAdd("whitelist", rq.IP).SetVal(1)

		jReq, err := json.Marshal(rq)
		require.NoError(t, err)

		rdr := bytes.NewReader(jReq)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/whitelist", rdr)
		require.NoError(t, err)

		// execute query and expect statusOK
		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		// send request to check if the network was added to WL and
		// no matter how much there will be requests server returns statusOK
		r := app.Request{
			Login:    "login",
			Password: "password",
			IP:       "33.33.33.33",
		}

		jReq, err = json.Marshal(r)
		require.NoError(t, err)

		for i := 1; i < 100; i++ {
			mock.ExpectSMembers("blacklist").SetVal([]string{})
			mock.ExpectSMembers("whitelist").SetVal([]string{"33.33.33.0/24"})

			rdr := bytes.NewReader(jReq)
			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodPost, "/try", rdr)
			require.NoError(t, err)

			serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

			require.Equal(t, http.StatusOK, rr.Code)
		}

		// send request to remove from WL
		mock.ExpectSRem("whitelist", rq.IP).SetVal(1)

		jReq, err = json.Marshal(rq)
		require.NoError(t, err)
		rdr = bytes.NewReader(jReq)
		rr = httptest.NewRecorder()

		req, err = http.NewRequest(http.MethodDelete, "/whitelist", rdr)
		require.NoError(t, err)

		// execute query and expect statusOK
		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		// send 10 requests to reach the limit and then make sure
		// that next requests aren't allowed
		jReq, err = json.Marshal(r)
		require.NoError(t, err)
		for i := 1; i < 10; i++ {
			mock.ExpectSMembers("blacklist").SetVal([]string{})
			mock.ExpectSMembers("whitelist").SetVal([]string{})
			mock.ExpectGet(r.Login).SetVal(strconv.Itoa(i))
			mock.ExpectGet(r.Password).SetVal(strconv.Itoa(i))
			mock.ExpectGet(r.IP).SetVal(strconv.Itoa(i))

			mock.ExpectTxPipeline()
			mock.ExpectIncr(r.Login).SetVal(1)
			mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
			mock.ExpectIncr(r.Password).SetVal(1)
			mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
			mock.ExpectIncr(r.IP).SetVal(1)
			mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
			mock.ExpectTxPipelineExec()

			rdr := bytes.NewReader(jReq)
			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodPost, "/try", rdr)
			require.NoError(t, err)

			serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

			require.Equal(t, http.StatusOK, rr.Code)
		}

		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("11")
		mock.ExpectGet(r.Password).SetVal("11")
		mock.ExpectGet(r.IP).SetVal("11")

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		rdr = bytes.NewReader(jReq)
		rr = httptest.NewRecorder()

		req, err = http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusTooManyRequests, rr.Code)
		err = mock.ExpectationsWereMet()
		require.NoError(t, err)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}

//nolint:funlen
func TestReset(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	testCfg := config.Config{
		Logger: config.LoggerConf{
			Level: "INFO",
		},
		Limiter: config.LimiterConf{
			N: 10,
			M: 90,
			K: 100,
		},
	}

	logger := logger.New(testCfg)
	tService := redislimiter.NewServiceMock(rdb, testCfg)

	serv := New(tService, testCfg, logger)

	t.Run("reset buckets for login", func(t *testing.T) {
		// send 10 requests to reach login's limit
		r := app.Request{
			Login:    "login",
			Password: "password",
			IP:       "33.33.34.33",
		}
		jReq, err := json.Marshal(r)
		require.NoError(t, err)

		for i := 1; i < 10; i++ {
			mock.ExpectSMembers("blacklist").SetVal([]string{})
			mock.ExpectSMembers("whitelist").SetVal([]string{})
			mock.ExpectGet(r.Login).SetVal(strconv.Itoa(i))
			mock.ExpectGet(r.Password).SetVal(strconv.Itoa(i))
			mock.ExpectGet(r.IP).SetVal(strconv.Itoa(i))

			mock.ExpectTxPipeline()
			mock.ExpectIncr(r.Login).SetVal(1)
			mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
			mock.ExpectIncr(r.Password).SetVal(1)
			mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
			mock.ExpectIncr(r.IP).SetVal(1)
			mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
			mock.ExpectTxPipelineExec()

			rdr := bytes.NewReader(jReq)
			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodPost, "/try", rdr)
			require.NoError(t, err)

			serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

			require.Equal(t, http.StatusOK, rr.Code)
		}

		// make sure that limit is reached and server returns statusTooManyRequest
		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("11")
		mock.ExpectGet(r.Password).SetVal("11")
		mock.ExpectGet(r.IP).SetVal("11")

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		rdr := bytes.NewReader(jReq)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusTooManyRequests, rr.Code)

		// send request to reset buckets for this login
		mock.ExpectDel(r.Login, r.IP).SetVal(1)

		rdr = bytes.NewReader(jReq)
		rr = httptest.NewRecorder()

		req, err = http.NewRequest(http.MethodPost, "/reset", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		// make sure that buckets for login is reset and server retursns StatusOK
		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("1")
		mock.ExpectGet(r.Password).SetVal("1")
		mock.ExpectGet(r.IP).SetVal("1")

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		rdr = bytes.NewReader(jReq)
		rr = httptest.NewRecorder()

		req, err = http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
	t.Run("test reset buckets for ip", func(t *testing.T) {
		// send 90 requests to reach ip's limit
		var i int
		for i = 1; i < 100; i++ {
			r := app.Request{
				Login:    "login" + strconv.Itoa(i),
				Password: "password" + strconv.Itoa(i),
				IP:       "33.33.33.33",
			}
			mock.ExpectSMembers("blacklist").SetVal([]string{})
			mock.ExpectSMembers("whitelist").SetVal([]string{})
			mock.ExpectGet(r.Login).SetVal("1")
			mock.ExpectGet(r.Password).SetVal("1")
			mock.ExpectGet(r.IP).SetVal(strconv.Itoa(i))

			mock.ExpectTxPipeline()
			mock.ExpectIncr(r.Login).SetVal(1)
			mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
			mock.ExpectIncr(r.Password).SetVal(1)
			mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
			mock.ExpectIncr(r.IP).SetVal(1)
			mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
			mock.ExpectTxPipelineExec()

			jReq, err := json.Marshal(r)
			require.NoError(t, err)
			rdr := bytes.NewReader(jReq)

			rr := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodPost, "/try", rdr)
			require.NoError(t, err)

			serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

			require.Equal(t, http.StatusOK, rr.Code)
		}

		// make sure that limit is reached and server returns statusTooManyRequest
		r := app.Request{
			Login:    "login",
			Password: "password",
			IP:       "33.33.33.33",
		}

		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("1")
		mock.ExpectGet(r.Password).SetVal("1")
		mock.ExpectGet(r.IP).SetVal(strconv.Itoa(i))

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		jReq, err := json.Marshal(r)
		require.NoError(t, err)
		rdr := bytes.NewReader(jReq)
		rr := httptest.NewRecorder()

		req, err := http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusTooManyRequests, rr.Code)

		// send request to reset buckets for ip
		mock.ExpectDel(r.Login, r.IP).SetVal(1)

		rdr = bytes.NewReader(jReq)
		rr = httptest.NewRecorder()

		req, err = http.NewRequest(http.MethodPost, "/reset", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		// make sure that buckets for ip is reset and server retursns StatusOK
		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("1")
		mock.ExpectGet(r.Password).SetVal("1")
		mock.ExpectGet(r.IP).SetVal("1")

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		rdr = bytes.NewReader(jReq)
		rr = httptest.NewRecorder()

		req, err = http.NewRequest(http.MethodPost, "/try", rdr)
		require.NoError(t, err)

		serv.server.Handler.ServeHTTP(rr, req.WithContext(context.Background()))

		require.Equal(t, http.StatusOK, rr.Code)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}
