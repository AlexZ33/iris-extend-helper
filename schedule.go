package iris_extend_helper

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/robfig/cron/v3"
)

type Job struct {
	Schedule string
	Task     func()
}

func RegisterScheduledTasks(app *iris.Application, jobs []Job) {
	c := cron.New(cron.WithSeconds())
	for _, job := range jobs {
		schedule := job.Schedule
		task := job.Task
		if strings.Contains(schedule, " ") {
			if _, ok := ParseSchedule(schedule); ok {
				c.AddFunc(schedule, task)
			}
		} else {
			duration, err := time.ParseDuration(schedule)
			if err != nil {
				log.Println(err)
			} else {
				seconds := int(duration.Seconds())
				if seconds == 1 {
					c.AddFunc("* * * * * *", task)
				} else if seconds >= 2 && seconds < 60 {
					for i := 0; i < seconds; i++ {
						expr := fmt.Sprintf("%d/%d * * * * *", i, seconds)
						c.AddFunc(expr, task)
					}
				} else {
					minutes := int(duration.Minutes())
					if minutes == 1 {
						for i := 0; i < 60; i++ {
							expr := fmt.Sprintf("%d * * * * *", i)
							c.AddFunc(expr, task)
						}
					} else if minutes >= 2 && minutes < 60 {
						for i := 0; i < 60; i++ {
							for j := 0; j < minutes; j++ {
								expr := fmt.Sprintf("%d %d/%d * * * *", i, j, minutes)
								c.AddFunc(expr, task)
							}
						}
					} else {
						hours := int(duration.Hours())
						if hours == 1 {
							for i := 0; i < 60; i++ {
								for j := 0; j < 60; j++ {
									expr := fmt.Sprintf("%d %d * * * *", i, j)
									c.AddFunc(expr, task)
								}
							}
						} else if hours >= 2 && hours < 60 {
							for i := 0; i < 60; i++ {
								for j := 0; j < 60; j++ {
									for k := 0; k < hours; k++ {
										expr := fmt.Sprintf("%d %d %d/%d * * *", i, j, k, hours)
										c.AddFunc(expr, task)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	c.Start()
}

func ParseSchedule(expr string) (cron.Schedule, bool) {
	scheduler := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if schedule, err := scheduler.Parse(expr); err != nil {
		log.Println(err)
	} else {
		return schedule, true
	}
	return &cron.SpecSchedule{}, false
}
