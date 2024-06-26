package backend

import (
	"bufio"
	"bytes"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gohugoio/hugo/config"
	"github.com/gohugoio/hugo/config/allconfig"
	"github.com/gohugoio/hugo/create/skeletons"
	"github.com/gohugoio/hugo/deps"
	"github.com/gohugoio/hugo/hugofs"
	"github.com/gohugoio/hugo/hugolib"
	"github.com/gohugoio/hugo/livereload"
	"github.com/pkg/errors"
	"golang.org/x/exp/slog"

	"github.com/BurntSushi/toml"
	cp "github.com/otiai10/copy"
)

const AboutAid string = "about"

type _hugo struct {
	PublicDir     string
	SitePath      string
	ImageDir      string
	hugo          string
	articleDir    string
	articleImgDir string
	themeDir      string
	aboutDir      string
	aboutFile     string
	cnameFile     string
	configFile    string
	server        *http.Server
	serverRunning atomic.Bool
}

var Hugo = _hugo{}

type Meta struct {
	Title       string   `json:"title"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Date        string   `json:"date"`
	Lastmod     string   `json:"lastmod"`
}

type Config struct {
	Title                  string        `json:"title"`
	Description            string        `json:"description"`
	DefaultContentLanguage string        `json:"defaultContentLanguage"`
	Theme                  string        `json:"theme"`
	Copyright              string        `json:"copyright"`
	Params                 *ConfigParams `json:"params"`
}

type ConfigParams struct {
	Author *ConfigAuthor `json:"author"`
}

type ConfigAuthor struct {
	Name string `json:"name"`
}

func (h *_hugo) Initialize() {
	slog.Info("init hugo start")
	h.hugo = path.Join(AppHome, "hugo")
	h.SitePath = path.Join(AppHome, "site")
	h.articleDir = path.Join(h.SitePath, "content", "post")
	h.articleImgDir = path.Join(h.articleDir, "images")
	h.ImageDir = path.Join(h.SitePath, "static", "images")
	h.cnameFile = path.Join(h.SitePath, "static", "CNAME")
	h.themeDir = path.Join(h.SitePath, "themes")
	h.aboutDir = path.Join(h.SitePath, "content", AboutAid)
	h.aboutFile = path.Join(h.SitePath, "content", AboutAid, "index.md")
	h.configFile = path.Join(h.SitePath, "hugo.toml")
	h.PublicDir = path.Join(h.SitePath, "public")

	h.NewSite()

	h.server = &http.Server{Addr: ":1313"}
	http.Handle("/", http.FileServer(http.Dir(path.Join(h.SitePath, "public"))))
	livereload.Initialize()
	http.HandleFunc("/livereload.js", livereload.ServeJS)
	http.HandleFunc("/livereload", livereload.Handler)

	slog.Info("init hugo done")
}

func (h *_hugo) NewSite() {
	slog.Info("start new site")
	if existed, _ := PathExists(h.SitePath); existed {
		slog.Info("site existed")
		return
	}

	// call hugo to create site
	err := skeletons.CreateSite(h.SitePath, hugofs.Os, false, "toml")
	if err != nil {
		slog.Error("copy config fail", err)
		return
	}

	// copy config file
	os.Remove(h.configFile)
	err = CopyAsset("hugo.toml", h.configFile)
	if err != nil {
		slog.Error("copy config fail", err)
		return
	}
	slog.Info("copy config file")

	err = h.setWorkingDirConfig()
	if err != nil {
		slog.Error("set workingDir config fail", err)
		return
	}

	// copy theme zip file
	themeCopyDstPath := path.Join(AppHome, "themes.zip")
	err = CopyAsset("themes.zip", themeCopyDstPath)
	if err != nil {
		slog.Error("copy themes.zip fail", err)
	}
	// unzip
	err = UnZip(themeCopyDstPath, h.SitePath)
	if err != nil {
		slog.Error("unzip themes file fail", err)
		return
	}
	// remove zip
	os.Remove(themeCopyDstPath)
	slog.Info("unzip theme file")

	os.Mkdir(h.articleDir, os.ModePerm)
	os.Mkdir(h.articleImgDir, os.ModePerm)
	os.Mkdir(h.ImageDir, os.ModePerm)
	os.Create(h.cnameFile)
	// create about post
	os.Mkdir(h.aboutDir, os.ModePerm)
	os.Create(h.aboutFile)

	slog.Info("new site success")
}

func (h *_hugo) Build() (err error) {
	configs, err := allconfig.LoadConfig(allconfig.ConfigSourceDescriptor{
		Fs: hugofs.Os, Filename: path.Join(h.SitePath, "hugo.toml"),
	})
	if err != nil {
		slog.Error("load hugo build config fail", err)
		return
	}

	fs := hugofs.NewFrom(hugofs.Os, config.BaseConfig{WorkingDir: h.SitePath, PublishDir: "public"})
	s, err := hugolib.NewHugoSites(deps.DepsCfg{Configs: configs, Fs: fs})
	if err != nil {
		slog.Error("new hugo sites fail", err)
		return
	}

	// copy static
	t, err := h.getCurrentTheme()
	if err != nil {
		return errors.Wrap(err, "get current theme fail")
	}
	err = cp.Copy(path.Join(h.themeDir, t, "static"), h.PublicDir)
	if err != nil {
		return errors.Wrap(err, "copy theme static file fail")
	}

	// build
	err = s.Build(hugolib.BuildCfg{ErrRecovery: true})
	if err != nil {
		slog.Error("hugo site build fail", err)
		return
	}
	return
}

func (h *_hugo) Preview() error {
	err := h.Build()
	if err != nil {
		return err
	}
	if h.serverRunning.Swap(true) {
		return nil
	}
	go func() {
		err := h.server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("preview server fail", err)
		}
		slog.Info("preview running")
	}()
	return nil
}

func (h *_hugo) ClosePreview() error {
	if h.serverRunning.Swap(false) {
		slog.Info("close preview server")
		return h.server.Close()
	}
	return nil
}

func (h *_hugo) WriteArticle(aid string, meta Meta, content string) error {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(meta)
	if err != nil {
		slog.Error("encode meta fail", err)
		return err
	}
	metaString := "+++\n" + buf.String() + "+++\n"
	content = metaString + content

	var articleF string
	if aid == AboutAid {
		// about file
		articleF = h.aboutFile
	} else {
		// common article file
		adir := path.Join(h.articleDir, aid)
		if e, _ := PathExists(adir); !e {
			err = os.MkdirAll(adir, os.ModePerm)
			if err != nil {
				return err
			}
		}
		articleF = path.Join(adir, "index.md")
	}

	err = os.WriteFile(articleF, []byte(content), os.ModePerm)
	if err != nil {
		slog.Error("write article fail", err)
		return err
	}
	return nil
}

func (h *_hugo) ReadArticle(aid string) (meta Meta, content string, err error) {
	var p string
	if aid == AboutAid {
		p = h.aboutFile // about file
	} else {
		p = path.Join(h.articleDir, aid, "index.md") // common article file
	}
	a, err := os.ReadFile(p)
	if err != nil {
		slog.Error("read article fail", err)
		return Meta{}, "", err
	}
	m, c := h.SplitMetaAndContent(string(a))
	meta = Meta{}
	_, err = toml.Decode(m, &meta)
	if err != nil {
		slog.Error("decode meta fail when reading article", err)
		return Meta{}, "", err
	}
	return meta, c, nil
}

func (h *_hugo) DeleteArticle(aid string) error {
	p := path.Join(h.articleDir, aid)
	err := os.RemoveAll(p)
	if err != nil {
		slog.Error("remove article file fail", err)
		return err
	}
	return nil
}

func (h *_hugo) getArticleImageDir(aid string) string {
	return path.Join(h.SitePath, "/static/images", aid)
}

func (h *_hugo) genArticleImagePath(aid string) (localPath string, sitePath string) {
	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ".png"
	sitePath = path.Join("/static/images", aid, filename)
	localPath = path.Join(h.SitePath, sitePath)
	return localPath, sitePath
}

func (h *_hugo) ReadConfig() (c Config, err error) {
	b, err := os.ReadFile(h.configFile)
	if err != nil {
		slog.Error("read config fail", err)
		return Config{}, err
	}
	r := Config{}
	_, err = toml.Decode(string(b), &r)
	if err != nil {
		slog.Error("decode config fail", err)
		return Config{}, err
	}
	return r, nil
}

func (h *_hugo) WriteConfig(c Config) error {
	b, err := os.ReadFile(h.configFile)
	if err != nil {
		return err
	}
	old := make(map[string]interface{})
	_, err = toml.Decode(string(b), &old)
	if err != nil {
		slog.Error("decode config fail", err)
		return err
	}

	old["title"] = c.Title
	old["description"] = c.Description
	old["defaultContentLanguage"] = c.DefaultContentLanguage
	old["theme"] = c.Theme
	old["copyright"] = c.Copyright
	oldParams := old["params"].(map[string]interface{})
	oldAuthor := oldParams["author"].(map[string]interface{})
	oldAuthor["name"] = c.Params.Author.Name
	oldParams["author"] = oldAuthor
	old["params"] = oldParams

	buf := new(bytes.Buffer)
	err = toml.NewEncoder(buf).Encode(old)
	if err != nil {
		slog.Error("encode config fail", err)
		return err
	}

	err = os.WriteFile(h.configFile, buf.Bytes(), os.ModePerm)
	if err != nil {
		slog.Error("write config fail", err)
		return err
	}

	return nil
}

func (h *_hugo) SplitMetaAndContent(article string) (meta string, content string) {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(article))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if len(lines) == 0 {
		return "", ""
	}

	inMeta := false
	// the index of second +++
	var secondCodeMark int
	for i, line := range lines {
		if !inMeta && strings.HasPrefix(line, "+++") {
			inMeta = true
			continue
		}
		if inMeta && strings.HasPrefix(line, "+++") {
			secondCodeMark = i
			continue
		}
	}

	meta = strings.Join(lines[1:secondCodeMark], "\n")
	content = strings.Join(lines[secondCodeMark+1:], "\n")

	return meta, content
}

func (h *_hugo) GetThemes() (themes []string, err error) {
	es, err := os.ReadDir(h.themeDir)
	if err != nil {
		return
	}
	for _, e := range es {
		themes = append(themes, e.Name())
	}
	return
}

func (h *_hugo) setWorkingDirConfig() error {
	b, err := os.ReadFile(h.configFile)
	if err != nil {
		slog.Error("read config fail", err)
		return err
	}
	c := make(map[string]interface{})
	_, err = toml.Decode(string(b), &c)
	if err != nil {
		slog.Error("decode config fail", err)
		return err
	}

	c["workingDir"] = h.SitePath

	buf := new(bytes.Buffer)
	err = toml.NewEncoder(buf).Encode(c)
	if err != nil {
		slog.Error("encode config fail", err)
		return err
	}
	err = os.WriteFile(h.configFile, buf.Bytes(), os.ModePerm)
	if err != nil {
		slog.Error("write config fail", err)
		return err
	}
	return nil
}

func (h *_hugo) getCurrentTheme() (string, error) {
	c, err := h.ReadConfig()
	if err != nil {
		return "", err
	}
	return c.Theme, nil
}
