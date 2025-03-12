package DaoDb

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var Ttl = 2 * 7 * 24 * time.Hour

type DaoDb struct {
	redisclient *redis.Client
	redisKey    string
}

func (d DaoDb) Save(key string, startTime time.Time, users string) error {
	return d.redisclient.Set(context.Background(), key+"-"+d.redisKey+"-"+startTime.Format("Mon-02-Jan-2006"), users, Ttl).Err()
}
func (d DaoDb) Get(key string, startTime time.Time) (string, error) {
	return d.redisclient.Get(context.Background(), key+"-"+d.redisKey+"-"+startTime.Format("Mon-02-Jan-2006")).Result()
}

func (d DaoDb) SaveBot(startTime time.Time, bot string) error {
	return d.redisclient.Set(context.Background(), "bot-"+d.redisKey+"-"+startTime.Format("Mon-02-Jan-2006"), bot, Ttl).Err()
}
func (d DaoDb) GetBot(startTime time.Time) (string, error) {
	return d.redisclient.Get(context.Background(), "bot-"+d.redisKey+"-"+startTime.Format("Mon-02-Jan-2006")).Result()
}

func (d DaoDb) DeleteBot(startTime time.Time) error {
	_, err := d.redisclient.Del(context.Background(), "bot-"+d.redisKey+"-"+startTime.Format("Mon-02-Jan-2006")).Result()
	return err
}

func (d DaoDb) SaveUsers(usersId []int64) error {
	res := ""
	for i, v := range usersId {
		if i != 0 {
			res += " "
		}
		res += strconv.Itoa(int(v))
	}
	return d.redisclient.Set(context.Background(), "users-"+d.redisKey, res, Ttl).Err()
}

func (d DaoDb) GetUsers() ([]int64, error) {
	res := []int64{}
	val, err := d.redisclient.Get(context.Background(), "users-"+d.redisKey).Result()
	if err != nil {
		return res, err
	} else {
		users := strings.Split(val, " ")
		for _, v := range users {
			i, err := strconv.Atoi(v)
			if err != nil {
				return res, err
			}
			res = append(res, int64(i))
		}
	}
	return res, nil
}

func (d DaoDb) SaveLogs(logs string) error {
	return d.redisclient.Set(context.Background(), "logs-"+d.redisKey, logs, Ttl).Err()

}
func (d DaoDb) GetLogs() (string, error) {
	return d.redisclient.Get(context.Background(), "logs-"+d.redisKey).Result()
}

func New(apiToken string, redisclient *redis.Client) *DaoDb {
	h := sha256.New()
	h.Write([]byte(apiToken))
	bs := h.Sum(nil)
	res := &DaoDb{
		redisKey:    fmt.Sprintf("%x", bs),
		redisclient: redisclient,
	}
	return res
}

func (d DaoDb) GetKey() string {
	return d.redisKey
}

func (d DaoDb) SaveHset24Hours(key, ip string) (int64, error) {
	ctx := context.Background()
	timestamp := time.Now().Unix()

	// Start a transaction
	pipeline := d.redisclient.TxPipeline()

	// Add or update the IP timestamp
	pipeline.HSet(ctx, key, ip, timestamp)

	// Get all IPs and timestamps
	entries, err := d.redisclient.HGetAll(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Remove outdated entries (older than 24 hours)
	threshold := timestamp - 24*3600
	for storedIP, storedTimeStr := range entries {
		storedTime, err := strconv.ParseInt(storedTimeStr, 10, 64)
		if err != nil {
			continue // Skip invalid entries
		}
		if storedTime < threshold {
			pipeline.HDel(ctx, key, storedIP)
		}
	}

	// Execute pipeline
	_, err = pipeline.Exec(ctx)
	if err != nil {
		return 0, err
	}

	// Get the updated count of elements
	count, err := d.redisclient.HLen(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return count, nil
}
