package models

import (
	"context"
	"github.com/revel/revel"
	"time"
)

func SetDataInRedis(key string, value []byte, timeLive time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := RDB.Set(ctx, key, value, timeLive).Err()
	if err != nil {
		return err
	}
	revel.AppLog.Debug("установили данные " + string(value))
	return nil
}

func GetDataFromRedis(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return RDB.Get(ctx, key).Bytes()
}

func DeleteDataFromRedis(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := RDB.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	revel.AppLog.Debug("удалили по ключу " + key)
	return nil
}
