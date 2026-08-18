package workouts

import "time"

func nextStreak(last *time.Time, current int32, today time.Time) int32 {
	if last == nil {
		return 1
	}
	days := int(today.Sub(*last).Hours() / 24)
	switch days {
	case 0:
		return current
	case 1:
		return current + 1
	default:
		return 1
	}
}
