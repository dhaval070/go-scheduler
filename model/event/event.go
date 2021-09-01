package event
import (
    "gsch/db"
    "database/sql"
)

type SchEvent struct {
    League string
    Id int
    Start int
    End int
    LocationId int
    ManualSchedule bool
    ScheduleSignal string
    OverlayVisible bool
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

func (ev *SchEvent) Scan(r *sql.Rows) {
    err := r.Scan(&ev.League, &ev.Id, &ev.Start, &ev.End, &ev.LocationId, &ev.ManualSchedule, &ev.ScheduleSignal,
        &ev.OverlayVisible, &ev.Sport, &ev.Dir, &ev.LocalVodName,
        &ev.Team1, &ev.Team2, &ev.TargetLeague, &ev.TargetId,
        &ev.Flood, &ev.NevcoCode, &ev.Location, &ev.Stream, &ev.Address,
        &ev.Port, &ev.RedisPort, &ev.LocCopy, &ev.GlobalId, &ev.CopyMethod, &ev.UseRclone)

    if err != nil {
        panic(err)
    }
}

type OpsLocation struct {
    Id int
    Status int
    Locked_by_league sql.NullString
    Locked_by_event_id sql.NullInt32
}

func FindOpsLocation(id int) *OpsLocation {
    row := db.QueryRow(
        "select id, status, locked_by_league, locked_by_event_id from ops.location where id = ?",
        id)

    var loc OpsLocation

    db.ScanRow(row, &loc.Id, &loc.Status, &loc.Locked_by_league, &loc.Locked_by_event_id)
    return &loc
}

type Cmd struct {
    Camera string
    StartBcast sql.NullString
    StopBcast sql.NullString
    StartRecord sql.NullString
    StopRecord sql.NullString
}

func FindCmd(dbname string, locId int) []Cmd {
    rows, done := db.Query(`select
        camera,
        start_bcast_cmd,
        stop_bcast_cmd,
        start_local_record_cmd,
        stop_local_record_cmd
        from ` + dbname + `.camera_location where location_id=?`, locId)

    defer done()

    var result []Cmd

    for rows.Next() {
        var cmd Cmd

        db.ScanRows(rows, &cmd.Camera, &cmd.StartBcast, &cmd.StopBcast, &cmd.StartRecord, &cmd.StopRecord)

        result = append(result, cmd)
    }
    return result
}

type EventStream struct {
    EventId int
    Camera string
    Broadcast bool
    Record bool
    RecordAt string
    AltStream string
}

func GetEventStreams(dbname string, eventId int) ([]EventStream){
    res, done := db.Query(`select
        event_id,
        camera,
        broadcast,
        record,
        record_at,
        alt_stream
    from ` + dbname + `.event_stream where event_id=?`, eventId)

    defer done()

    var result []EventStream

    for res.Next() {
        var r EventStream

        db.ScanRows(res, &r.EventId, &r.Camera, &r.Broadcast, &r.Record, &r.RecordAt, &r.AltStream)

        result = append(result, r)
    }
    return result
}

