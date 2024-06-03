package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/slog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	_ "github.com/mattn/go-sqlite3"
	rt "github.com/wailsapp/wails/v2/pkg/runtime"
)

var AppHome string
var DB *sqlx.DB

const InitSql = `CREATE TABLE IF NOT EXISTS t_article(
    id INTEGER PRIMARY KEY autoincrement,
    title VARCHAR NOT NULL,
    tags VARCHAR,
    create_time DATETIME,
    update_time DATETIME
);
CREATE INDEX IF NOT EXISTS idx_t_article_title ON t_article(title);
CREATE INDEX IF NOT EXISTS idx_t_article_tags ON t_article(tags);
CREATE INDEX IF NOT EXISTS idx_t_article_create_time ON t_article(create_time);
CREATE INDEX IF NOT EXISTS idx_t_article_update_time ON t_article(update_time);`

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
func (a *App) Startup(ctx context.Context) {
	// initialize
	initialize()
	a.ctx = ctx
}

func initialize() {
	// global cons init
	initAppHome()

	// db init
	initDB()

	// component init
	Conf.Initialize()
	Hugo.Initialize()
}

func initAppHome() {
	u, err := user.Current()
	if err != nil {
		slog.Error("get user dir fail", err)
	}
	AppHome = path.Join(u.HomeDir, ".swallow")
	if e, _ := PathExists(AppHome); !e {
		if err = os.Mkdir(AppHome, os.ModePerm); err != nil {
			slog.Error("make app home dir fail", err)
		}
	}
}

func initDB() {
	DB = sqlx.MustOpen("sqlite3", path.Join(AppHome, "db"))
	DB.MustExec(InitSql)
}

const (
	CodeSuccess = 1
	CodeError   = 0
)

type R struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (a *App) SitePreview() *R {
	err := Hugo.Preview()
	if err != nil {
		slog.Error("preview fail", err)
		return failM(err.Error())
	}
	err = OpenBrowser("http://localhost:1313/")
	if err != nil {
		slog.Error("preview fail", err)
		return failM(err.Error())
	}
	return success(nil)
}

func (a *App) SiteDeploy() *R {
	err := Hugo.Build()
	if err != nil {
		slog.Error("hugo generate error", err)
		return failM(err.Error())
	}

	g, err := Conf.Read(GITHUB)
	if err != nil {
		slog.Error("get config fail", err)
		return failM(err.Error())
	}
	if g == nil {
		return failM("please set github config")
	}
	github := g.(Github)

	_, err = git.PlainInit(Hugo.PublicDir, false)
	if err != nil {
		slog.Error("git init fail", err)
		return failM(err.Error())
	}

	r, err := git.PlainOpen(Hugo.PublicDir)
	if err != nil {
		slog.Error("open git repository error", err)
		return failM(err.Error())
	}
	w, err := r.Worktree()
	if err != nil {
		slog.Error("open git worktree error", err)
		return failM(err.Error())
	}
	_, err = w.Add(".")
	if err != nil {
		slog.Error("git add error", err)
		return failM(err.Error())
	}
	_, err = w.Commit("deploy", &git.CommitOptions{
		Author: &object.Signature{
			Email: github.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		slog.Error("git commit error", err)
		return failM(err.Error())
	}

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{github.Repository},
	})
	if err != nil {
		slog.Error("git remote error", err)
	}

	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Force:      true,
		Auth: &http.BasicAuth{
			Username: github.Username,
			Password: github.Token,
		},
	})
	if err != nil {
		slog.Error("git push error", err)
		return failM(err.Error())
	}

	return success(nil)
}

func (a *App) ArticleList(search string) *R {
	sql := "select * from t_article"
	if search != "" {
		sql += " where title like ? or tags like ?"
	}
	sql += " order by update_time desc"
	search = "%" + search + "%"
	r := []Article{}
	err := DB.Select(&r, sql, search, search, search)
	if err != nil {
		slog.Error("query article fail", err)
		return failM(err.Error())
	}
	slog.Debug("article list", "sql", sql, "result", r)
	return success(r)
}

func (a *App) ArticleSave(aid string, meta Meta, content string) *R {
	slog.Debug("article save", meta)

	n := time.Now().Format("2006-01-02 15:04:05")
	meta.Lastmod = n
	if meta.Date == "" {
		meta.Date = n
	}

	if aid != AboutAid {
		// common article need save db
		err := saveArticleToDB(&aid, meta)
		if err != nil {
			slog.Error("save article into db fail", err)
			return failM(err.Error())
		}
	}

	err := Hugo.WriteArticle(aid, meta, content)
	if err != nil {
		slog.Error("article write fail", err)
		return failM(err.Error())
	}

	return success(aid)
}

func (a *App) ArticleGet(aid string) *R {
	meta, content, err := Hugo.ReadArticle(aid)
	if err != nil {
		slog.Error("get article fail", err)
		return failM(err.Error())
	}
	data := map[string]interface{}{
		"meta":    meta,
		"content": content,
	}
	return success(data)
}

