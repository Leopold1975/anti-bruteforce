package redislimiter

import (
	"context"
	"errors"
	"time"

	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/app"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/config"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/pkg/ip"
	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	N int64 // attempts for login
	M int64 // attempts for password
	K int64 // attempts for ip
}

type ABFService struct {
	rdb     *redis.Client
	limiter Limiter
}

func New(cfg config.Config) *ABFService {
	return &ABFService{
		rdb: redis.NewClient(&redis.Options{
			Addr:     cfg.RedisDB.Addr,
			Password: cfg.RedisDB.Password,
			DB:       cfg.RedisDB.DB,
		}),
		limiter: Limiter{
			N: cfg.Limiter.N,
			M: cfg.Limiter.M,
			K: cfg.Limiter.K,
		},
	}
}

func (a *ABFService) Shutdown(ctx context.Context) error {
	return a.rdb.Shutdown(ctx).Err()
}

func (a *ABFService) TryAuth(ctx context.Context, r app.Request) (bool, error) {
	mem, err := a.rdb.SMembersMap(ctx, "blacklist").Result()
	if err != nil {
		return false, err
	}
	for k := range mem {
		ok, err := ip.BelongsToNetwork(k, r.IP)
		if err != nil {
			return false, err
		}
		if ok {
			return false, nil
		}
	}

	mem, err = a.rdb.SMembersMap(ctx, "whitelist").Result()
	if err != nil {
		return false, err
	}
	for k := range mem {
		ok, err := ip.BelongsToNetwork(k, r.IP)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	currentLogin, err := a.rdb.Get(ctx, r.Login).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, err
	}

	currentPassword, err := a.rdb.Get(ctx, r.Password).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, err
	}

	currentIP, err := a.rdb.Get(ctx, r.IP).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, err
	}

	pipe := a.rdb.TxPipeline()

	pipe.Incr(ctx, r.Login)
	pipe.Expire(ctx, r.Login, time.Minute)

	pipe.Incr(ctx, r.Password)
	pipe.Expire(ctx, r.Password, time.Minute)

	pipe.Incr(ctx, r.IP)
	pipe.Expire(ctx, r.IP, time.Minute)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	return a.limiter.CheckRateLimitLogin(currentLogin) && a.limiter.CheckRateLimitPassword(currentPassword) &&
			a.limiter.CheckRateLimitIP(currentIP),
		nil
}

func (a *ABFService) ResetBuckets(ctx context.Context, login string, ipAddr string) error {
	_, err := a.rdb.Del(ctx, login, ipAddr).Result()
	if err != nil {
		return err
	}
	return nil
}

func (a *ABFService) AddToBlacklist(ctx context.Context, network app.Network) error {
	_, err := a.rdb.SAdd(ctx, "blacklist", network.IP).Result()
	if err != nil {
		return err
	}
	return nil
}

func (a *ABFService) RemoveFromBlacklist(ctx context.Context, network app.Network) error {
	_, err := a.rdb.SRem(ctx, "blacklist", network.IP).Result()
	if err != nil {
		return err
	}
	return nil
}

func (a *ABFService) AddToWhitelist(ctx context.Context, network app.Network) error {
	_, err := a.rdb.SAdd(ctx, "whitelist", network.IP).Result()
	if err != nil {
		return err
	}
	return nil
}

func (a *ABFService) RemoveFromWhitelist(ctx context.Context, network app.Network) error {
	_, err := a.rdb.SRem(ctx, "whitelist", network.IP).Result()
	if err != nil {
		return err
	}
	return nil
}

func (l *Limiter) CheckRateLimitLogin(current int64) bool {
	return current < l.N
}

func (l *Limiter) CheckRateLimitPassword(current int64) bool {
	return current < l.M
}

func (l *Limiter) CheckRateLimitIP(current int64) bool {
	return current < l.K
}
