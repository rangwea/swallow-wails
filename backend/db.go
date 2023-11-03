package backend

import (
	"database/sql"
	"path"
	"reflect"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const initSql = `
CREATE TABLE IF NOT EXISTS t_article(
    id INTEGER PRIMARY KEY autoincrement,
    title VARCHAR NOT NULL,
    tags VARCHAR,
    create_time DATETIME,
    update_time DATETIME
);
CREATE INDEX idx_t_article_title ON t_article(title);
CREATE INDEX idx_t_article_tags ON t_article(tags);
CREATE INDEX idx_t_article_create_time ON t_article(create_time);
CREATE INDEX idx_t_article_update_time ON t_article(update_time);
`

type _db struct {
	base *sql.DB
}

var DB = _db{}

func (db *_db) Initialize() {
	dbPath := path.Join(APP_HOME, "db")

	var err error
	databse, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	DB.base = databse
	db.execSql(initSql)
}

func (db *_db) Insert(sqls string, params ...interface{}) (id int64, err error) {
	stmt, err := db.base.Prepare(sqls)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(params...)
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, nil
	}
	return id, nil
}

func (db *_db) Delete(sqls string, params ...interface{}) error {
	return db.execSql(sqls, params...)
}

func (db *_db) Update(sqls string, params ...interface{}) error {
	return db.execSql(sqls, params...)
}

func (db *_db) Query(sqls string, target interface{}, params ...interface{}) (r []interface{}, err error) {
	results := make([]interface{}, 0)

	t := reflect.TypeOf(target)

	rows, err := db.base.Query(sqls, params...)
	if err != nil {
		return nil, err
	}
	cols, _ := rows.Columns()
	values := make([]sql.RawBytes, len(cols))
	scanArgs := make([]interface{}, len(values))
	for i := 0; i < len(values); i++ {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		rows.Scan(scanArgs...)
		obj := reflect.New(t).Elem()
		for i, v := range values {
			fieldName := uderscoreToUpperCamelCase(cols[i])
			field := obj.FieldByName(fieldName)
			if !field.IsValid() {
				continue
			} else {
				switch field.Kind() {
				case reflect.String:
					field.SetString(string(v))
				case reflect.Int64:
					i, _ := strconv.ParseInt(string(v), 10, 64)
					field.SetInt(i)
				}
			}
		}
		results = append(results, obj.Interface())
	}

	return results, nil
}

func (db *_db) execSql(sqls string, params ...interface{}) error {
	stmt, err := db.base.Prepare(sqls)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(params...)
	if err != nil {
		return err
	}
	return nil
}

func uderscoreToUpperCamelCase(s string) string {
	s = strings.Replace(s, "_", " ", -1)
	caser := cases.Title(language.English)
	s = caser.String(s)
	return strings.Replace(s, " ", "", -1)
}
