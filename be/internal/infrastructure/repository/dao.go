package dao

import "time"

type Dao interface {
	Save(key string, startTime time.Time, users string) error
	Get(key string, startTime time.Time) (string, error)

	SaveLogs(logs string) error
	GetLogs() (string, error)
	GetKey() string
	SaveBot(startTime time.Time, bot string) error
	GetBot(startTime time.Time) (string, error)
	DeleteBot(startTime time.Time) error
}
