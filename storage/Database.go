package Storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
)

type URLDB struct {
	DB    *sql.DB
	Redis *redis.Client
	Ctx   context.Context
	Mut   sync.Mutex
	Wg    sync.WaitGroup
}

func ConnectToDB(sqlitePath string, redisAddr string) (*URLDB, error) {
	db, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return nil, fmt.Errorf("sqlite error: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis connection error: %w", err)
	}
	ctx := context.Background()
	return &URLDB{
		DB:    db,
		Redis: rdb,
		Ctx:   ctx,
		Mut:   sync.Mutex{},
		Wg:    sync.WaitGroup{},
	}, nil
}

func (URLDB *URLDB) Close() error {
	err := URLDB.DB.Close()
	if err != nil {
		return err
	}
	err = URLDB.Redis.Close()
	if err != nil {
		return err
	}
	return nil
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

func (URLDB *URLDB) createUsertable() error {
	_, err := URLDB.DB.Exec(`CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT NOT NULL,
		Created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		Updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`,
	)
	if err != nil {
		return fmt.Errorf("User table creation error: %w", err)
	}
	return nil
}

func (URLDB *URLDB) CreateTable() error {
	err := URLDB.createURLtable()
	if err != nil {
		return err
	}
	err = URLDB.createUsertable()
	if err != nil {
		return err
	}
	return nil
}

func (URLDB *URLDB) SaveURL(short, long string) error {
	URLDB.Mut.Lock()
	defer URLDB.Mut.Unlock()
	_, err := URLDB.DB.Exec("INSERT INTO urls (short,long) VALUES (?,?)", short, long)
	if err != nil {
		return fmt.Errorf("Error inserting URLS: %w", err)
	}
	return URLDB.Redis.Set(URLDB.Ctx, short, long, time.Hour*24).Err()
}

func (URLDB *URLDB) GetURL(short string) (string, error) {
	URLDB.Mut.Lock()
	defer URLDB.Mut.Unlock()
	val, err := URLDB.Redis.Get(URLDB.Ctx, short).Result()
	if err != redis.Nil {
		return val, nil
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
	URLDB.Mut.Lock()
	defer URLDB.Mut.Unlock()
	_, err := URLDB.DB.Exec("DELETE FROM urls WHERE short = ?", short)
	if err != nil {
		return fmt.Errorf("Errors Deleting URL: %w", err)
	}
	return URLDB.Redis.Del(URLDB.Ctx, short).Err()
}

func (URLDB *URLDB) EditURL(short string, newlong string) error {
	URLDB.Mut.Lock()
	defer URLDB.Mut.Unlock()
	_, err := URLDB.DB.Exec("UPDATE urls SET long = ? WHERE short = ?", newlong, short)
	if err != nil {
		return fmt.Errorf("Error updating urls: %w", err)
	}
	return URLDB.Redis.Set(URLDB.Ctx, short, newlong, time.Hour*24).Err()
}

func (URLDB *URLDB) CheckShortURLExists(short string) (bool, error) {
	URLDB.Mut.Lock()
	defer URLDB.Mut.Unlock()
	_, err := URLDB.Redis.Get(URLDB.Ctx, short).Result()
	if err != redis.Nil {
		return true, nil
	}
	var exists bool
	err = URLDB.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM urls WHERE short = ? LIMIT 1)", short).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("Error checking if short url exists: %w", err)
	}
	return exists, nil
}
