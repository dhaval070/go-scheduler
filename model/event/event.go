package event
import (
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

func (ev *SchEvent) Scan(r *sql.Rows) error {
    err := r.Scan(&ev.League, &ev.Id, &ev.Start, &ev.End, &ev.LocationId, &ev.ManualSchedule, &ev.ScheduleSignal,
        &ev.OverlayVisible, &ev.Sport, &ev.Dir, &ev.LocalVodName,
        &ev.Team1, &ev.Team2, &ev.TargetLeague, &ev.TargetId,
        &ev.Flood, &ev.NevcoCode, &ev.Location, &ev.Stream, &ev.Address,
        &ev.Port, &ev.RedisPort, &ev.LocCopy, &ev.GlobalId, &ev.CopyMethod, &ev.UseRclone)

    return err
}

