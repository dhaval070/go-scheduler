package scheduler
import (
    "log"
    "gsch/model/event"
    "gsch/db"
    "strconv"
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

func (sch *Scheduler) broadcast(ev *event.SchEvent, camera string, alt_stream string) {
    log.Printf("#%d %s broadcast", ev.Id, camera)

}

func (sch *Scheduler) record(ev *event.SchEvent, camera string) {
    log.Printf("#%d %s record", ev.Id, camera)

}
