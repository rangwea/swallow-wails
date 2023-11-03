package backend

import (
	"fmt"
	"os"
	"testing"

	"go.uber.org/zap"
)

func before() {
	APP_HOME = ""
	if e, _ := PathExists(APP_HOME); !e {
		if err := os.Mkdir(APP_HOME, os.ModePerm); err != nil {
			Logger.Fatal("make app home dir fail", zap.Error(err))
		}
	}
	Zap.Initialize()
	Hugos.Initialize()
}

func TestNew(t *testing.T) {
	before()
	Hugos.NewSite()
}

func TestPreview(t *testing.T) {
	before()
	err := Hugos.Preview()
	if err != nil {
		fmt.Println(err)
	}
}

func TestWriteArticle(t *testing.T) {
	before()
	err := Hugos.WriteArticle("1", Meta{Title: "第一篇",
		Tags:        []string{"t1", "t2"},
		Description: "描述1",
		Date:        "2023-09-22 17:00:21",
		Lastmod:     "2023-09-22 17:00:21",
	},
		"哈哈哈，我的第一篇博客")
	fmt.Println(err)
}

func TestReadArticle(t *testing.T) {
	before()
	meta, content, err := Hugos.ReadArticle("1")
	fmt.Printf("%v\n%v\n%v\n", meta, content, err)
}

func TestSplitMetaAndContent(t *testing.T) {
	before()
	m, c := Hugos.SplitMetaAndContent(`+++
aaaa
+++
bbbb
cccc
dddd
	`)
	fmt.Printf("%v;%v\n", m, c)
}

func TestReadConfig(t *testing.T) {
	before()
	config, err := Hugos.ReadConfig()
	fmt.Printf("%v\n%v\n", config, err)
}

func TestWriteConfig(t *testing.T) {
	before()
	config := Config{
		Title:       "Title",
		Description: "Description",
		Theme:       "stack",
		Copyright:   "copyright",
		Author: &ConfigAuthor{
			Name: "wikia1",
		},
	}
	err := Hugos.WriteConfig(config)
	fmt.Printf("%v\n%v\n", config, err)
}

func TestGenerate(t *testing.T) {
	before()
	Hugos.Generate()
}
