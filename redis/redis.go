package rclient
import (
    "github.com/go-redis/redis/v8"
    "sync"
)

var pool map[string]*redis.Client
var connMutex sync.Mutex

func init() {
    pool = make(map[string]*redis.Client)
}

type Options redis.Options

func GetClient(options Options) *redis.Client {
    connMutex.Lock()
    defer connMutex.Unlock()

    if pool[options.Addr] != nil {
        return pool[options.Addr]
    }

    var opt = redis.Options(options)
    var client = redis.NewClient(&opt)

    pool[options.Addr] = client
    return client
}
