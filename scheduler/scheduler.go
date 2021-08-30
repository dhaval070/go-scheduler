package scheduler
import (
    "os"
    "fmt"
    "database/sql"
    "gsch/db"
    "context"
    "time"
    "log"
    "sync"
    "gsch/model/event"
)

type DueEvents struct {
    stopDue []event.SchEvent
    startDue []event.SchEvent
}

type Scheduler struct {
    league string
    conn *sql.Conn
}

func NewScheduler(league string) *Scheduler {
    return &Scheduler { league, nil }
}

func (sch *Scheduler) getConnection() (*sql.Conn, error) {
    var conn *sql.Conn
    var ctx = context.Background()
    var err error
    conn, err = db.Db().Conn(ctx)

    if err != nil {
        return nil, err
    }

    if _, err := conn.ExecContext(ctx, "use gos_" + league); err != nil {
        return nil, err
    }
    return conn, nil
}

func (sch *Scheduler) Work() {
    defer func () {
        if sch.conn != nil {
            sch.conn.Close()
        }
    }()

    var err error
    if sch.conn, err = sch.getConnection(); err != nil {
        log.Println(err)
        return
    }

    log.Println("processing " + sch.league)
    var dueEvents *DueEvents
    var wg sync.WaitGroup

    if dueEvents, err = sch.GetEvents(); err != nil {
        log.Println(err)
        return
    }
    log.Println(dueEvents)

    for _, ev := range dueEvents.stopDue {
        if ev.ManualSchedule && ev.ScheduleSignal != "stop" {
            continue
        }
        wg.Add(1)
        go sch.stopEvent(&ev, &wg)
    }

    for _, ev := range dueEvents.startDue {
        if ev.ManualSchedule && ev.ScheduleSignal != "start" {
            continue
        }
        wg.Add(1)
        go sch.startEvent(&ev, &wg)
    }
    wg.Wait()
}

func (sch *Scheduler) stopEvent(ev *event.SchEvent, wg *sync.WaitGroup) {
    defer wg.Done()
    log.Println(ev.Id)
    // set stopping flag in hRedis
    // loc redis set status 2

}

func (sch *Scheduler) startEvent(ev *event.SchEvent, wg *sync.WaitGroup) {
    defer wg.Done()
    log.Printf("start event #%d", ev.Id)

    if os.Getenv("CHECK_CURRENT_STREAM" != "") {
        // TODO
    }

    // sch.LockLocation(ev)
    // hRedis update - starting the event
    // load wse for the loc id
    // if address != "" and overlayvisible then set overlayconf
    // load event streams and process
    // hRedis update - wowza status
    // update event.status = 1
    // if target id then update related db
    // end of start event
}

func (sch *Scheduler) eventStreams(id int) *sql.Rows {
    res, err := sch.conn.QueryContext(context.Background(), "select * from event_stream where event_id=?",
        id)

    if err != nil {
        panic(err)
    }
    return res
}

// retuns stop due and start due events
func (sch *Scheduler) GetEvents() (*DueEvents, error) {
    var err error

    var stopEv, startEv []event.SchEvent
    now := time.Now().Unix()

    stopEv, err = sch.queryEvents(fmt.Sprintf(` where e.start < %d and e.status = 1 group by
        e.id, ex.league, elv.local_vod_name`, now))

    if err != nil {
    log.Println(err)
        return nil, err
    }

    start := now + 60
    end := now

    startEv, err = sch.queryEvents(fmt.Sprintf(` where e.start <= %d and e.end > %d and e.status = 0 group by e.id, ex.league, elv.local_vod_name`, start, end))

    if err != nil {
        return nil, err
    }

    return &DueEvents{ stopEv, startEv }, nil
}

func (sch *Scheduler) queryEvents(query string) ([]event.SchEvent, error) {
    rows, err := sch.conn.QueryContext(context.Background(), selectQry(sch.league) + query)

    if err != nil {
        return nil, err
    }

    var result []event.SchEvent

    for rows.Next() {
        var ev event.SchEvent
        if err := ev.Scan(rows); err != nil {
            return nil, err
        }
        result = append(result, ev)
    }
    return result, nil
}

// retuns common select query
func selectQry(league string) string {
    return `SELECT "` + league + `" as league,  e.id, e.start, e.end, e.location_id,
    ifnull(e.manual_schedule, 0) manual_schedule, ifnull(e.schedule_signal, "") schedule_signal,
    e.overlay_visible, e.sport, e.dir,
        elv.local_vod_name,
        t1.name team1,
        t2.name team2,
        ex.league AS target_league, ex.target_id,
        d.flood,
        sport.nevco_code,
        l.location, l.stream, l.address, l.port, l.redis_port, l.inmediate or e.inmediate AS loc_copy,
        global_id,
        ops.copy_method,
        use_rclone
    FROM .event e INNER JOIN location l ON e.location_id = l.id
        INNER JOIN team t1 ON t1.id = e.team_id1
        INNER JOIN team t2 ON t2.id = e.team_id2
        INNER JOIN sport sport ON sport.name = e.sport
        INNER JOIN division d ON d.id = e.division_id
        LEFT JOIN ops.location ops ON ops.id = l.global_id
        LEFT JOIN event_export ex ON ex.event_id = e.id
        LEFT JOIN event_local_vod elv ON elv.event_id = e.id`
}
