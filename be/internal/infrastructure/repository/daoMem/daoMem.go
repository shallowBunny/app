package DaoMem

import (
	"errors"
	"time"
)

type DaoMem struct {
}

func (d DaoMem) Save(key string, startTime time.Time, users string) error {
	return nil
}
func (d DaoMem) Get(key string, startTime time.Time) (string, error) {
	return "", errors.New("memory")
}

func (d DaoMem) SaveBot(startTime time.Time, bot string) error {
	return nil
}
func (d DaoMem) GetBot(startTime time.Time) (string, error) {
	return "", errors.New("memory")
}
func (d DaoMem) DeleteBot(startTime time.Time) error {
	return nil
}

func (d DaoMem) SaveUsers(usersId []int64) error {
	return nil
}

func (d DaoMem) GetUsers() ([]int64, error) {
	res := []int64{}
	return res, nil
}

func (d DaoMem) SaveLogs(logs string) error {
	return nil
}
func (d DaoMem) GetLogs() (string, error) {
	return "", nil
}

func New() *DaoMem {
	return &DaoMem{}

}

func (d DaoMem) GetKey() string {
	return "memory"
}
