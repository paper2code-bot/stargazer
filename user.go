package main

import (
	"time"

	"github.com/sirupsen/logrus"
)

func execUserRoutine(dbClient *databaseClient, ghClient *githubClient, expirationDelay int) error {
	logrus.Info("user routine: get stargazers from database")
	ss, err := dbClient.getStargazers()
	if err != nil {
		return err
	}

	logrus.Infof("user routine: iterate over %d stargazers", len(ss))
	for i := range ss {
		rawUser := ss[i].Data["user"].(object)
		login := rawUser["login"].(string)

		u, err := dbClient.getUser(login)
		if err != nil {
			return err
		}
		if u == nil || (u.Expire.Before(time.Now()) && expirationDelay > 0) {
			logrus.Infof("user routine: get user %s from Github (%d/%d)", login, i+1, len(ss))
			o, err := ghClient.getUser(login)
			if err != nil {
				return err
			}

			expire := time.Now().Add(time.Hour * time.Duration(expirationDelay))
			if u == nil {
				logrus.Infof("user routine: insert user %s in database", login)
				if err := dbClient.insertUser(&user{
					Expire: expire,
					Login:  login,
					Data:   o,
				}); err != nil {
					return err
				}
			} else {
				u.Expire = expire
				u.Data = o
				logrus.Infof("user routine: update user %s in database", login)
				if err := dbClient.updateUser(u); err != nil {
					return err
				}
			}

			logrus.Info("user routine: wait 50ms")
			time.Sleep(50 * time.Millisecond)
		}
	}

	return nil
}