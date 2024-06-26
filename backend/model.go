package backend

import (
	"strings"
	"time"
)

type LocalTime time.Time

func (c *LocalTime) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`) //get rid of "
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse("2006-01-02 15:04:05", value) //parse time
	if err != nil {
		return err
	}

	*c = LocalTime(t) //set result using the pointer
	return nil
}

func (c LocalTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(c).Format("2006-01-02") + `"`), nil
}

type Article struct {
	Id          int64     `json:"id"`
	Title       string    `json:"title"`
	Tags        string    `json:"tags"`
	Description string    `json:"description"`
	CreateTime  LocalTime `json:"createTime" db:"create_time"`
	UpdateTime  LocalTime `json:"updateTime" db:"update_time"`
}
