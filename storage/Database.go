package Storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type URLDB struct {
	DB          *pgxpool.Pool
	Redis       *redis.Client
	Ctx         context.Context
	Mut         sync.Mutex
	Wg          sync.WaitGroup
	insertQueue chan urlpair
}
type urlpair struct {
	short string
	long  string
}

func ConnectToDB(pgconn string, redisAddr string) (*URLDB, error) {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, pgconn)
	if err != nil {
		return nil, fmt.Errorf("sqlite error: %w", err)
	}
	// Force a connection check by pinging
	err = db.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("sqlite ping error: %w", err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis connection error: %w", err)
	}
	URLDB := &URLDB{
		DB:          db,
		Redis:       rdb,
		Ctx:         ctx,
		Mut:         sync.Mutex{},
		Wg:          sync.WaitGroup{},
		insertQueue: make(chan urlpair, 200),
	}
	URLDB.startInsertWorkers(5)
	return URLDB, nil
}

func (URLDB *URLDB) Close() error {
	close(URLDB.insertQueue)
	URLDB.Wg.Wait()
	URLDB.DB.Close()
	err := URLDB.Redis.Close()
	if err != nil {
		return err
	}
	return nil
}

func (URLDB *URLDB) createURLtable() error {
	_, err := URLDB.DB.Exec(URLDB.Ctx, `
		CREATE TABLE IF NOT EXISTS urls(
			id SERIAL PRIMARY KEY,
			short TEXT UNIQUE NOT NULL,
			long TEXT NOT NULL
		);`,
	)
	if err != nil {
		return fmt.Errorf("URL table creation error: %w", err)
	}
	_, err = URLDB.DB.Exec(URLDB.Ctx, `CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short);`)
	if err != nil {
		return fmt.Errorf("Index creation error: %w", err)
	}
	return nil
}

func (URLDB *URLDB) createUsertable() error {
	_, err := URLDB.DB.Exec(URLDB.Ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("User table creation error: %w", err)
	}
	return nil
}

func (URLDB *URLDB) CreateTables() error {
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

func (URLDB *URLDB) startInsertWorkers(n int) {
	for range n {
		go func() {
			for pair := range URLDB.insertQueue {
				_, err := URLDB.DB.Exec(URLDB.Ctx, "INSERT INTO urls (short,long) VALUES ($1,$2) ON CONFLICT (short) DO NOTHING", pair.short, pair.long)
				if err != nil {
					fmt.Printf("Insert error: %v\n", err)
				}
				URLDB.Wg.Done()
			}
		}()
	}
}

func (URLDB *URLDB) SaveURL(short, long string) error {
	select {
	case URLDB.insertQueue <- urlpair{short: short, long: long}:
		URLDB.Wg.Add(1)
		return nil
	default:
		return fmt.Errorf("insert queue is full")
	}
}

func (URLDB *URLDB) GetURL(short string) (string, error) {
	// 1. Try Redis first (fast path)
	if val, err := URLDB.Redis.Get(URLDB.Ctx, short).Result(); err != redis.Nil {
		return val, nil
	}

	// 2. Fallback to SQLite (cold path)
	var long string
	err := URLDB.DB.QueryRow(URLDB.Ctx, "SELECT long FROM urls WHERE short = $1", short).Scan(&long)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("not found")
		}
		return "", fmt.Errorf("sqlite fetch error: %w", err)
	}

	// 3. Set in Redis synchronously with timeout context (avoids goroutine storm)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err = URLDB.Redis.Set(ctx, short, long, time.Hour*24).Err()
	if err != nil {
		fmt.Printf("redis set error: %v\n", err) // log, but don’t fail
	}

	return long, nil
}

func (URLDB *URLDB) DeleteURL(short string) error {
	_, err := URLDB.DB.Exec(URLDB.Ctx, "DELETE FROM urls WHERE short = $1", short)
	if err != nil {
		return fmt.Errorf("Errors Deleting URL: %w", err)
	}
	return URLDB.Redis.Del(URLDB.Ctx, short).Err()
}

func (URLDB *URLDB) EditURL(short string, newlong string) error {
	_, err := URLDB.DB.Exec(URLDB.Ctx, "UPDATE urls SET long = $1 WHERE short = $2", newlong, short)
	if err != nil {
		return fmt.Errorf("Error updating urls: %w", err)
	}
	return URLDB.Redis.Set(URLDB.Ctx, short, newlong, time.Hour*24).Err()
}

func (URLDB *URLDB) CheckShortURLExists(short string) (bool, error) {
	_, err := URLDB.Redis.Get(URLDB.Ctx, short).Result()
	if err != redis.Nil {
		return true, nil
	}
	var exists bool
	err = URLDB.DB.QueryRow(URLDB.Ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE short = $1)", short).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("Error checking if short url exists: %w", err)
	}
	return exists, nil
}
