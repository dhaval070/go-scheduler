package event
import (
    "fmt"
    "database/sql"
    "gsch/db"
    "context"
    "time"
    "log"
)

type SchEvent struct {
    League string
    Id int
    Start int
    End int
    LocationId int
    ManualSchedule sql.NullInt32
    ScheduleSignal sql.NullString
    OverlayVisible sql.NullInt32
    TargetId sql.NullInt32
    TargetLeague sql.NullString
    Sport sql.NullString
    Dir string
    LocalVodName sql.NullString
    Team1 string
    Team2 string
    Flood sql.NullInt32
    NevcoCode sql.NullString
    Location string
    Stream sql.NullString
    Address sql.NullString
    Port sql.NullString
    RedisPort sql.NullString
    LocCopy sql.NullInt32
    GlobalId int
    CopyMethod string
    UseRclone int
}

func (ev *SchEvent) Scan(r *sql.Rows) error {
    err := r.Scan(&ev.League, &ev.Id, &ev.Start, &ev.End, &ev.LocationId, &ev.ManualSchedule, &ev.ScheduleSignal,
        &ev.OverlayVisible, &ev.Sport, &ev.Dir, &ev.LocalVodName,
        &ev.Team1, &ev.Team2, &ev.TargetLeague, &ev.TargetId,
        &ev.Flood, &ev.NevcoCode, &ev.Location, &ev.Stream, &ev.Address,
        &ev.Port, &ev.RedisPort, &ev.LocCopy, &ev.GlobalId, &ev.CopyMethod, &ev.UseRclone)

    return err
}

type DueEvents struct {
    stopDue []SchEvent
    startDue []SchEvent
}

type Scheduler struct {
    league string
    conn *sql.Conn
}

func NewScheduler(league string) *Scheduler {
    var conn *sql.Conn
    var ctx = context.Background()
    var err error
    conn, err = db.Db().Conn(ctx)

    if err != nil {
        log.Fatal(err)
        return nil
    }

    if _, err := conn.ExecContext(ctx, "use gos_" + league); err != nil {
        log.Fatal(err)
    }

    return &Scheduler { league, conn }
}

func (sch *Scheduler) Destroy() {
    sch.conn.Close()
}

func (sch *Scheduler) Work() {
    log.Println("processing " + sch.league)
    var events *DueEvents
    var err error

    if dueEvents, err = sch.GetEvents(); err == nil {
        sch.stopEvents(dueEvents.stopDue)
        return
    }
    log.Fatal(err)
}

func (sch *Scheduler) StopEvents([]SchEvent) {


}

// retuns stop due and start due events
func (sch *Scheduler) GetEvents() (*DueEvents, error) {
    var err error

    var stopEv, startEv []SchEvent
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

func (sch *Scheduler) queryEvents(query string) ([]SchEvent, error) {
    rows, err := sch.conn.QueryContext(context.Background(), selectQry(sch.league) + query)

    if err != nil {
        return nil, err
    }

    var result []SchEvent

    for rows.Next() {
        var ev SchEvent
        if err := ev.Scan(rows); err != nil {
            return nil, err
        }
        result = append(result, ev)
    }
    return result, nil
}

// retuns common select query
func selectQry(league string) string {
    return `SELECT "` + league + `" as league,  e.id, e.start, e.end, e.location_id, e.manual_schedule, e.schedule_signal, e.overlay_visible,
        e.sport, e.dir,
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
