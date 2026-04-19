package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type DB interface {
	AddFeed(userJID, url, title string) (int64, error)
	RemoveFeed(userJID, url string) (int64, error)
	ListFeeds(userJID string) ([]Feed, error)
	GetAllFeeds() ([]Feed, error)
	UpdateFeedLastGUID(id int64, guid string) error
	SetFeedDisabled(userJID, url string, disabled bool) (int64, error)
	IncrementErrorCount(id int64) (int, error)
	ResetErrorCount(id int64) error
	Close() error
}

type Feed struct {
	ID         int64
	UserJID    string
	URL        string
	Title      string
	LastGUID   string
	LastPoll   time.Time
	Disabled   int
	ErrorCount int
}

func NewDB(cfg DatabaseConfig) (DB, error) {
	switch cfg.Type {
	case "postgres":
		return newPostgresDB(cfg)
	case "sqlite":
		fallthrough
	default:
		return newSQLiteDB(cfg)
	}
}

type sqliteDB struct {
	conn *sql.DB
}

func newSQLiteDB(cfg DatabaseConfig) (*sqliteDB, error) {
	conn, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, err
	}

	db := &sqliteDB{conn: conn}
	if err := db.init(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *sqliteDB) init() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS feeds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_jid TEXT NOT NULL,
			url TEXT NOT NULL,
			title TEXT,
			last_guid TEXT,
			last_poll TIMESTAMP,
			disabled INTEGER DEFAULT 0,
			error_count INTEGER DEFAULT 0,
			UNIQUE(user_jid, url)
		);
		CREATE INDEX IF NOT EXISTS idx_feeds_user ON feeds(user_jid);
	`)
	return err
}

func (db *sqliteDB) AddFeed(userJID, url, title string) (int64, error) {
	result, err := db.conn.Exec(
		"INSERT INTO feeds (user_jid, url, title, last_poll) VALUES (?, ?, ?, ?)",
		userJID, url, title, time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *sqliteDB) RemoveFeed(userJID, url string) (int64, error) {
	result, err := db.conn.Exec(
		"DELETE FROM feeds WHERE user_jid = ? AND url = ?",
		userJID, url,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *sqliteDB) ListFeeds(userJID string) ([]Feed, error) {
	rows, err := db.conn.Query(
		"SELECT id, url, title, last_guid, last_poll, disabled, error_count FROM feeds WHERE user_jid = ?",
		userJID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		var f Feed
		f.UserJID = userJID
		if err := rows.Scan(&f.ID, &f.URL, &f.Title, &f.LastGUID, &f.LastPoll, &f.Disabled, &f.ErrorCount); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, nil
}

func (db *sqliteDB) GetAllFeeds() ([]Feed, error) {
	rows, err := db.conn.Query(
		"SELECT id, user_jid, url, title, last_guid, last_poll, disabled, error_count FROM feeds",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		var f Feed
		if err := rows.Scan(&f.ID, &f.UserJID, &f.URL, &f.Title, &f.LastGUID, &f.LastPoll, &f.Disabled, &f.ErrorCount); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, nil
}

func (db *sqliteDB) UpdateFeedLastGUID(id int64, guid string) error {
	_, err := db.conn.Exec(
		"UPDATE feeds SET last_guid = ?, last_poll = ? WHERE id = ?",
		guid, time.Now(), id,
	)
	return err
}

func (db *sqliteDB) SetFeedDisabled(userJID, url string, disabled bool) (int64, error) {
	disabledInt := 0
	if disabled {
		disabledInt = 1
	}
	result, err := db.conn.Exec(
		"UPDATE feeds SET disabled = ? WHERE user_jid = ? AND url = ?",
		disabledInt, userJID, url,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *sqliteDB) IncrementErrorCount(id int64) (int, error) {
	_, err := db.conn.Exec(
		"UPDATE feeds SET error_count = error_count + 1 WHERE id = ?",
		id,
	)
	if err != nil {
		return 0, err
	}

	var count int
	err = db.conn.QueryRow("SELECT error_count FROM feeds WHERE id = ?", id).Scan(&count)
	return count, err
}

func (db *sqliteDB) ResetErrorCount(id int64) error {
	_, err := db.conn.Exec(
		"UPDATE feeds SET error_count = 0 WHERE id = ?",
		id,
	)
	return err
}

func (db *sqliteDB) Close() error {
	return db.conn.Close()
}

type postgresDB struct {
	conn *sql.DB
}

func newPostgresDB(cfg DatabaseConfig) (*postgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	db := &postgresDB{conn: conn}
	if err := db.init(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *postgresDB) init() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS feeds (
			id SERIAL PRIMARY KEY,
			user_jid TEXT NOT NULL,
			url TEXT NOT NULL,
			title TEXT,
			last_guid TEXT,
			last_poll TIMESTAMP,
			disabled INTEGER DEFAULT 0,
			error_count INTEGER DEFAULT 0,
			UNIQUE(user_jid, url)
		);
		CREATE INDEX IF NOT EXISTS idx_feeds_user ON feeds(user_jid);
	`)
	return err
}

