package Storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

type URLDB struct {
	DB    *sql.DB
	Redis *redis.Client
	Ctx   context.Context
}

func ConnectToDB(sqlitePath string, redisAddr string) (*URLDB, error) {
	db, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return nil, fmt.Errorf("sqlite error: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	ctx := context.Background()
	return &URLDB{
		DB:    db,
		Redis: rdb,
		Ctx:   ctx,
	}, nil
}
func (URLDB *URLDB) createURLtable() error {
	_, err := URLDB.DB.Exec(`CREATE TABLE IF NOT EXISTS urls(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		short TEXT UNIQUE,
		long TEXT NOT NULL);`,
	)
	if err != nil {
		return fmt.Errorf("URL table creation error: %w", err)
	}
	return nil
}
func (URLDB *URLDB) CreateTable() error {
	err := URLDB.createURLtable()
	if err != nil {
		return err
	}
	return nil
}
func (URLDB *URLDB) SaveURL(short, long string) error {
	_, err := URLDB.DB.Exec("INSERT INTO urls (short,long) VALUES (?,?)", short, long)
	if err != nil {
		return fmt.Errorf("Error inserting URLS: %w", err)
	}
	return URLDB.Redis.Set(URLDB.Ctx, short, long, time.Hour*24).Err()
}
func (URLDB *URLDB) GetURL(short string) (string, error) {
	val, err := URLDB.Redis.Get(URLDB.Ctx, short).Result()
	if err != redis.Nil {
		return val, nil
	}
	if err != nil {
		return "", fmt.Errorf("Error fetching LongURL from redis: %w", err)
	}
	var long string
	err = URLDB.DB.QueryRow("SELECT long FROM urls WHERE short = ?", short).Scan(&long)
	if err != nil {
		return "", fmt.Errorf("Error fetching longURL from sqlite: %w", err)
	}
	URLDB.Redis.Set(URLDB.Ctx, short, long, time.Hour*24)
	return long, nil
}
func (URLDB *URLDB) DeleteURL(short string) error {
	_, err := URLDB.DB.Exec("DELETE FROM urls WHERE short = ?", short)
	if err != nil {
		return fmt.Errorf("Errors Deleting URL: %w", err)
	}
	return URLDB.Redis.Del(URLDB.Ctx, short).Err()
}
