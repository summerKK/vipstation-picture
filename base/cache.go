package base

import (
	"sync"
	"fmt"
)

//图片缓存
type ImgCache interface {
	//将请求放入请求缓存
	Put(resouce map[string]string) bool
	//从请求缓存中获最早被放入且人在其中的请求
	Get() map[string]string
	//获得请求缓存的容量
	Capacity() int
	//获得请求缓存的实时长度,即其中的请求的即时数量
	Length() int
	//关闭请求缓存
	Close()
	//获取请求缓存的摘要信息
	Summary() string
}

type myImgCache struct {
	cache []map[string]string
	mutex sync.Mutex
	//代表请求状态0代表初始化,1代表关闭
	status byte
}

var (
	summaryTemplate = "status:%s," + "length:%d," + "capacity:%d"
	//状态
	statusMap = map[byte]string{
		0: "running",
		1: "closed",
	}
)

//创建缓存请求
func NewCache() ImgCache {
	ic := &myImgCache{
		cache: make([]map[string]string, 0),
	}
	return ic
}

func (cache *myImgCache) Put(source map[string]string) bool {

	if len(source) == 0 {
		return false
	}

	if cache.status == 1 {
		return false
	}

	cache.mutex.Lock()
	cache.cache = append(cache.cache, source)
	cache.mutex.Unlock()

	return true
}

func (cache *myImgCache) Get() map[string]string {
	if cache.Length() == 0 {
		return nil
	}

	if cache.status == 1 {
		return nil
	}

	cache.mutex.Lock()
	source := cache.cache[0]
	cache.cache = cache.cache[1:]
	cache.mutex.Unlock()

	return source
}

func (cache *myImgCache) Capacity() int {
	return cap(cache.cache)
}

func (cache *myImgCache) Length() int {
	return len(cache.cache)
}

func (cache *myImgCache) Close() {
	if cache.status == 1 {
		return
	}
	cache.status = 1
}

func (cache *myImgCache) Summary() string {
	summary := fmt.Sprintf(summaryTemplate,
		statusMap[cache.status],
		cache.Length(),
		cache.Capacity())
	return summary
}
