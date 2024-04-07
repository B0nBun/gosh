package dbservice

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
	"strings"
)

const alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randAlphanumString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	for i := 0; i < n; i++ {
		idx := rand.Int63() % int64(len(alphanumeric))
		sb.WriteByte(alphanumeric[idx])
	}

	return sb.String()
}

const SlugLen = 5

type DBService struct {
	db *sql.DB
}

func MakeDBService(dsName string) (s DBService, err error) {
	s.db, err = sql.Open("sqlite3", dsName)
	if err != nil {
		return
	}
	s.db.SetConnMaxLifetime(0)
	s.db.SetMaxIdleConns(50)
	s.db.SetMaxOpenConns(50)

	_, err = s.db.Exec(fmt.Sprintf(`
	create table if not exists shorturls (
		slug varchar(%d) primary key,
		url text not null,
		visits unsigned big int default 0
	);`, SlugLen))

	return
}

func (s *DBService) CreateShortenedUrl(url string) (string, error) {
	for {
		slug := randAlphanumString(SlugLen)
		res, err := s.db.Exec("insert into shorturls (slug, url) values (?, ?)", slug, url)
		if err != nil {
			return "", err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return "", err
		}
		if affected != 0 {
			return slug, nil
		}
	}
}

func (s *DBService) GetUrl(slug string) (url string, err error, exists bool) {
	// TODO: Increment visits
	err = s.db.QueryRow("select url from shorturls where slug = $1", slug).Scan(&url)
	exists = err != sql.ErrNoRows
	return
}

func (s *DBService) Close() error {
	return s.db.Close()
}
