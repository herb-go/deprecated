package cacheoverseer

import (
	"github.com/herb-go/deprecated/cache"
	"github.com/herb-go/worker"
)

var cacheworker = cache.New()
var Team = worker.GetWorkerTeam(&cacheworker)

func GetCacheByID(id string) cache.Cacheable {
	w := worker.FindWorker(id)
	if w == nil {
		return nil
	}
	c, ok := w.Interface.(**cache.Cache)
	if ok == false || c == nil {
		return nil
	}
	return *c
}
