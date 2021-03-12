package main

import (
	"flag"
	"fmt"
	"github.com/wailovet/gofunc"
	"github.com/wailovet/nuwa"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func search(root []string, exclude string, filter string, day int, sizeKb int, do func(dir string, info os.FileInfo)) {
	s := searchLoger{
		Root: root,
	}
	s.Start(func(dir string, info os.FileInfo) {
		if strings.HasSuffix(info.Name(), ".log") {
			if exclude != "" {
				if strings.Index(info.Name(), exclude) > -1 || strings.Index(dir, exclude) > -1 {
					return
				}
			}

			if filter != "" {
				if strings.Index(info.Name(), filter) < 0 {
					return
				}
			}

			if info.Size() < 10 {
				return
			}

			if info.ModTime().Before(time.Now().AddDate(0, 0, -day)) {
				do(dir, info)
				log.Println("已经过期:", dir, info.Name(), info.ModTime().Format("2006-01-02 15:04:05"), formatFileSize(info.Size()))
				return
			}

			if float64(info.Size())/float64(1024) > float64(sizeKb) {
				do(dir, info)
				log.Println("文件过大:", dir, info.Name(), info.ModTime().Format("2006-01-02 15:04:05"), formatFileSize(info.Size()))
				return
			}
		}
	})
	log.Println("搜索完成")
}

func formatFileSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}

var roots []string
var exclude string
var filter string
var size int
var day int
var bakZipPath string

func searchCleanBakTask() {

	gofunc.New(func() {
		for {
			entries, _ := ioutil.ReadDir(bakZipPath)
			for e := range entries {
				if strings.HasSuffix(entries[e].Name(), ".zip") {
					if entries[e].ModTime().Before(time.Now().AddDate(0, -1, 0)) {
						log.Println("备份文件已过期", entries[e].Name())
						filePath := filepath.Join(bakZipPath, entries[e].Name())
						_ = os.Remove(filePath)
					}
				}
			}

			time.Sleep(time.Hour * 24)
		}

	}).Catch(func(i interface{}) {
		log.Println("协程2异常:", i)
		time.Sleep(time.Second * 5)
		searchCleanBakTask()
	})
}
func searchCleanLogTask() {
	gofunc.New(func() {
		for {
			id := time.Now().Format("2006_01_02__15_04_05")

			search(roots, exclude, filter, day, size, func(dir string, info os.FileInfo) {
				filePath := filepath.Join(dir, info.Name())

				if !nuwa.Helper().PathExists(bakZipPath) {
					_ = os.MkdirAll(bakZipPath, 0644)
				}

				bigZipFilePath := filepath.Join(bakZipPath, fmt.Sprint(id, ".", info.Name(), ".big.zip"))
				bakZipFilePath := filepath.Join(bakZipPath, fmt.Sprint(id, ".zip"))

				if info.Size() > 80*1024*1024 {
					log.Println("压缩至:", bigZipFilePath)
					_, err := exec.Command("/usr/bin/zip", "-r", "-9", bigZipFilePath, filePath).CombinedOutput()
					if err != nil {
						log.Println("bash zip error:", err)
					} else {
						_ = ioutil.WriteFile(filePath, []byte{}, info.Mode().Perm())
					}
					return
				} else {
					log.Println("压缩至:", bakZipFilePath)
					_, err := exec.Command("/usr/bin/zip", "-r", "-9", bakZipFilePath, filePath).CombinedOutput()
					if err != nil {
						log.Println("bash zip error:", err)
					} else {
						_ = ioutil.WriteFile(filePath, []byte{}, info.Mode().Perm())
					}
					return
				}

				time.Sleep(time.Second / 10)
			})

			time.Sleep(time.Hour * 8)
		}
	}).Catch(func(i interface{}) {
		log.Println("协程异常:", i)
		time.Sleep(time.Second * 5)
		searchCleanLogTask()
	})
}

func main() {
	root := ""
	flag.StringVar(&root, "dir", "/root/", "查找日志目录")
	flag.StringVar(&exclude, "e", "", "字符串排除")
	flag.StringVar(&filter, "f", "", "字符串过滤")
	flag.StringVar(&bakZipPath, "bak", "./logs_bak", "压缩备份文件路径")
	flag.IntVar(&size, "s", 1024*200, "大小阈值/KB")
	flag.IntVar(&day, "d", 2, "过期天数")

	flag.Parse()

	roots_ := strings.Split(root, ",")
	for e := range roots_ {
		if roots_[e] != "" {
			roots = append(roots, roots_[e])
		}
	}

	searchCleanLogTask()
	//searchCleanBakTask()

	_ = nuwa.Http().Run()
}
