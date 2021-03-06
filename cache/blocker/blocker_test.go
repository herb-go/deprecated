package blocker

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/herb-go/deprecated/cache"
	_ "github.com/herb-go/deprecated/cache/drivers/syncmapcache"
)

func newTestCache(ttl int64) *cache.Cache {
	c := cache.New()
	oc := cache.NewOptionConfig()
	oc.Driver = "syncmapcache"
	oc.TTL = ttl * int64(time.Second)
	oc.Config = nil
	oc.Marshaler = "json"
	err := c.Init(oc)
	if err != nil {
		panic(err)
	}
	err = c.Flush()
	if err != nil {
		panic(err)
	}
	return c

}
func testIdentifier(r *http.Request) (string, error) {
	return r.Header.Get("name"), nil
}
func TestBlock(t *testing.T) {
	var rep *http.Response
	var err error
	blocker := New(newTestCache(1 * 3600))
	blocker.Identifier = testIdentifier
	blocker.Block(0, 20, 1*time.Hour)
	blocker.Block(404, 5, 1*time.Hour)
	blocker.Block(403, 5, 1*time.Hour)
	mux := http.NewServeMux()
	mux.HandleFunc("/200", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		if err != nil {
			panic(err)
		}
	})
	mux.HandleFunc("/403", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(403), 403)
	})
	mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(404), 404)
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		blocker.ServeMiddleware(w, r, mux.ServeHTTP)
	}))
	defer server.Close()
	req, err := http.NewRequest("get", server.URL+"/403", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("name", "test1")
	for i := 0; i < 5; i++ {
		rep, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		rep.Body.Close()
		if rep.StatusCode != 403 {
			t.Error(rep.StatusCode)
		}
		time.Sleep(10 * time.Millisecond)
	}
	rep, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	rep.Body.Close()
	if rep.StatusCode != 429 {
		t.Error(rep.StatusCode)
	}
	time.Sleep(10 * time.Millisecond)
	req, err = http.NewRequest("get", server.URL+"/403", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("name", "test2")
	rep, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	rep.Body.Close()
	if rep.StatusCode != 403 {
		t.Error(rep.StatusCode)
	}
	time.Sleep(10 * time.Millisecond)
	blocker.StatusCodeBlocked = 400
	req, err = http.NewRequest("get", server.URL+"/404", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("name", "test2")
	for i := 0; i < 5; i++ {
		rep, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		rep.Body.Close()
		if rep.StatusCode != 404 {
			t.Error(rep.StatusCode)
		}
		time.Sleep(10 * time.Millisecond)
	}
	rep, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	rep.Body.Close()
	if rep.StatusCode != 400 {
		t.Error(rep.StatusCode)
	}
	time.Sleep(10 * time.Millisecond)
	blocker.StatusCodeBlocked = defaultBlockedStatus
	req, err = http.NewRequest("get", server.URL+"/403", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("name", "test3")
	for i := 0; i < 4; i++ {
		rep, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		rep.Body.Close()
		if rep.StatusCode != 403 {
			t.Error(rep.StatusCode)
		}
		time.Sleep(10 * time.Millisecond)
	}

	req, err = http.NewRequest("get", server.URL+"/404", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("name", "test3")
	for i := 0; i < 4; i++ {
		rep, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		rep.Body.Close()
		if rep.StatusCode != 404 {
			t.Error(rep.StatusCode)
		}
		time.Sleep(10 * time.Millisecond)
	}

	req, err = http.NewRequest("get", server.URL+"/200", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("name", "test3")
	for i := 0; i < 12; i++ {
		rep, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		rep.Body.Close()
		if rep.StatusCode != 200 {
			t.Error(rep.StatusCode)
		}
		time.Sleep(10 * time.Millisecond)
	}
	rep, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	rep.Body.Close()
	if rep.StatusCode != 429 {
		t.Error(rep.StatusCode)
	}
}

func TestAnyError(t *testing.T) {
	var rep *http.Response
	var req *http.Request
	var err error
	blocker := New(newTestCache(1 * 3600))
	blocker.Block(StatusAnyError, 5, 1*time.Hour)
	mux := http.NewServeMux()
	mux.HandleFunc("/200", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		if err != nil {
			panic(err)
		}
	})
	mux.HandleFunc("/403", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(403), 403)
	})
	mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(404), 404)
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		blocker.ServeMiddleware(w, r, mux.ServeHTTP)
	}))
	defer server.Close()
	req, err = http.NewRequest("get", server.URL+"/403", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("name", "test3")
	for i := 0; i < 2; i++ {
		rep, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		rep.Body.Close()
		if rep.StatusCode != 403 {
			t.Error(rep.StatusCode)
		}
		time.Sleep(10 * time.Millisecond)
	}

	req, err = http.NewRequest("get", server.URL+"/404", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("name", "test3")
	for i := 0; i < 3; i++ {
		rep, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		rep.Body.Close()
		if rep.StatusCode != 404 {
			t.Error(rep.StatusCode)
		}
		time.Sleep(10 * time.Millisecond)
	}
	rep, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	rep.Body.Close()
	if rep.StatusCode != 429 {
		t.Error(rep.StatusCode)
	}

}
func TestIPIdentifier(t *testing.T) {
	var rep *http.Response
	var err error
	blocker := New(newTestCache(1 * 3600))
	blocker.Block(StatusAny, 20, 1*time.Hour)
	blocker.Block(404, 5, 1*time.Hour)
	blocker.Block(403, 5, 1*time.Hour)
	mux := http.NewServeMux()
	mux.HandleFunc("/200", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		if err != nil {
			panic(err)
		}
	})
	mux.HandleFunc("/403", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(403), 403)
	})
	mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(404), 404)
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		blocker.ServeMiddleware(w, r, mux.ServeHTTP)
	}))
	defer server.Close()
	req, err := http.NewRequest("get", server.URL+"/403", nil)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 5; i++ {
		rep, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		rep.Body.Close()
		if rep.StatusCode != 403 {
			t.Error(rep.StatusCode)
		}
		time.Sleep(10 * time.Millisecond)
	}
	rep, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	rep.Body.Close()
	if rep.StatusCode != 429 {
		t.Error(rep.StatusCode)
	}
	time.Sleep(10 * time.Millisecond)
}
