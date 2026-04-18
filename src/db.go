package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

type Feed struct {
	ID         int64
	UserJID    string
	URL        string
	Title      string
	LastGUID   string
	LastPoll   time.Time
	Disabled   bool
	ErrorCount int
}

func NewDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.init(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) init() error {
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

func (db *DB) AddFeed(userJID, url, title string) (int64, error) {
	result, err := db.conn.Exec(
		"INSERT INTO feeds (user_jid, url, title, last_poll) VALUES (?, ?, ?, ?)",
		userJID, url, title, time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) RemoveFeed(userJID, url string) (int64, error) {
	result, err := db.conn.Exec(
		"DELETE FROM feeds WHERE user_jid = ? AND url = ?",
		userJID, url,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *DB) ListFeeds(userJID string) ([]Feed, error) {
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

func (db *DB) GetAllFeeds() ([]Feed, error) {
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

func (db *DB) UpdateFeedLastGUID(id int64, guid string) error {
	_, err := db.conn.Exec(
		"UPDATE feeds SET last_guid = ?, last_poll = ? WHERE id = ?",
		guid, time.Now(), id,
	)
	return err
}

func (db *DB) SetFeedDisabled(userJID, url string, disabled bool) (int64, error) {
	result, err := db.conn.Exec(
		"UPDATE feeds SET disabled = ? WHERE user_jid = ? AND url = ?",
		disabled, userJID, url,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *DB) IncrementErrorCount(id int64) (int, error) {
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

func (db *DB) ResetErrorCount(id int64) error {
	_, err := db.conn.Exec(
		"UPDATE feeds SET error_count = 0 WHERE id = ?",
		id,
	)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}
