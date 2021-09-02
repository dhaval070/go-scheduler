package scheduler
import (
    "log"
    "gsch/model/event"
    "gsch/db"
    "strconv"
    "reflect"
    "gsch/redis"
    "context"
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

type overlayConf struct {
    network_logo string
    overlay_graphic_file string
    home_team string
    guest_team string
    home_team_logo string
    guest_team_logo string
    league_logo string
    template string
    max_periods string
    last_period_duration string
}

func (sch *Scheduler) setOverlayConf(ev *event.SchEvent) {
    if ev.Address == "" || ev.RedisPort == "" {
        return
    }

    rows, done := db.Query(`select
        ifnull(network_logo, "") network_logo,
        ifnull(overlay_graphic_file, "") overlay_graphic_file,
        ifnull(home_team, "") home_team,
        ifnull(guest_team, "") guest_team,
        ifnull(home_team_logo, "") home_team_logo,
        ifnull(guest_team_logo, "") guest_team_logo,
        ifnull(league_logo, "") league_logo,
        ifnull(template, "") template,
        ifnull(max_periods, "") max_periods,
        ifnull(last_period_duration,"") last_period_duration FROM `+sch.dbname+`.overlay_conf`)

    defer done()

    if !rows.Next() {
        return
    }
    var ov overlayConf

    db.ScanRows(rows, &ov.network_logo, &ov.overlay_graphic_file, &ov.home_team, &ov.guest_team, &ov.home_team_logo,
        &ov.guest_team_logo, &ov.league_logo, &ov.template, &ov.max_periods, &ov.last_period_duration)

    if ev.Team1LogoFile != "" {
        ov.home_team_logo = ev.Team1LogoFile
    }
    if ev.Team1ShortName != "" {
        ov.home_team = ev.Team1ShortName
    }
    if ev.Team2LogoFile != "" {
        ov.guest_team_logo = ev.Team2LogoFile
    }
    if ev.Team2ShortName != "" {
        ov.guest_team = ev.Team2ShortName
    }

    log.Println(ev.Address + ":" + ev.RedisPort)
    var client = rclient.GetClient(rclient.Options{
        Addr: ev.Address + ":" + ev.RedisPort,
    })
    ctx := context.Background()

    for i, field := range reflect.VisibleFields(reflect.TypeOf(ov)) {
        var value = reflect.ValueOf(ov).Field(i)
        //log.Println(string(field.String()))
        client.HSet(ctx, "scoreboard", field.Name, value.String())
    }

    client.Set(ctx, "nevco_code", ev.NevcoCode.String, 0)
    client.HSet(ctx, "scoreboard", "visible", ev.OverlayVisible)
    client.HSet(ctx, "scoreboard", "event_id", ev.Id)
    client.HSet(ctx, "scoreboard", "status", 1)
    client.HSet(ctx, "scoreboard", "league", sch.league)
    client.HSet(ctx, "scoreboard", "flood", ev.Flood.Int32)

    // insert into event_redis with status= "error" or "success"
}
