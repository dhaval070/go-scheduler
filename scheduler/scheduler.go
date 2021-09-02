package scheduler
import (
    "os"
    "fmt"
    "gsch/db"
    "time"
    "log"
    "sync"
    "gsch/model/event"
    "gsch/hredis"
)

type DueEvents struct {
    stopDue []event.SchEvent
    startDue []event.SchEvent
}

type Scheduler struct {
    league string
    dbname string
    wseCmd map[int]map[string]event.Cmd // locId => camera => event.Cmd
    wseCmdMutex sync.Mutex
}

func NewScheduler(league string) *Scheduler {
    var sch Scheduler
    sch.league = league
    sch.dbname = "gos_" + league
    return &sch
}

func (sch *Scheduler) Work() {
    var err error
    var dueEvents *DueEvents
    var wg sync.WaitGroup

    if dueEvents, err = sch.GetEvents(); err != nil {
        log.Println(err)
        return
    }
    log.Println(dueEvents)

    wg.Add(len(dueEvents.stopDue))

    for _, ev := range dueEvents.stopDue {
        go sch.stopEvent(&ev, &wg)
    }
    wg.Wait()

    wg.Add(len(dueEvents.startDue))

    for _, ev := range dueEvents.startDue {
        go sch.startEvent(&ev, &wg)
    }
    wg.Wait()
}

func (sch *Scheduler) stopEvent(ev *event.SchEvent, wg *sync.WaitGroup) {
    defer wg.Done()

    if ev.ManualSchedule && ev.ScheduleSignal != "stop" {
        return
    }
    log.Println(ev.Id)
    db.Exec("update "+sch.dbname+".event set status = 2 where id = ?", ev.Id)
    // set stopping flag in hRedis
    // loc redis set status 2

}

func (sch *Scheduler) startEvent(ev *event.SchEvent, wg *sync.WaitGroup) {
    defer wg.Done()

    defer func() {
        if r := recover(); r != nil {
            log.Println(r)
        }
    }()
    if ev.ManualSchedule && ev.ScheduleSignal != "start" {
        return
    }

    var streams = event.GetEventStreams(sch.dbname, ev.Id)

    if len(streams) == 0 {
        return
    }

    if os.Getenv("CHECK_CURRENT_STREAM") != "" {
        // TODO
    }

    if !sch.lockLocation(ev.Id, ev.GlobalId) {
        return
    }

    log.Printf("start event #%d", ev.Id)

    var wend = "success"

    hredis.WowzaStarting(sch.league, ev.Id)
    sch.loadWseCmd(ev.LocationId)
    // if address != "" and overlayvisible then set overlayconf and run camPreset

    for _, stream := range streams {
        if stream.Broadcast {
            if !sch.broadcast(ev, stream.Camera, stream.AltStream) {
                wend = "error"
            }
        }

        if stream.Record {
            if !sch.record(ev, stream.Camera) {
                wend = "error"
            }
        }
    }
    hredis.WowzaDone(sch.league, ev.Id, wend)

    db.Exec("update "+sch.dbname+".event set status = 1 where id = ?", ev.Id)
    // if target id then update related db
    // end of start event
}

func (sch *Scheduler) loadWseCmd(locId int) (map[string]event.Cmd) {
    sch.wseCmdMutex.Lock()
    defer sch.wseCmdMutex.Unlock()

    if sch.wseCmd == nil {
        rows := event.FindCmd(sch.dbname, locId)

        sch.wseCmd = make(map[int]map[string]event.Cmd)

        for _, row := range rows {
            if sch.wseCmd[locId] == nil {
                sch.wseCmd[locId] = make(map[string]event.Cmd)
            }
            sch.wseCmd[locId][row.Camera] = row
        }
    }

    return sch.wseCmd[locId]
}

// retuns stop due and start due events
func (sch *Scheduler) GetEvents() (*DueEvents, error) {
    var stopEv, startEv []event.SchEvent
    now := time.Now().Unix()

    stopEv = sch.queryEvents(fmt.Sprintf(` where e.start < %d and e.status = 1 group by
        e.id, ex.league, elv.local_vod_name`, now))

    start := now + 60
    end := now

    startEv = sch.queryEvents(fmt.Sprintf(` where e.start <= %d and e.end > %d and e.status = 0 group by e.id, ex.league, elv.local_vod_name`, start, end))

    return &DueEvents{ stopEv, startEv }, nil
}

func (sch *Scheduler) queryEvents(query string) ([]event.SchEvent) {
    rows, done := db.Query(selectQry(sch.league) + query)
    defer done()

    var result []event.SchEvent

    for rows.Next() {
        var ev event.SchEvent

        ev.Scan(rows)
        result = append(result, ev)
    }
    return result
}

// retuns common select query
func selectQry(league string) string {
    var db = "gos_" + league

    return `SELECT "` + league + `" as league,
        e.id, e.start, e.end, e.location_id,
        ifnull(e.manual_schedule, 0) manual_schedule,
        ifnull(e.schedule_signal, "") schedule_signal,
        e.overlay_visible,
        e.sport,
        e.dir,
        elv.local_vod_name,
        t1.name team1,
        ifnull(t1.logo_file, "") t1_logo_file,
        ifnull(t1.short_name, "") t1_short_name,
        t2.name team2,
        ifnull(t2.logo_file, "") t2_logo_file,
        ifnull(t2.short_name, "") t2_short_name,
        ex.league AS target_league,
        ex.target_id,
        d.flood,
        sport.nevco_code,
        l.location, l.stream,
        ifnull(l.address, "") address,
        ifnull(l.port, "") port,
        ifnull(l.redis_port, "") redis_port, l.inmediate or e.inmediate AS loc_copy,
        global_id,
        ops.copy_method,
        use_rclone
    FROM `+db+`.event e INNER JOIN gos_`+league+`.location l ON e.location_id = l.id
        INNER JOIN `+db+`.team t1 ON t1.id = e.team_id1
        INNER JOIN `+db+`.team t2 ON t2.id = e.team_id2
        INNER JOIN `+db+`.sport sport ON sport.name = e.sport
        INNER JOIN `+db+`.division d ON d.id = e.division_id
        LEFT JOIN ops.location ops ON ops.id = l.global_id
        LEFT JOIN `+db+`.event_export ex ON ex.event_id = e.id
        LEFT JOIN `+db+`.event_local_vod elv ON elv.event_id = e.id`
}
