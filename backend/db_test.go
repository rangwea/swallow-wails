package backend

import (
	"fmt"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	DB.Initialize()
	_, err := DB.Insert("insert into t_article(title,tags,create_time,update_time) values(?,?,?,?,?)", "标题1", "tag1,tag1", "2022-03-11 00:01:02", time.Now())
	if err != nil {
		fmt.Println(err)
	}
}

func TestQuery(t *testing.T) {
	DB.Initialize()
	d, err := DB.Query("select * from t_article", Article{})
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range d {
		fmt.Println(v)
	}
}
