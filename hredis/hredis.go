/*
This module is used to push the event start/stop status to redis which is used by ops application
*/
package hredis
import (
    "log"
    _ "gsch/dotenv"
    "github.com/go-redis/redis/v8"
    "context"
    "os"
    "fmt"
)

var hRedis *redis.Client
var ctx = context.Background()

func init() {
    hRedis = redis.NewClient(&redis.Options {
        Addr: os.Getenv("REDIS_HOST"),
    })
}

func WowzaStarting(league string, eventId int) {
    var err error
    err = hRedis.LPush(ctx, "event-msg", fmt.Sprintf("wowza-start#%s%d", league, eventId)).Err()

    if err != nil {
        log.Println("redis error " + err.Error())
    }

    err = hRedis.HSet(ctx, fmt.Sprintf("sh-%s-%d", league, eventId), "wowza", "running").Err()

    if err != nil {
        log.Println("redis error " + err.Error())
    }
}

func WowzaDone(league string, eventId int, outcome string) {
    var err error

    err = hRedis.LPush(ctx, "event-msg", fmt.Sprintf("wowza-done#%s%d@%s", league, eventId, outcome)).Err()

    if err != nil {
        log.Println("redis error " + err.Error())
    }

    err = hRedis.HSet(ctx, fmt.Sprintf("sh-%s-%d", league, eventId), "wowza", outcome).Err()

    if err != nil {
        log.Println("redis error " + err.Error())
    }
}
