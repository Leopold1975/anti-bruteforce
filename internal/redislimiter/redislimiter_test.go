//nolint:dupl
package redislimiter

import (
	"context"
	"testing"
	"time"

	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/app"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/config"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/require"
)

func TestTryAuth(t *testing.T) {
	ctx := context.Background()

	testCfg := config.Config{
		Limiter: config.LimiterConf{
			N: 10,
			M: 100,
			K: 100,
		},
	}
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()

	tService := NewServiceMock(rdb, testCfg)

	r := app.Request{
		Login:    "login",
		Password: "password",
		IP:       "33.33.33.33",
	}

	t.Run("IP in backlist", func(t *testing.T) {
		mock.ExpectSMembers("blacklist").SetVal([]string{"33.33.33.0/24"})

		b, err := tService.TryAuth(ctx, r)
		require.NoError(t, err)
		require.False(t, b)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
	t.Run("IP in backlist", func(t *testing.T) {
		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{"33.33.33.0/24"})

		b, err := tService.TryAuth(ctx, r)
		require.NoError(t, err)
		require.True(t, b)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
	t.Run("IP  nor in backlistnor in whitelist", func(t *testing.T) {
		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("0")
		mock.ExpectGet(r.Password).SetVal("0")
		mock.ExpectGet(r.IP).SetVal("0")

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		b, err := tService.TryAuth(ctx, r)
		require.NoError(t, err)
		require.True(t, b)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
	t.Run("IP  nor in backlistnor in whitelist and limit is reached", func(t *testing.T) {
		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("10")
		mock.ExpectGet(r.Password).SetVal("10")
		mock.ExpectGet(r.IP).SetVal("10")

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		b, err := tService.TryAuth(ctx, r)
		require.NoError(t, err)
		require.False(t, b)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
	t.Run("too many attempts for password", func(t *testing.T) {
		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("0")
		mock.ExpectGet(r.Password).SetVal("100")
		mock.ExpectGet(r.IP).SetVal("10")

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		b, err := tService.TryAuth(ctx, r)
		require.NoError(t, err)
		require.False(t, b)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
	t.Run("too many attempts for ip", func(t *testing.T) {
		mock.ExpectSMembers("blacklist").SetVal([]string{})
		mock.ExpectSMembers("whitelist").SetVal([]string{})
		mock.ExpectGet(r.Login).SetVal("0")
		mock.ExpectGet(r.Password).SetVal("0")
		mock.ExpectGet(r.IP).SetVal("100")

		mock.ExpectTxPipeline()
		mock.ExpectIncr(r.Login).SetVal(1)
		mock.ExpectExpire(r.Login, time.Minute).SetVal(false)
		mock.ExpectIncr(r.Password).SetVal(1)
		mock.ExpectExpire(r.Password, time.Minute).SetVal(false)
		mock.ExpectIncr(r.IP).SetVal(1)
		mock.ExpectExpire(r.IP, time.Minute).SetVal(false)
		mock.ExpectTxPipelineExec()

		b, err := tService.TryAuth(ctx, r)
		require.NoError(t, err)
		require.False(t, b)

		err = mock.ExpectationsWereMet()
		require.NoError(t, err)
	})
}
