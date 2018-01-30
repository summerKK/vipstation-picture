package main

import (
	"vipstation-picture/database"
	"fmt"
	"strings"
	"vipstation-picture/config"
	"regexp"
	"log"
	"strconv"
	"os"
	"net/http"
	"time"
	"io/ioutil"
	"io"
	"bytes"
	"sync"
	"sort"
)

var (
	saveDir string = config.Config.SaveDir
)

func main() {
	connection := database.NewDatabase()
	connection.GetProducts()
	connection.ReceiveMediaGallery()
	sign := make(chan struct{})

	go func() {
		for {
			if source := connection.RowCache.Get(); source != nil {
				mediaGallery := handlePicture(source)
				connection.UpdateChan <- map[string]string{"img": mediaGallery, "sku": source["sku"]}
			} else {
				break
			}
		}
		//关闭接收channel
		close(connection.UpdateChan)
		sign <- struct{}{}
	}()

	<-sign

	fmt.Println("Done")
}

func handlePicture(source map[string]string) string {
	imgs := strings.Split(source["imgs"], ";")
	imgDir := getDirPath(source["sku"])
	imgPaths := make([]string, 0, len(imgs))
	var wg sync.WaitGroup
	var mu sync.Mutex
	for index, img := range imgs {
		wg.Add(1)
		go func(img string, index int) {
			s := downloadImg(img, imgDir, index, 3)
			mu.Lock()
			imgPaths = append(imgPaths, s)
			mu.Unlock()
			wg.Done()
		}(img, index)
	}
	wg.Wait()

	sort.Strings(imgPaths)
	return strings.Join(imgPaths, ";")
}

func getDirPath(subDir string) string {

	path := saveDir + subDir + "/"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Fatalf("创建%s文件夹失败!%s", path, err)
		}
	}
	return path
}

func downloadImg(url string, path string, order int, retry int) string {
	regexFile := regexp.MustCompile(`(?i)(\w|\d|_)*\.(jpg|jpeg|png|gif)`)
	fileName := regexFile.FindString(url)
	if len(fileName) == 0 {
		log.Println("img url解析失败!", url)
		return ""
	}

	if order >= 0 {
		fileName = strconv.Itoa(order) + "_" + fileName
	}

	//判断是否超过指定长度
	if len(fileName) > 70 {
		rs := []rune(fileName)

		fileName = string(rs[:70]) + string(rs[strings.LastIndex(fileName, "."):])
	}

	//判断文件是否已经下载过
	if fileInfo, err := os.Stat(path + fileName); os.IsNotExist(err) || fileInfo.Size() < 5120 {

		//文件小于5120b
		if err == nil {
			os.Remove(path + fileName)
		}

		resp, err := http.Get(url)
		//是否需要关闭资源
		if err == nil {
			defer resp.Body.Close()
		}

		if err != nil || resp.StatusCode != 200 {
			time.Sleep(time.Second)
			if retry > 0 {
				fmt.Printf("下载图片失败!重试中(%d)...", retry)
				return downloadImg(url, path, order, retry-1)
			} else {
				fmt.Println("下载图片失败!重试3次")
				return ""
			}
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			out, _ := os.Create(path + fileName)
			io.Copy(out, bytes.NewReader(body))
			out.Close()
			os.Chmod(path+fileName, os.ModePerm)
			return path + fileName
		}

	} else {
		return path + fileName
	}
}
