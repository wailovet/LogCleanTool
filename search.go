package main

import (
	"github.com/wailovet/gofunc"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type searchLoger struct {
	Root  []string
	cb    func(dir string, info os.FileInfo)
	count int
}

func (sl *searchLoger) Start(cb func(dir string, info os.FileInfo)) {
	sl.cb = cb
	isOk := false
	gofunc.New(func() {
		for {
			if isOk {
				return
			}
			log.Println("进行中:", " 检索次数 ", sl.count)
			time.Sleep(10 * time.Second)
		}
	})

	for e := range sl.Root {
		log.Println("开始检索:", sl.Root[e])
		sl.walkDir(sl.Root[e])
		log.Println("检索结束:", sl.Root[e], " 检索次数", sl.count)
		sl.count = 0
	}
	isOk = true
}

//获取目录dir下的文件大小
func (sl *searchLoger) walkDir(dir string) {
	for _, entry := range sl.dirents(dir) {
		if entry.IsDir() { //目录
			subDir := filepath.Join(dir, entry.Name())
			sl.walkDir(subDir)
		} else {
			if strings.HasSuffix(entry.Name(), ".log") {
				sl.cb(dir, entry)
			}

		}
		sl.count++
	}
}

//读取目录dir下的文件信息
func (sl *searchLoger) dirents(dir string) []os.FileInfo {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}
	return entries
}
