package localcache

import (
	"fmt"
	"sync"
	"time"

	localctx "github.com/superops-team/hyperops/pkg/ops/context"
	"github.com/superops-team/hyperops/pkg/ops/util"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const Name = "localcache"
const ModuleName = "localcache.star"

var (
	once     sync.Once
	instance *CacheManager
)

var Module = &starlarkstruct.Module{
	Name: "localcache",
	Members: starlark.StringDict{
		"new": localctx.AddBuiltin("localcache.new", NewLocalCache),
	},
}

// localCache singleton cache
type CacheManager struct {
	cacheStore sync.Map
}

func (c *CacheManager) GetByDir(dir string) *LocalCache {
	key := dir
	if dir == "" {
		key = "__cache_in_mem__"
	}
	l, ok := c.cacheStore.Load(key)
	if ok {
		lc, _ := l.(*LocalCache)
		return lc
	}
	if key != dir {
		store, _ := NewMemStore()
		lc := &LocalCache{
			cachedir: dir,
			cachedb:  store,
		}
		c.cacheStore.Store(dir, lc)
		return lc
	}
	store, _ := NewStore(dir)
	lc := &LocalCache{
		cachedir: dir,
		cachedb:  store,
	}
	c.cacheStore.Store(key, lc)
	return lc
}

type LocalCache struct {
	cachedir string
	cachedb  *Store
}

func (l *LocalCache) Struct() *starlarkstruct.Struct {
	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"set":          localctx.AddBuiltin("localcache.set", l.Set),
		"set_with_ttl": localctx.AddBuiltin("localcache.set_with_ttl", l.SetWithTTL),
		"get":          localctx.AddBuiltin("localcache.get", l.Get),
		"delete":       localctx.AddBuiltin("localcache.delete", l.Delete),
		"filter":       localctx.AddBuiltin("localcache.filter", l.Filter),
		"filter_key":   localctx.AddBuiltin("localcache.filter_key", l.FilterKey),
		"clear":        localctx.AddBuiltin("localcache.clear", l.Clear),
		"exist":        localctx.AddBuiltin("localcache.exist", l.Exist),
	})
}

// NewLocalCache 默认磁盘cache, 如果开启内存cache需要memory=True
func NewLocalCache(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	cacheDir, err := params.GetString(0)
	if err != nil {
		cacheDir, err = params.GetStringByName("dir")
		if err != nil {
			cacheDir = fmt.Sprintf("%s/cache", thread.Name)
		}
	}
	return getLocalCache(cacheDir).Struct(), nil
}

func getLocalCache(cachedir string) *LocalCache {
	dir := cachedir
	if cachedir != "" {
		dir = util.EnsureWorkdir(cachedir)
	}
	once.Do(func() {
		instance = &CacheManager{
			cacheStore: sync.Map{},
		}
	})
	return instance.GetByDir(dir)
}

func (l *LocalCache) Set(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	key, err := params.GetString(0)
	if err != nil {
		key, err = params.GetStringByName("key")
		if err != nil {
			return starlark.None, err
		}
	}
	val, err := params.GetString(1)
	if err != nil {
		val, err = params.GetStringByName("val")
		if err != nil {
			return starlark.None, err
		}
	}
	if l.cachedb == nil {
		return starlark.None, err
	}

	err = l.cachedb.Set([]byte(key), []byte(val))
	if err != nil {
		return starlark.Bool(false), err
	}
	return starlark.Bool(true), nil
}

func (l *LocalCache) SetWithTTL(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	key, err := params.GetString(0)
	if err != nil {
		key, err = params.GetStringByName("key")
		if err != nil {
			return starlark.None, err
		}
	}
	val, err := params.GetString(1)
	if err != nil {
		val, err = params.GetStringByName("val")
		if err != nil {
			return starlark.None, err
		}
	}
	ttl, err := params.GetString(2)
	if err != nil {
		ttl, err = params.GetStringByName("ttl")
		if err != nil {
			return starlark.None, err
		}
	}
	duration, err := time.ParseDuration(ttl)
	if err != nil {
		return starlark.None, err
	}
	if l.cachedb == nil {
		return starlark.None, err
	}

	err = l.cachedb.SetWithTTL([]byte(key), []byte(val), duration)
	if err != nil {
		return starlark.None, err
	}
	return starlark.Bool(true), nil
}

func (l *LocalCache) Get(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	key, err := params.GetString(0)
	if err != nil {
		key, err = params.GetStringByName("key")
		if err != nil {
			return starlark.None, err
		}
	}
	if l.cachedb == nil {
		return starlark.None, err
	}
	if !l.cachedb.Exist([]byte(key)) {
		return starlark.None, nil
	}
	val, err := l.cachedb.Get([]byte(key))
	if err != nil {
		return starlark.None, err
	}
	return starlark.String(string(val)), nil
}

func (l *LocalCache) Delete(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	key, err := params.GetString(0)
	if err != nil {
		key, err = params.GetStringByName("key")
		if err != nil {
			return starlark.None, err
		}
	}
	if l.cachedb == nil {
		return starlark.None, err
	}

	err = l.cachedb.Delete([]byte(key))
	if err != nil {
		return starlark.None, err
	}
	return starlark.Bool(true), nil
}

func (l *LocalCache) Exist(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	key, err := params.GetString(0)
	if err != nil {
		key, err = params.GetStringByName("key")
		if err != nil {
			return starlark.None, err
		}
	}
	if l.cachedb == nil {
		return starlark.None, err
	}

	flag := l.cachedb.Exist([]byte(key))
	return starlark.Bool(flag), nil
}

func (l *LocalCache) Filter(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	prefix, err := params.GetString(0)
	if err != nil {
		prefix, err = params.GetStringByName("prefix")
		if err != nil {
			return starlark.None, err
		}
	}
	if l.cachedb == nil {
		return starlark.None, err
	}

	data, err := l.cachedb.Filter([]byte(prefix))
	if err != nil {
		return starlark.None, err
	}
	datakv := starlark.NewDict(len(data))
	for key, val := range data {
		vals := string(val)
		_ = datakv.SetKey(starlark.String(key), starlark.String(vals))
	}
	return datakv, nil
}

func (l *LocalCache) FilterKey(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	prefix, err := params.GetString(0)
	if err != nil {
		prefix, err = params.GetStringByName("prefix")
		if err != nil {
			return starlark.None, err
		}
	}
	if l.cachedb == nil {
		return starlark.None, err
	}

	data, err := l.cachedb.FilterKey([]byte(prefix))
	if err != nil {
		return starlark.None, err
	}
	if data == nil {
		return starlark.None, nil
	}
	datalist := []starlark.Value{}
	for _, val := range data {
		vals := string(val)
		datalist = append(datalist, starlark.String(vals))
	}
	return starlark.NewList(datalist), nil
}

func (l *LocalCache) Clear(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	prefix, err := params.GetString(0)
	if err != nil {
		prefix, err = params.GetStringByName("prefix")
		if err != nil {
			return starlark.None, err
		}
	}
	if l.cachedb == nil {
		return starlark.None, err
	}
	err = l.cachedb.Clear([]byte(prefix))
	if err != nil {
		return starlark.None, err
	}
	return starlark.Bool(true), nil
}
