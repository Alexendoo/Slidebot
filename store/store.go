package store

import (
	"time"

	"github.com/boltdb/bolt"
)

var (
	BucketLastFM = []byte("last.fm")
)

var DB *bolt.DB

func Open(path string) (err error) {
	DB, err = bolt.Open(path, 0600, &bolt.Options{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return err
	}

	return DB.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists(BucketLastFM)

		return nil
	})
}

func Get(bucket []byte, key string) (string, error) {
	tx, err := DB.Begin(false)
	if err != nil {
		return "", err
	}

	return string(tx.Bucket(bucket).Get([]byte(key))), nil
}

func Set(bucket []byte, key, value string) error {
	tx, err := DB.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.Bucket(bucket).Put([]byte(key), []byte(value))
	if err != nil {
		return err
	}

	return tx.Commit()
}
