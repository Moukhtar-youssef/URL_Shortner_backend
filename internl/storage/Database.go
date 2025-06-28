package Storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type URLDB struct {
	DB          *pgxpool.Pool
	Redis       *redis.Client
	Cache       *cache
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

	cache := newCache()

	URLDB := &URLDB{
		DB:          db,
		Redis:       rdb,
		Cache:       cache,
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

	URLDB.Cache.Stop()

	err := URLDB.Redis.Close()
	if err != nil {
		return err
	}

	return nil
}

func (URLDB *URLDB) createURLtable() error {
	_, err := URLDB.DB.Exec(URLDB.Ctx, `
	CREATE TABLE IF NOT EXISTS urls(
        id BIGSERIAL PRIMARY KEY,
        short VARCHAR(10) UNIQUE NOT NULL,
        long TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`,
	)
	if err != nil {
		return fmt.Errorf("URL table creation error: %w", err)
	}

	_, err = URLDB.DB.Exec(URLDB.Ctx, `
   		CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_urls_short ON urls(short);
	`)
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

func (db *URLDB) startInsertWorkers(n int) {
	for i := range n {
		go func(workerID int) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Worker %d recovered from panic: %v", workerID, r)
				}
			}()

			for pair := range db.insertQueue {
				// Use context with timeout for database operations
				ctx, cancel := context.WithTimeout(db.Ctx, 5*time.Second)
				defer cancel()

				// 1. First try Redis
				redisShort := fmt.Sprintf("URL:%s", pair.short)
				err := db.Redis.Set(ctx, redisShort, pair.long, 24*time.Hour).Err()
				if err != nil {
					log.Printf("Worker %d: Redis insert error for %s: %v", workerID, pair.short, err)
					db.Wg.Done()
					continue
				}

				// 2. Then PostgreSQL with retry logic
				maxRetries := 3
				for attempt := 1; attempt <= maxRetries; attempt++ {
					_, err = db.DB.Exec(ctx, `
                        INSERT INTO urls (short, long) 
                        VALUES ($1, $2) 
                        ON CONFLICT (short) DO NOTHING`,
						pair.short, pair.long)

					if err == nil {
						break // Success
					}

					if attempt < maxRetries {
						log.Printf("Worker %d: Insert attempt %d failed for %s: %v",
							workerID, attempt, pair.short, err)
						time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
					} else {
						log.Printf("Worker %d: Final insert failed for %s: %v",
							workerID, pair.short, err)
					}
				}

				db.Wg.Done()
			}
		}(i)
	}
}

func (URLDB *URLDB) SaveURL(short, long string) error {
	URLDB.Cache.Set(short, long, 5*time.Minute)
	select {
	case URLDB.insertQueue <- urlpair{short: short, long: long}:
		URLDB.Wg.Add(1)
		return nil
	default:
		return fmt.Errorf("insert queue is full")
	}
}

func (URLDB *URLDB) GetURL(short string) (string, error) {
	longcache, exists := URLDB.Cache.Get(short)
	if exists {
		return longcache, nil
	}
	redisShort := fmt.Sprintf("URL:%s", short)
	if val, err := URLDB.Redis.Get(URLDB.Ctx, redisShort).Result(); err != redis.Nil {
		return val, nil
	}

	var long string
	err := URLDB.DB.QueryRow(URLDB.Ctx, "SELECT long FROM urls WHERE short = $1", short).Scan(&long)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("not found")
		}
		return "", fmt.Errorf("sqlite fetch error: %w", err)
	}

	go func() {
		URLDB.Cache.Set(short, long, 5*time.Minute)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		err = URLDB.Redis.Set(ctx, short, long, time.Hour*24).Err()
		if err != nil {
			fmt.Printf("redis set error: %v\n", err)
		}
	}()
	return long, nil
}

func (URLDB *URLDB) DeleteURL(short string) error {
	URLDB.Cache.Delete(short)
	redisShort := fmt.Sprintf("URL:%s", short)

	_, err := URLDB.DB.Exec(URLDB.Ctx, "DELETE FROM urls WHERE short = $1", short)
	if err != nil {
		return fmt.Errorf("Errors Deleting URL: %w", err)
	}

	return URLDB.Redis.Del(URLDB.Ctx, redisShort).Err()
}

func (URLDB *URLDB) EditURL(short string, newlong string) error {
	URLDB.Cache.Delete(short)
	URLDB.Cache.Set(short, newlong, 5*time.Minute)
	redisShort := fmt.Sprintf("URL:%s", short)

	_, err := URLDB.DB.Exec(URLDB.Ctx, "UPDATE urls SET long = $1 WHERE short = $2", newlong, short)
	if err != nil {
		return fmt.Errorf("Error updating urls: %w", err)
	}

	return URLDB.Redis.Set(URLDB.Ctx, redisShort, newlong, time.Hour*24).Err()
}

func (URLDB *URLDB) CheckShortURLExists(short string) (bool, error) {
	redisShort := fmt.Sprintf("URL:%s", short)

	Exists, err := URLDB.Redis.Exists(URLDB.Ctx, redisShort).Result()
	if err != nil {
		return true, fmt.Errorf("Error checking if short url exists: %w", err)
	}

	if Exists != 0 {
		return true, nil
	} else if Exists == 0 {
		return false, nil
	}

	var exists bool
	err = URLDB.DB.QueryRow(URLDB.Ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE short = $1)", short).Scan(&exists)
	if err != nil {
		return true, fmt.Errorf("Error checking if short url exists: %w", err)
	}

	return exists, nil
}
