package datebase

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type Client struct {
	DB *sql.DB
}

func NewClient() Client {
	db, err := sql.Open("sqlite3", "/usr/local/goseeder.db")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS Torrent (torrent_hash CHAR,title CHAR,torrent_announce CHAR,create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,PRIMARY KEY('torrent_hash'))")
	if err != nil {
		panic(err)
	}

	return Client{
		DB: db,
	}
}

func (c *Client) Get(hashId string) bool {
	var torrent_hash string
	var title string
	var torrent_announce string
	var create_time time.Time

	// 查询数据
	rows, err := c.DB.Query("SELECT * FROM Torrent WHERE torrent_hash == '" + hashId + "'")
	if err != nil {
		panic(err)
	}

	defer rows.Close()
	rows.Next()

	err = rows.Scan(&torrent_hash, &title, &torrent_announce, &create_time)
	if err == nil {
		return true
	}

	return false
}

func (c *Client) Insert(Title string, TorrentHash string, TorrentAnnounce string) bool {
	// 插入数据
	stmt, err := c.DB.Prepare("INSERT INTO Torrent ('torrent_hash', 'title', 'torrent_announce') VALUES (?,?,?)")
	if err != nil {
		panic(err)
	}

	_, err = stmt.Exec(TorrentHash, Title, TorrentAnnounce)
	if err != nil {
		panic(err)
	}

	return true
}
