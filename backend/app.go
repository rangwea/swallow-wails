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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	rt "github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
)

var APP_HOME string

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	// initialize
	initialize()

	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func initialize() {
	// global cons init
	u, err := user.Current()
	if err != nil {
		Logger.Fatal("get user dir fail")
	}
	APP_HOME = path.Join(u.HomeDir, ".swallow")
	if e, _ := PathExists(APP_HOME); !e {
		if err = os.Mkdir(APP_HOME, os.ModePerm); err != nil {
			Logger.Fatal("make app home dir fail", zap.Error(err))
		}
	}

	// component init
	Zap.Initialize()
	DB.Initialize()
	Conf.Initialize()
	Hugos.Initialize()
}

const (
	CODE_SUCCESS = 1
	CODE_ERROR   = 0
)

var (
	ERROR_R = &R{Code: CODE_ERROR}
)

type R struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (a *App) SitePreview(name string) *R {
	Hugos.Preview()
	OpenBrowser("http://localhost:1313/")
	return success(nil)
}

func (a *App) SiteDeploy() *R {
	err := Hugos.Generate()
	if err != nil {
		Logger.Error("hugo generate error", zap.Error(err))
		return ERROR_R
	}

	g, err := Conf.Read(GITHUB)
	if err != nil {
		Logger.Error("get config fail", zap.Error(err))
		return ERROR_R
	}
	github := g.(Github)

	git.PlainInit(Hugos.PublicDir, false)

	r, err := git.PlainOpen(Hugos.PublicDir)
	if err != nil {
		Logger.Error("open git repository error", zap.Error(err))
		return ERROR_R
	}
	w, err := r.Worktree()
	if err != nil {
		Logger.Error("open git worktree error", zap.Error(err))
		return ERROR_R
	}
	_, err = w.Add(".")
	if err != nil {
		Logger.Error("git add error", zap.Error(err))
		return ERROR_R
	}
	_, err = w.Commit("deploy", &git.CommitOptions{
		Author: &object.Signature{
			Email: github.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		Logger.Error("git commit error", zap.Error(err))
		return ERROR_R
	}

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{github.Repository},
	})
	if err != nil {
		Logger.Error("git remote error", zap.Error(err))
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
		Logger.Error("git push error", zap.Error(err))
		return ERROR_R
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
	r, err := DB.Query(sql, Article{}, search, search, search)
	if err != nil {
		Logger.Error("query article fail", zap.String("sql", sql), zap.Error(err))
		return ERROR_R
	}
	Logger.Info("article list", zap.String("sql", sql), zap.Any("result", r))
	return success(r)
}

func (a *App) ArticleSave(aid string, meta Meta, content string) *R {
	Logger.Info("article save", zap.Any("meta", meta))

	meta.Lastmod = time.Now().Format("2006-01-02 15:04:05")

	if aid != ABOUT_AID {
		// common article need save db
		err := savedb(&aid, meta)
		if err != nil {
			Logger.Error("save article into db fail", zap.Error(err))
			return ERROR_R
		}
	}

	err := Hugos.WriteArticle(aid, meta, content)
	if err != nil {
		Logger.Error("article write fail", zap.Error(err))
		return ERROR_R
	}

	return success(aid)
}

func (a *App) ArticleGet(aid string) *R {
	meta, content, err := Hugos.ReadArticle(aid)
	if err != nil {
		Logger.Error("get article fail", zap.String("aid", aid), zap.Error(err))
		return ERROR_R
	}
	data := map[string]interface{}{
		"meta":    meta,
		"content": content,
	}
	return success(data)
}

func (a *App) ArticleRemove(aids []string) *R {
	err := DB.Delete(fmt.Sprintf("delete from t_article where id in(%s)", strings.Join(aids[:], ",")))
	if err != nil {
		Logger.Error("delete article fail", zap.Any("aids", aids), zap.Error(err))
		return ERROR_R
	}
	for _, aid := range aids {
		Hugos.DeleteArticle(aid)
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
		Logger.Error("select image fail", zap.Error(err))
		return ERROR_R
	}

	Logger.Info("", zap.Any("select", selection))

	imageDir := Hugos.getArticleImageDir(aid)
	os.Mkdir(imageDir, os.ModePerm)

	locaPath, sitePath := Hugos.genArticleImagePath(aid)
	err = CopyFile(selection, locaPath)
	if err != nil {
		Logger.Error("copy image fail", zap.String("selection", selection), zap.Error(err))
		return ERROR_R
	}

	return success(sitePath)
}

func (a *App) ArticleInsertImageBlob(aid int, blob string) *R {
	file := []byte{}
	if err := json.Unmarshal([]byte(blob), &file); err != nil {
		Logger.Error("parse file", zap.Error(err))
		return ERROR_R
	}

	aida := strconv.Itoa(aid)

	imageDir := Hugos.getArticleImageDir(aida)
	os.Mkdir(imageDir, os.ModePerm)

	locaPath, sitePath := Hugos.genArticleImagePath(aida)
	err := os.WriteFile(locaPath, file, os.ModePerm)
	if err != nil {
		Logger.Error("write image fail", zap.Error(err))
		return ERROR_R
	}

	return success(sitePath)
}

func (a *App) SiteConfigGet() *R {
	c, err := Hugos.ReadConfig()
	if err != nil {
		Logger.Error("get site confi fail", zap.Error(err))
		return ERROR_R
	}
	return success(c)
}

func (a *App) SiteConfigSave(c Config) *R {
	err := Hugos.WriteConfig(c)
	if err != nil {
		Logger.Error("save site confi fail", zap.Error(err))
		return ERROR_R
	}
	return success(nil)
}

func (a *App) ConfGet(t ConfType) *R {
	v, err := Conf.Read(t)
	if err != nil {
		Logger.Error("read conf fail", zap.Any("type", t), zap.Error(err))
		return ERROR_R
	}
	return success(v)
}

func (a *App) ConfSave(t ConfType, v interface{}) *R {
	err := Conf.Write(t, v)
	if err != nil {
		Logger.Error("save conf fail", zap.Any("type", t), zap.Any("data", v), zap.Error(err))
		return ERROR_R
	}
	return success(nil)
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
		Logger.Error("select image fail", zap.Error(err))
		return ERROR_R
	}

	p := path.Join(Hugos.SitePath, imgPath)
	// remove old
	os.Remove(p)

	// copy conf image
	err = CopyFile(selection, p)
	if err != nil {
		Logger.Error("copy image fail", zap.String("selection", selection), zap.Error(err))
		return ERROR_R
	}

	return success(nil)
}

func savedb(aidpr *string, meta Meta) error {
	title := meta.Title
	createTime := meta.Date
	tags := strings.Join(meta.Tags, ",")
	updateTime := meta.Lastmod

	aid := *aidpr

	if aid == "" {
		id, err := DB.Insert("insert into t_article(title, tags, create_time, update_time) values(?,?,?,?)",
			title, tags, createTime, updateTime)
		if err != nil {
			Logger.Error("article save fail", zap.Error(err))
			return err
		}
		*aidpr = strconv.FormatInt(id, 10)
	} else {
		id, err := strconv.Atoi(aid)
		if err != nil {
			Logger.Error("article save fail, id invalid", zap.Any("aid", aid), zap.Error(err))
			return err
		}
		err = DB.Update("update t_article set title=?, tags=?, update_time=? where id=?",
			title, tags, updateTime, id)
		if err != nil {
			Logger.Error("article save fail", zap.Error(err))
			return err
		}
	}
	return nil
}

func success(data interface{}) *R {
	return &R{Code: CODE_SUCCESS, Msg: "成功", Data: data}
}
