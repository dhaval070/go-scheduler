package scheduler
import (
    "log"
    "gsch/model/event"
    "gsch/db"
    "strconv"
    "database/sql"
)

func (sch *Scheduler) lockLocation(eventId int, globalId int) bool {
    loc := event.FindOpsLocation(globalId)

    if loc.Status > 0 {
        if loc.Locked_by_league.String != sch.league {
            log.Println("location locked by league " + loc.Locked_by_league.String + ":" +
            strconv.Itoa(int(loc.Locked_by_event_id.Int32)))
            return false
        }
    }

    db.Exec("update ops.location set status = 1, locked_by_league=?, locked_by_event_id=? where id=?", sch.league, eventId, globalId)

    return true
}

func (sch *Scheduler) broadcast(ev *event.SchEvent, camera string, alt_stream string) bool {
    log.Printf("#%d %s broadcast", ev.Id, camera)

    return true
}

func (sch *Scheduler) record(ev *event.SchEvent, camera string) bool {
    log.Printf("#%d %s record", ev.Id, camera)

    return true
}

type OverlayConf struct {
    NetworkLogo sql.NullString
    OverlayFile sql.NullString
    HomeTeam sql.NullString
    GuestTeam sql.NullString
    HomeTeamLogo sql.NullString
    GuestTeamLogo sql.NullString
    LeagueLogo sql.NullString
    Template sql.NullString
    MaxPeriods sql.NullString
    LastPeriodDuration sql.NullString
}

func (sch *Scheduler) setOverlayConf(ev *event.SchEvent) {
    if ev.RedisPort == "" {
        return
    }

    rows, done := db.Query(`select
        network_logo, overlay_graphic_file, home_team, guest_team, home_team_logo, guest_team_logo,
        league_logo, template, max_periods, last_period_duration FROM `+sch.dbname+`.overlay_conf`)

    defer done()

    if !rows.Next() {
        return
    }
    var ov OverlayConf

    db.ScanRows(rows, &ov.NetworkLogo, &ov.OverlayFile, &ov.HomeTeam, &ov.GuestTeam, &ov.HomeTeamLogo,
        &ov.GuestTeamLogo, &ov.LeagueLogo, &ov.Template, &ov.MaxPeriods, &ov.LastPeriodDuration)

    if ev.Team1LogoFile != "" {
        ov.HomeTeamLogo = sql.NullString{ev.Team1LogoFile, true }
    }

    if ev.Team1ShortName != "" {
        ov.HomeTeam = sql.NullString{ ev.Team1ShortName, true }
    }
    if ev.Team2LogoFile != "" {
        ov.GuestTeamLogo = sql.NullString{ev.Team2LogoFile, true }
    }

    if ev.Team2ShortName != "" {
        ov.GuestTeam = sql.NullString{ ev.Team2ShortName, true }
    }
}