func (a *App) ArticleRemove(aids []string) *R {
	_, err := DB.Exec(fmt.Sprintf("delete from t_article where id in(%s)", strings.Join(aids[:], ",")))
	if err != nil {
		slog.Error("delete article fail", err)
		return failM(err.Error())
	}
	for _, aid := range aids {
		Hugo.DeleteArticle(aid)
	}
	return success(nil)
}

func (a *App) ArticleInsertImage(aid string) *R {
	selection, err := rt.OpenFileDialog(a.ctx, rt.OpenDialogOptions{
		Title: "Select Image",
		Filters: []rt.FileFilter{
			{
				DisplayName: "Images (*.png;*.jpg;*.gif;*.jpeg)",
				Pattern:     "*.png;*.jpg;*.gif;*.jpeg",
			},
		},
	})
	if err != nil {
		slog.Error("select image fail", err)
		return failM(err.Error())
	}

	imageDir := Hugo.getArticleImageDir(aid)
	os.Mkdir(imageDir, os.ModePerm)

	localPath, sitePath := Hugo.genArticleImagePath(aid)
	err = CopyFile(selection, localPath)
	if err != nil {
		slog.Error("copy image fail", err)
		return failM(err.Error())
	}

	return success(sitePath)
}

func (a *App) ArticleInsertImageBlob(aid int, blob string) *R {
	var file []byte
	if err := json.Unmarshal([]byte(blob), &file); err != nil {
		slog.Error("parse file", err)
		return failM(err.Error())
	}

	aida := strconv.Itoa(aid)

	imageDir := Hugo.getArticleImageDir(aida)
	os.Mkdir(imageDir, os.ModePerm)

	localPath, sitePath := Hugo.genArticleImagePath(aida)
	err := os.WriteFile(localPath, file, os.ModePerm)
	if err != nil {
		slog.Error("write image fail", err)
		return failM(err.Error())
	}

	return success(sitePath)
}

func (a *App) SiteConfigGet() *R {
	c, err := Hugo.ReadConfig()
	if err != nil {
		slog.Error("get site config fail", err)
		return failM(err.Error())
	}
	return success(c)
}

func (a *App) SiteConfigSave(c Config) *R {
	err := Hugo.WriteConfig(c)
	if err != nil {
		slog.Error("save site config fail", err)
		return failM(err.Error())
	}
	return success(nil)
}

func (a *App) ConfGet(t ConfType) *R {
	v, err := Conf.Read(t)
	if err != nil {
		slog.Error("read conf fail", err)
		return failM(err.Error())
	}
	return success(v)
}

func (a *App) ConfSave(t ConfType, v interface{}) *R {
	err := Conf.Write(t, v)
	if err != nil {
		slog.Error("save conf fail", err)
		return failM(err.Error())
	}
	return success(nil)
}

func (a *App) ConfGetThemes() *R {
	ts, err := Hugo.GetThemes()
	if err != nil {
		return failM(err.Error())
	}
	return success(ts)
}

func (a *App) SelectConfImage(imgPath string) *R {
	selection, err := rt.OpenFileDialog(a.ctx, rt.OpenDialogOptions{
		Title: "Select Image",
		Filters: []rt.FileFilter{
			{
				DisplayName: "Images (*.png;*.jpg;*.gif;*.jpeg;*.ico)",
				Pattern:     "*.png;*.jpg;*.gif;*.jpeg;*.ico",
			},
		},
	})
	if err != nil {
		slog.Error("select image fail", err)
		return failM(err.Error())
	}

	p := path.Join(Hugo.SitePath, imgPath)
	// remove old
	os.Remove(p)

	// copy conf image
	err = CopyAsset(selection, p)
	if err != nil {
		slog.Error("copy image fail", err)
		return failM(err.Error())
	}

	return success(nil)
}

func saveArticleToDB(aidpr *string, meta Meta) error {
	title := meta.Title
	createTime := meta.Date
	tags := strings.Join(meta.Tags, ",")
	updateTime := meta.Lastmod

	aid := *aidpr

	if aid == "" {
		r, err := DB.Exec("insert into t_article(title, tags, create_time, update_time) values(?,?,?,?)",
			title, tags, createTime, updateTime)
		if err != nil {
			slog.Error("article save fail", err)
			return err
		}
		nid, err := r.LastInsertId()
		if err != nil {
			slog.Error("article save fail", err)
			return err
		}
		*aidpr = strconv.FormatInt(nid, 10)
	} else {
		id, err := strconv.Atoi(aid)
		if err != nil {
			slog.Error("article save fail, id invalid", err)
			return err
		}
		_, err = DB.Exec("update t_article set title=?, tags=?, update_time=? where id=?",
			title, tags, updateTime, id)
		if err != nil {
			slog.Error("article save fail", err)
			return err
		}
	}
	return nil
}

func success(data interface{}) *R {
	return &R{Code: CodeSuccess, Msg: "success", Data: data}
}

func fail() *R {
	return &R{Code: CodeError, Msg: "error"}
}

func failM(msg string) *R {
	return &R{Code: CodeError, Msg: msg}
}
