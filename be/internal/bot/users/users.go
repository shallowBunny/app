package users

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	dao "github.com/shallowBunny/app/be/internal/infrastructure/repository"
)

type UserInfo struct {
	Notifications bool
	Deleted       bool
	NewUser       bool
	MagicButton1  int
	MagicButton2  int
}

type Users struct {
	usersInfo map[int64]*UserInfo
	dao       dao.Dao
	startTime time.Time
}

func New(dao dao.Dao, startTime time.Time) Users {
	usersInfo := make(map[int64]*UserInfo)
	usersString, err := dao.Get("users", startTime)
	if err != nil {
		if err.Error() == "redis: nil" {
			log.Warn().Msg("empty dao.Users")
		} else {
			log.Error().Msg(err.Error())
		}
	} else {
		err = json.Unmarshal([]byte(usersString), &usersInfo)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}
	res := Users{
		usersInfo: usersInfo,
		dao:       dao,
		startTime: startTime,
	}
	return res
}
func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

func (u *Users) SaveUsers() error {
	bytes, err := json.Marshal(u.usersInfo)
	if err != nil {
		panic(err)
	}
	s := string(bytes)
	res, err := PrettyString(s)
	if err != nil {
		log.Error().Msg(err.Error())
	}
	log.Trace().Msg(res)
	return u.dao.Save("users", u.startTime, s)
}

func (u Users) HasUserNotifications(userId int64) (bool, error) {
	_, ok := u.usersInfo[userId]
	if !ok {
		log.Warn().Msg(fmt.Sprintf("HasUserNotifications %d", userId))
		return true, errors.New("HasUserNotifications on unknown user")
	}
	return u.usersInfo[userId].Notifications, nil
}

func (u *Users) DeleteUser(userId int64) error {
	_, ok := u.usersInfo[userId]
	if !ok {
		return errors.New("trying to delete unknown user")
	}
	u.usersInfo[userId].Deleted = true
	return u.SaveUsers()
}

func (u *Users) SetNotificationsUser(userId int64, notification bool) error {
	_, ok := u.usersInfo[userId]
	if !ok {
		return errors.New("trying to set notifications on unknown user")
	}
	u.usersInfo[userId].Notifications = notification
	return u.SaveUsers()
}

func (u *Users) SetUserAsNew(userId int64) error {
	_, ok := u.usersInfo[userId]
	if !ok {
		return errors.New("trying to set SetUserAsNew on unknown user")
	}
	u.usersInfo[userId].NewUser = true
	return u.SaveUsers()
}

func (u *Users) GetMagicButtons(userId int64) (int, int) {
	info, ok := u.usersInfo[userId]
	if !ok {
		return 0, 1
	}
	if info.MagicButton1 == info.MagicButton2 {
		return 0, 1
	}
	return info.MagicButton1, info.MagicButton2
}

func (u *Users) UpdateMagicButtons(userId int64, room int, nbRooms int) error {
	_, ok := u.usersInfo[userId]
	if !ok {
		return errors.New("trying to UpdateMagicButtons on unknown user")
	}
	if u.usersInfo[userId].MagicButton1 == room {
		i := (room + 1) % nbRooms
		if i == u.usersInfo[userId].MagicButton2 {
			i = (i + 1) % nbRooms
		}
		u.usersInfo[userId].MagicButton1 = i
	}
	if u.usersInfo[userId].MagicButton2 == room {
		i := (room + 1) % nbRooms
		if i == u.usersInfo[userId].MagicButton1 {
			i = (i + 1) % nbRooms
		}
		u.usersInfo[userId].MagicButton2 = i
	}
	return u.SaveUsers()
}

func (u *Users) DoesUserExists(userId int64) bool {
	_, ok := u.usersInfo[userId]
	if ok {
		if u.usersInfo[userId].Deleted {
			u.usersInfo[userId].Deleted = false
			u.SaveUsers()
		}
		return true
	}
	u.usersInfo[userId] = &UserInfo{
		Notifications: false,
		Deleted:       false,
	}
	err := u.SaveUsers()
	if err != nil {
		log.Error().Msg(err.Error())
	}
	return false
}

func (u Users) UsersWithNotifications() []int64 {
	res := []int64{}
	for k, v := range u.usersInfo {
		if !v.Deleted && v.Notifications {
			res = append(res, k)
		}
	}
	return res
}

func (u Users) UsersStats() (int, int, int, int) {
	newUsers := 0
	totalUsers := 0
	Deleted := 0
	Notifications := 0
	for _, v := range u.usersInfo {
		if v.NewUser {
			newUsers++
		}
		if !v.Deleted {
			totalUsers++
		}
		if v.Deleted {
			Deleted++
			totalUsers--
		}
		if v.Notifications {
			Notifications++
		}
	}
	return newUsers, totalUsers, Deleted, Notifications
}
