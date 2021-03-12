package main

import (
	"bytes"
	"github.com/juju/zip"
	"github.com/wailovet/nuwa"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"log"
	"path"
	"runtime"
	"strings"
)

func utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Zip(id string, dst, src string) (err error) {
	var fw *bytes.Buffer
	var zw *zip.Writer

	fw = bytes.NewBuffer(nil)

	// 通过 fw 来创建 zip.Write
	zw = zipNewWriter(fw)

	// 创建准备写入的文件
	if nuwa.Helper().PathExists(dst) {
		data, _ := ioutil.ReadFile(dst)
		ofw := bytes.NewBuffer(data)

		// read it back
		r, err := zip.NewReader(bytes.NewReader(ofw.Bytes()), int64(ofw.Len()))

		if err != nil {
			return err
		}
		//log.Println(nuwa.Helper().JsonEncode(r.File))
		zw = r.Append(fw)
		srcBuf, _ := ioutil.ReadFile(src)

		zipSrcPath := strings.Replace(src, `\`, "/", -1)
		zipSrcPath = strings.Replace(zipSrcPath, `:`, "", -1)
		zipSrcPath = path.Join(id, zipSrcPath)
		if runtime.GOOS == "windows" {
			zipSrcPath_, err := utf8ToGbk([]byte(zipSrcPath))
			if err != nil {
				log.Println(err)
			} else {
				zipSrcPath = string(zipSrcPath_)
			}
		}

		err = zipCreate(zw, zipSrcPath, srcBuf)
		if err != nil {
			return err
		}

		allBytes := append(ofw.Bytes(), fw.Bytes()...)

		err = ioutil.WriteFile(dst, allBytes, 0644)
		if err != nil {
			log.Println("WriteFile:", err)
			return err
		}
		return err
	}

	srcBuf, _ := ioutil.ReadFile(src)

	zipSrcPath := strings.Replace(src, `\`, "/", -1)
	zipSrcPath = strings.Replace(zipSrcPath, `:`, "", -1)
	zipSrcPath = path.Join(id, zipSrcPath)

	f, err := zw.Create(zipSrcPath)
	if err != nil {
		log.Println("f, err := zw.Create(src):", err)
		return
	}
	n, err := f.Write(srcBuf)
	log.Println("n = ", n)
	if err != nil {
		log.Println("_, err = f.Write(srcBuf):", err)
		return
	}

	_ = zw.Close()

	err = ioutil.WriteFile(dst, fw.Bytes(), 0644)
	if err != nil {
		log.Println("WriteFile:", err)
		return
	}
	return

}

func zipNewWriter(w io.Writer) *zip.Writer {
	znw := zip.NewWriter(w)

	return znw
}

func zipCreate(w *zip.Writer, name string, data []byte) (err error) {
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	}
	header.SetMode(0755)

	f, err := w.CreateHeader(header)
	if err != nil {
		return
	}
	_, err = f.Write(data)
	if err != nil {
		return
	}
	w.Close()
	return
}
