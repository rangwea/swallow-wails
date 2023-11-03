package backend

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
)

const ABOUT_AID string = "about"

type _hugos struct {
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
}

var Hugos = _hugos{}

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
	Author                 *ConfigAuthor `json:"author"`
}

type ConfigAuthor struct {
	Name string `json:"name"`
}

func (h *_hugos) Initialize() {
	Logger.Info("init hugos start")
	h.hugo = path.Join(APP_HOME, "hugo")
	h.SitePath = path.Join(APP_HOME, "site")
	h.articleDir = path.Join(h.SitePath, "content", "post")
	h.articleImgDir = path.Join(h.articleDir, "images")
	h.ImageDir = path.Join(h.SitePath, "static", "images")
	h.cnameFile = path.Join(h.SitePath, "static", "CNAME")
	h.themeDir = path.Join(h.SitePath, "themes")
	h.aboutDir = path.Join(h.SitePath, "content", ABOUT_AID)
	h.aboutFile = path.Join(h.SitePath, "content", ABOUT_AID, "index.md")
	h.configFile = path.Join(h.SitePath, "hugo.toml")
	h.PublicDir = path.Join(h.SitePath, "public")

	err := CopyFile("assets/hugo", h.hugo, 0755)
	if err != nil {
		Logger.Error("copy config fail", zap.Error(err))
		return
	}

	h.NewSite()

	Logger.Info("init hugos done")
}

func (h *_hugos) NewSite() {
	Logger.Info("start new site")
	if existed, _ := PathExists(h.SitePath); existed {
		Logger.Info("site existed")
		return
	}

	// hugo cmd: create a site
	h.execHugoCmd(APP_HOME, "new", "site", h.SitePath)
	os.Mkdir(h.articleDir, os.ModePerm)
	os.Mkdir(h.articleImgDir, os.ModePerm)
	os.Mkdir(h.ImageDir, os.ModePerm)
	os.Create(h.cnameFile)
	// create about post
	os.Mkdir(h.aboutDir, os.ModePerm)
	os.Create(h.aboutFile)

	os.Remove(h.configFile)

	// copy config file
	err := CopyFile("assets/hugo.toml", h.configFile)
	if err != nil {
		Logger.Error("copy config fail", zap.Error(err))
		return
	}
	Logger.Info("copy config file")
	// copy theme zip file
	err = UnZip("assets/themes.zip", h.themeDir)
	if err != nil {
		Logger.Error("unzip theme file fail", zap.Error(err))
		return
	}
	Logger.Info("unzip theme file")

	Logger.Info("new site success")
}

func (h *_hugos) Preview() error {
	err := KillProcessByName("hugo")
	if err != nil {
		return err
	}
	err = h.execHugoCmdStart(h.SitePath, "server")
	if err != nil {
		return err
	}
	return nil
}

func (h *_hugos) ClosePreview() error {
	err := KillProcessByName("hugo")
	if err != nil {
		return err
	}
	return nil
}

func (h *_hugos) Generate() (err error) {
	err = h.execHugoCmd(h.SitePath)
	if err != nil {
		return nil
	}
	Logger.Info("hugo generate")
	return err
}

func (h *_hugos) WriteArticle(aid string, meta Meta, content string) error {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(meta)
	if err != nil {
		return err
	}
	metaString := "+++\n" + buf.String() + "+++\n"
	content = metaString + content

	var afile string
	if aid == ABOUT_AID {
		// about file
		afile = h.aboutFile
	} else {
		// common article file
		adir := path.Join(h.articleDir, aid)
		if e, _ := PathExists(adir); !e {
			os.Mkdir(adir, os.ModePerm)
		}
		afile = path.Join(adir, "index.md")
	}

	err = os.WriteFile(afile, []byte(content), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (h *_hugos) ReadArticle(aid string) (meta Meta, content string, err error) {
	var p string
	if aid == ABOUT_AID {
		p = h.aboutFile // about file
	} else {
		p = path.Join(h.articleDir, aid, "index.md") // common article file
	}
	a, err := os.ReadFile(p)
	if err != nil {
		return Meta{}, "", err
	}
	m, c := h.SplitMetaAndContent(string(a))
	meta = Meta{}
	toml.Decode(m, &meta)
	return meta, c, nil
}

func (h *_hugos) DeleteArticle(aid string) error {
	p := path.Join(h.articleDir, aid)
	os.RemoveAll(p)
	return nil
}

func (h *_hugos) getArticleImageDir(aid string) string {
	return path.Join(h.SitePath, "/static/images", aid)
}

func (h *_hugos) genArticleImagePath(aid string) (localPath string, sitePath string) {
	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + ".png"
	sitePath = path.Join("/static/images", aid, filename)
	localPath = path.Join(h.SitePath, sitePath)
	return localPath, sitePath
}

func (h *_hugos) ReadConfig() (c Config, err error) {
	b, err := os.ReadFile(h.configFile)
	if err != nil {
		return Config{}, err
	}
	r := Config{}
	toml.Decode(string(b), &r)
	return r, nil
}

func (h *_hugos) WriteConfig(c Config) error {
	b, err := os.ReadFile(h.configFile)
	if err != nil {
		return err
	}
	old := make(map[string]interface{})
	toml.Decode(string(b), &old)

	old["title"] = c.Title
	old["description"] = c.Description
	old["defaultContentLanguage"] = c.DefaultContentLanguage
	old["theme"] = c.Theme
	old["copyright"] = c.Copyright
	oldAuthor := old["author"].(map[string]interface{})
	oldAuthor["name"] = c.Author.Name
	old["author"] = oldAuthor

	buf := new(bytes.Buffer)
	toml.NewEncoder(buf).Encode(old)

	err = os.WriteFile(h.configFile, buf.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (h *_hugos) execHugoCmd(dir string, arg ...string) error {
	Logger.Info("exec hugo command", zap.Strings("arg", arg))
	cmd := exec.Command(h.hugo, arg...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		Logger.Error("exec hugo command error", zap.Strings("arg", arg), zap.Error(err))
		return err
	}
	Logger.Info("exec hugo command output", zap.Strings("arg", arg))
	return nil
}

func (h *_hugos) execHugoCmdStart(dir string, arg ...string) error {
	Logger.Info("exec hugo command", zap.Strings("arg", arg))
	cmd := exec.Command(h.hugo, arg...)
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		Logger.Error("exec hugo command error", zap.Strings("arg", arg), zap.Error(err))
		return err
	}
	Logger.Info("exec hugo command output", zap.Strings("arg", arg))
	return nil
}

func (h *_hugos) SplitMetaAndContent(article string) (meta string, content string) {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(article))
	for sc.Scan() {
		lines = append(lines, sc.Text())
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
