package backend

type Article struct {
	Id          int64  `json:"id"`
	Title       string `json:"title"`
	Tags        string `json:"tags"`
	Description string `json:"description"`
	CreateTime  string `json:"createTime"`
	UpdateTime  string `json:"updateTime"`
}
