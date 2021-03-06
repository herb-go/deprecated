package cache_test

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/herb-go/deprecated/cache/drivers/syncmapcache"

	"github.com/herb-go/deprecated/cache"
)

func NewEntiyTooLargeCache(ttl int64) cache.Cacheable {
	config := syncmapcache.Config{
		Size: 1,
	}
	buf := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buf)
	decoder := json.NewDecoder(buf)
	err := encoder.Encode(config)
	if err != nil {
		panic(err)
	}
	c := cache.New()
	oc := cache.NewOptionConfig()
	oc.Driver = "syncmapcache"
	oc.TTL = int64(ttl)
	oc.Config = decoder.Decode
	oc.Marshaler = "json"

	err = c.Init(oc)
	if err != nil {
		panic(err)
	}
	err = c.Flush()
	if err != nil {
		panic(err)
	}
	return c

}
func newTestCache(ttl int64) *cache.Cache {
	config := syncmapcache.Config{
		Size: 10000000,
	}
	buf := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buf)
	decoder := json.NewDecoder(buf)
	err := encoder.Encode(config)
	if err != nil {
		panic(err)
	}
	c := cache.New()
	oc := cache.NewOptionConfig()
	oc.Driver = "syncmapcache"
	oc.TTL = int64(ttl)
	oc.Config = decoder.Decode
	oc.Marshaler = "json"

	err = c.Init(oc)
	if err != nil {
		panic(err)
	}
	err = c.Flush()
	if err != nil {
		panic(err)
	}
	return c
}

var testLoader = func(key string) (interface{}, error) {
	return key, nil
}

func TestCacheEmptyKey(t *testing.T) {
	var err error
	c := newTestCache(3600)
	var data = []byte{}
	err = c.Set("", data, 0)
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	err = c.Update("", data, 0)
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	err = c.Get("", data)
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	err = c.SetBytesValue("", data, 0)
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	err = c.UpdateBytesValue("", data, 0)
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	_, err = c.GetBytesValue("")
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	err = c.Del("")
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	err = c.SetCounter("", 0, 0)
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}

	_, err = c.GetCounter("")
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	_, err = c.IncrCounter("", 0, 0)
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	err = c.DelCounter("")
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	err = c.Expire("", 1000)
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	err = c.ExpireCounter("", 1000)
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
	err = c.Load("", nil, 0, testLoader)
	if err != cache.ErrKeyUnavailable {
		t.Fatal(err)
	}
}
func TestLoader(t *testing.T) {
	var err error
	var result string
	c := newTestCache(3600)
	result = ""
	err = c.Load("test", &result, 0, testLoader)
	if err != nil {
		t.Fatal(err)
	}
	if result != "test" {
		t.Fatal(result)
	}
	err = c.Load("test", &result, 0, testLoader)
	if err != nil {
		t.Fatal(err)
	}
	if result != "test" {
		t.Fatal(result)
	}
}

func TestLoaderEntiyTooLarge(t *testing.T) {
	var err error
	var result string
	c := NewEntiyTooLargeCache(3600)
	result = ""
	err = c.Load("test", &result, 0, testLoader)
	if err != nil {
		t.Fatal(err)
	}
	if result != "test" {
		t.Fatal(result)
	}
	err = c.Load("test", &result, 0, testLoader)
	if err != nil {
		t.Fatal(err)
	}
	if result != "test" {
		t.Fatal(result)
	}
}

func TestNotFound(t *testing.T) {
	var err error
	c := newTestCache(3600)
	result := []byte{}
	err = c.Expire("notexists", 0)
	if err != nil {
		t.Fatal(err)
	}
	err = c.ExpireCounter("notexists", 0)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Update("notexists", result, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Get("notexists", &result)
	if err != cache.ErrNotFound {
		t.Fatal(err)
	}
	err = c.UpdateBytesValue("notexists", result, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetBytesValue("notexists")
	if err != cache.ErrNotFound {
		t.Fatal(err)
	}
	err = c.Del("notexists")
	if err != nil {
		t.Fatal(err)
	}
	err = c.DelCounter("notexists")
	if err != nil {
		t.Fatal(err)
	}
}

func TestFinalKey(t *testing.T) {
	c := newTestCache(3600)
	k := c.FinalKey("key")
	if k != cache.KeyPrefix+"key" {
		t.Fatal(k)
	}
}

func TestMutliConcurrent(t *testing.T) {
	c := newTestCache(3600)
	wg := &sync.WaitGroup{}
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var result = ""
			key := strconv.Itoa(rand.Intn(10))
			err := c.Load(key, &result, 0, testLoader)
			if err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()
	locker, _ := c.Util().Locker("test")
	locker.Lock()
	locker.Unlock()
	locker.Map.Range(func(key interface{}, value interface{}) bool {
		//check unlocked locker
		t.Fatal(key, value)
		return true
	})

}