func (db *postgresDB) AddFeed(userJID, url, title string) (int64, error) {
	var id int64
	err := db.conn.QueryRow(
		"INSERT INTO feeds (user_jid, url, title, last_poll) VALUES ($1, $2, $3, $4) RETURNING id",
		userJID, url, title, time.Now(),
	).Scan(&id)
	return id, err
}

func (db *postgresDB) RemoveFeed(userJID, url string) (int64, error) {
	result, err := db.conn.Exec(
		"DELETE FROM feeds WHERE user_jid = $1 AND url = $2",
		userJID, url,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *postgresDB) ListFeeds(userJID string) ([]Feed, error) {
	rows, err := db.conn.Query(
		"SELECT id, url, title, last_guid, last_poll, disabled, error_count FROM feeds WHERE user_jid = $1",
		userJID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		var f Feed
		f.UserJID = userJID
		if err := rows.Scan(&f.ID, &f.URL, &f.Title, &f.LastGUID, &f.LastPoll, &f.Disabled, &f.ErrorCount); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, nil
}

func (db *postgresDB) GetAllFeeds() ([]Feed, error) {
	rows, err := db.conn.Query(
		"SELECT id, user_jid, url, title, last_guid, last_poll, disabled, error_count FROM feeds",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		var f Feed
		if err := rows.Scan(&f.ID, &f.UserJID, &f.URL, &f.Title, &f.LastGUID, &f.LastPoll, &f.Disabled, &f.ErrorCount); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, nil
}

func (db *postgresDB) UpdateFeedLastGUID(id int64, guid string) error {
	_, err := db.conn.Exec(
		"UPDATE feeds SET last_guid = $1, last_poll = $2 WHERE id = $3",
		guid, time.Now(), id,
	)
	return err
}

func (db *postgresDB) SetFeedDisabled(userJID, url string, disabled bool) (int64, error) {
	result, err := db.conn.Exec(
		"UPDATE feeds SET disabled = $1 WHERE user_jid = $2 AND url = $3",
		disabled, userJID, url,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *postgresDB) IncrementErrorCount(id int64) (int, error) {
	_, err := db.conn.Exec(
		"UPDATE feeds SET error_count = error_count + 1 WHERE id = $1",
		id,
	)
	if err != nil {
		return 0, err
	}

	var count int
	err = db.conn.QueryRow("SELECT error_count FROM feeds WHERE id = $1", id).Scan(&count)
	return count, err
}

func (db *postgresDB) ResetErrorCount(id int64) error {
	_, err := db.conn.Exec(
		"UPDATE feeds SET error_count = 0 WHERE id = $1",
		id,
	)
	return err
}

func (db *postgresDB) Close() error {
	return db.conn.Close()
}
