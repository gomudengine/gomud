package gametime

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/util"
)

const roundDateCacheMax = 20

var (
	dayResetOffset int = 0

	roundDateCache    = map[uint64]GameDate{}
	roundDateCacheSeq []uint64
)

type RoundTimer struct {
	RoundStart uint64 `yaml:"roundstart,omitempty"`
	Period     string `yaml:"period,omitempty"`
	gd         GameDate
}

func (r *RoundTimer) Expired() bool {
	if r.Period == `` || r.RoundStart == 0 {
		return true
	}
	if r.gd.RoundNumber == 0 {
		r.gd = GetDate(r.RoundStart)
	}
	return r.gd.AddPeriod(r.Period) < util.GetRoundCount()
}

type GameDate struct {
	// which calendar this gamedate uses
	Calendar string
	// The round number this GameDate represents
	RoundNumber      uint64
	RoundsPerDay     int
	NightHoursPerDay int

	Year        int
	Month       int
	Week        int
	Day         int
	Hour        int
	Hour24      int
	Minute      int
	MinuteFloat float64
	AmPm        string
	Night       bool

	DayStart   int
	NightStart int
	DuskHours  int
	SunCount   int
	MoonCount  int
}

func (gd GameDate) String(symbolOnly ...bool) string {

	dayNight := `day`
	if gd.Night {
		dayNight = `night`
	} else {
		if gd.NightStart-gd.Hour24 < gd.DuskHours {
			dayNight = `day-dusk`
		}
	}

	if len(symbolOnly) > 0 && symbolOnly[0] {

		if gd.Night {
			return `<ansi fg="night">☾</ansi>` // •
		}
		return fmt.Sprintf(`<ansi fg="%s">☀️</ansi>`, dayNight) //
	}

	return fmt.Sprintf("<ansi fg=\"%s\">%d:%02d%s</ansi>", dayNight, gd.Hour, gd.Minute, gd.AmPm)
}

// Jumps the clock foward to the next night
// If a roundAdjustment is provided, it will be added to the offset
// This is useful to set to the round right before the rollover
func SetToNight(roundAdjustment ...int) {

	dayRound := GetLastPeriod(`sunset`, util.GetRoundCount())

	if len(roundAdjustment) > 0 {
		if roundAdjustment[0] < 0 {
			dayRound -= uint64(-1 * roundAdjustment[0])
		} else {
			dayRound += uint64(roundAdjustment[0])
		}
	}

	gd := GetDate(dayRound).Add(0, 1, 0)
	util.SetRoundCount(gd.RoundNumber)
}

// Jumps the clock forward to the next day
// If a roundAdjustment is provided, it will be added to the offset
// This is useful to set to the round right before the rollover
func SetToDay(roundAdjustment ...int) {

	dayRound := GetLastPeriod(`sunrise`, util.GetRoundCount())

	if len(roundAdjustment) > 0 {
		if roundAdjustment[0] < 0 {
			dayRound -= uint64(-1 * roundAdjustment[0])
		} else {
			dayRound += uint64(roundAdjustment[0])
		}
	}

	gd := GetDate(dayRound).Add(0, 1, 0)
	util.SetRoundCount(gd.RoundNumber)
}

// Jumps the clock forward a specific hour/minutes
// Between 0 and 23
func SetTime(setToHour int, setToMinutes ...int) {

	rpd := activeCalendar[`default`].roundsPerDay
	rph := activeCalendar[`default`].roundsPerHour

	setToHour = setToHour % 24
	dayResetOffset = int(math.Floor(float64(setToHour) * rph))
	if len(setToMinutes) > 0 {
		dayResetOffset += int(math.Ceil((float64(setToMinutes[0]) / 60) * rph))
	}

	roundOfDay := int(util.GetRoundCount() % rpd)
	dayResetOffset -= roundOfDay

	// Reset the cache
	clear(roundDateCache)
	roundDateCacheSeq = roundDateCacheSeq[:0]
}

func IsNight() bool {
	gd := GetDate()
	return gd.Night
}

// Gets the details of the current date
func GetDate(forceRound ...uint64) GameDate {

	currentRound := uint64(0)
	if len(forceRound) > 0 {
		currentRound = forceRound[0]
	} else {
		currentRound = util.GetRoundCount()
	}

	if d, ok := roundDateCache[currentRound]; ok {
		return d
	}

	if len(roundDateCache) >= roundDateCacheMax {
		delete(roundDateCache, roundDateCacheSeq[0])
		roundDateCacheSeq = roundDateCacheSeq[1:]
	}

	roundDateCache[currentRound] = getDate(currentRound)
	roundDateCacheSeq = append(roundDateCacheSeq, currentRound)

	return roundDateCache[currentRound]
}

func getDate(currentRound uint64) GameDate {

	calendarToUse := "default"

	gd := GameDate{Calendar: calendarToUse}

	ac := activeCalendar[calendarToUse]

	gd.RoundNumber = currentRound
	gd.RoundsPerDay = int(ac.roundsPerDay)
	gd.NightHoursPerDay = ac.nightHours
	gd.DuskHours = ac.duskHours
	gd.SunCount = ac.sunCount
	gd.MoonCount = ac.moonCount

	gd.ReCalculate()

	return gd
}

func (g *GameDate) ReCalculate() {

	ac := activeCalendar[g.Calendar]

	currentRoundAdjusted := (g.RoundNumber + uint64(dayResetOffset))
	roundOfDay := int(currentRoundAdjusted % uint64(g.RoundsPerDay))

	hourFloat, minutesFloat := math.Modf(float64(roundOfDay) / float64(g.RoundsPerDay) * 24)

	hour := int(hourFloat)
	hour24 := hour

	night := false
	halfNight := int(math.Floor(float64(g.NightHoursPerDay) / 2))
	nightStart := 24 - halfNight
	nightEnd := int(g.NightHoursPerDay) - halfNight
	if hour >= nightStart || hour < nightEnd {
		night = true
	}

	ampm := `AM`
	if hour >= 12 {
		ampm = `PM`
		hour -= 12
	}

	if hour == 0 {
		hour = 12
	}

	minute := math.Floor(minutesFloat * 60)

	daysPerYear := float64(ac.daysPerYear)
	numMonths := ac.numMonths

	day := math.Floor(float64(currentRoundAdjusted)/float64(g.RoundsPerDay)) + 1
	year := math.Ceil(day / daysPerYear)

	if year > 1 {
		day -= math.Floor((year - 1) * daysPerYear)
	}
	week := math.Floor(float64(day) / float64(ac.daysPerWeek))

	month := 1 + math.Floor((day*24)/ac.hoursPerMonth)
	if int(month) > numMonths {
		month = float64(numMonths)
	}

	g.Day = int(day)
	g.Year = int(year)
	g.Month = int(month)
	g.Week = int(week)
	g.Hour = hour
	g.Hour24 = hour24
	g.Minute = int(minute)
	g.MinuteFloat = minutesFloat * 60
	g.AmPm = ampm
	g.Night = night

	g.NightStart = nightStart
	g.DayStart = nightEnd
}

func (g GameDate) Add(adjustHours int, adjustDays int, adjustYears int) GameDate {

	ac := activeCalendar[g.Calendar]
	rStart := g.RoundNumber

	if adjustYears != 0 {
		yearRounds := uint64(ac.daysPerYear) * uint64(g.RoundsPerDay)
		if adjustYears < 1 {
			g.RoundNumber -= uint64(-1*adjustYears) * yearRounds
		} else {
			g.RoundNumber += uint64(adjustYears) * yearRounds
		}
	}

	if adjustDays != 0 {
		if adjustDays < 1 {
			g.RoundNumber -= uint64(-1 * adjustDays * g.RoundsPerDay)
		} else {
			g.RoundNumber += uint64(adjustDays * g.RoundsPerDay)
		}
	}

	if adjustHours != 0 {
		rph := ac.roundsPerHour
		if adjustHours < 1 {
			g.RoundNumber -= uint64(math.Floor(-1 * float64(adjustHours) * rph))
		} else {
			g.RoundNumber += uint64(math.Floor(float64(adjustHours) * rph))
		}
	}

	if rStart != g.RoundNumber {
		g.ReCalculate()
	}

	return g
}

// Example:
// gd := gametime.GetDate()
// nextPeriodRound := gd.AddPeriod(`10 days`)
// Accepts: x years, x months, x weeks, x days, x hours, x rounds
// If `IRL` or `real` are in the mix, such as `x irl days` or `x days irl`, then it will use real world time
func (g GameDate) AddPeriod(periodStr string) uint64 {

	if periodStr == `` {
		return g.RoundNumber
	}

	qty := 1
	timeStr := ``
	realTime := false
	roundsPerRealDay := 0
	roundsPerRealHour := 0
	roundsPerRealMinute := 0

	parts := strings.Split(strings.ToLower(periodStr), ` `)
	if len(parts) == 1 { // e.g. 2

		// try and parse a number, if not a number, must be a str
		if qty, _ = strconv.Atoi(parts[0]); qty < 1 {
			qty = 1
			timeStr = parts[0]
		}

	} else if len(parts) == 2 { // e.g. - 2 days
		// first arg is quantity, second is unit
		if qty, _ = strconv.Atoi(parts[0]); qty < 1 {
			qty = 1
		}
		timeStr = parts[1]

	} else if len(parts) == 3 {

		// first arg is quantity, second should be `real` and the last is the unit
		if qty, _ = strconv.Atoi(parts[0]); qty < 1 {
			qty = 1
		}

		// RoundSeconds is a real-time value — still read from the timing config (not the calendar).
		c := configs.GetTimingConfig()

		if parts[1] == `real` || parts[1] == `irl` { // e.g. - 2 irl days
			realTime = true
			roundsPerRealDay = 84600 / int(c.RoundSeconds)
			roundsPerRealHour = 3600 / int(c.RoundSeconds)
			roundsPerRealMinute = 60 / int(c.RoundSeconds)

			timeStr = parts[2]
		} else if parts[1] == `game` || parts[1] == `gametime` { // e.g. - 2 game days
			timeStr = parts[2]
		} else if parts[2] == `real` || parts[2] == `irl` { // e.g. - 2 days irl
			realTime = true
			roundsPerRealDay = 84600 / int(c.RoundSeconds)
			roundsPerRealHour = 3600 / int(c.RoundSeconds)
			roundsPerRealMinute = 60 / int(c.RoundSeconds)

			timeStr = parts[1]
		} else if parts[2] == `game` || parts[2] == `gametime` { // e.g. - 2 days gametime
			timeStr = parts[1]
		}

	}

	if len(timeStr) >= 3 {

		strShort := timeStr[0:3]

		if strShort == `yea` { // timeStr == `year` || timeStr == `years` || timeStr == `yearly` {

			if realTime {
				adjustment := uint64(qty * roundsPerRealDay * 365)
				return g.RoundNumber + adjustment
			}

			gNext := g.Add(0, 0, 1*qty)

			return gNext.RoundNumber

		} else if strShort == `mon` { // else if timeStr == `month` || timeStr == `months` || timeStr == `monthly` {

			if realTime {
				adjustment := uint64(qty * roundsPerRealHour * 730)
				return g.RoundNumber + adjustment
			}

			hoursPerMonth := activeCalendar[g.Calendar].hoursPerMonth
			gNext := g.Add(int(math.Round(hoursPerMonth))*qty, 0, 0)

			return gNext.RoundNumber

		} else if strShort == `wee` { //  else if timeStr == `week` || timeStr == `weeks` || timeStr == `weekly` {

			if realTime {
				adjustment := uint64(qty * roundsPerRealDay * 7)
				return g.RoundNumber + adjustment
			}

			gNext := g.Add(0, activeCalendar[g.Calendar].daysPerWeek*qty, 0)

			return gNext.RoundNumber

		} else if strShort == `day` || strShort == `dai` { //  else if timeStr == `day` || timeStr == `days` || timeStr == `daily` {

			if realTime {
				adjustment := uint64(qty * roundsPerRealDay)
				return g.RoundNumber + adjustment
			}

			gNext := g.Add(0, qty, 0)

			return gNext.RoundNumber

		} else if strShort == `hou` { // if timeStr == `hour` || timeStr == `hours` || timeStr == `hourly` {

			if realTime {
				adjustment := uint64(qty * roundsPerRealHour)
				return g.RoundNumber + adjustment
			}

			gNext := g.Add(qty, 0, 0)

			return gNext.RoundNumber

		} else if strShort == `min` { // if timeStr == `minute` || if timeStr == `minutes` || if timeStr == `minutely`

			if realTime {
				adjustment := uint64(qty * roundsPerRealMinute)
				return g.RoundNumber + adjustment
			}

			return g.RoundNumber + uint64(math.Floor(float64(qty)*activeCalendar[g.Calendar].roundsPerMinute))

		} else if strShort == `noo` { // if timeStr == `noon` || timeStr == `noons` {

			if realTime {
				mudlog.Error("AddPeriod", "error", "real time not supported for noon yet: "+timeStr)
			}

			g = getDate(GetLastPeriod(`noon`, g.RoundNumber))
			// adjusts by days
			gNext := g.Add(0, qty, 0)

			return gNext.RoundNumber

		} else if strShort == `mid` { // if timeStr == `midnight` || timeStr == `midnights` {

			if realTime {
				mudlog.Error("AddPeriod", "error", "real time not supported for midnight yet: "+timeStr)
			}

			g = getDate(GetLastPeriod(`day`, g.RoundNumber))
			// adjusts by days
			gNext := g.Add(0, qty, 0)

			return gNext.RoundNumber

		} else if timeStr == `sunrise` || timeStr == `sunrises` {

			if realTime {
				mudlog.Error("AddPeriod", "error", "real time not supported for sunrise yet: "+timeStr)
			}

			g = getDate(GetLastPeriod(`sunrise`, g.RoundNumber))
			// adjusts by days
			gNext := g.Add(0, qty, 0)

			return gNext.RoundNumber

		} else if timeStr == `sunset` || timeStr == `sunsets` {

			if realTime {
				mudlog.Error("AddPeriod", "error", "real time not supported for sunset yet: "+timeStr)
			}

			g = getDate(GetLastPeriod(`sunset`, g.RoundNumber))
			// adjusts by days
			gNext := g.Add(0, qty, 0)

			return gNext.RoundNumber

		}

		// Failover to rounds
		return g.RoundNumber + uint64(qty)
	}

	// Assume rounds?
	//if timeStr == `hour` || timeStr == `hours` || timeStr == `hourly` {

	gNext := g.Add(qty, 0, 0)

	return gNext.RoundNumber

	//}

}

func GetLastPeriod(periodName string, roundNumber uint64) uint64 {

	ac := activeCalendar[`default`]

	roundsPerDay := ac.roundsPerDay
	nightHoursPerDay := uint64(ac.nightHours)
	roundsPerHour := ac.roundsPerHour
	noonRound := ac.noonRound

	// What round started this week?
	roundOfWeek := roundNumber % ac.roundsPerWeek

	// What round started this day? (midnight)
	roundOfDay := roundNumber % roundsPerDay

	// What round started this hour?
	roundOfHour := roundOfDay % uint64(math.Floor(roundsPerHour))

	if periodName == `hour` { // Start of the current hour (or closest to it)

		roundNumber -= roundOfHour

	} else if periodName == `day` { // Start of current day

		roundNumber -= roundOfDay

	} else if periodName == `week` { // Start of current week

		roundNumber -= roundOfWeek // First go to the start of the day

	} else if periodName == `noon` { // Last time 12pm was hit

		roundNumber -= roundOfDay
		if roundOfDay < noonRound {
			// We haven't reached noon today yet; last noon was yesterday.
			roundNumber -= roundsPerDay - noonRound
		} else {
			// Noon has already passed today.
			roundNumber += noonRound
		}

	} else if periodName == `sunrise` { // last sunrise

		roundNumber -= roundOfDay                                                       // Strip rounds of today off
		roundNumber -= roundsPerDay                                                     // Subtract a day
		roundNumber += uint64(math.Ceil(float64(nightHoursPerDay) / 2 * roundsPerHour)) // add half a night

	} else if periodName == `sunset` { // 12am of next day, minus half of night

		roundNumber -= roundOfDay                                                       // Strip rounds of today off
		roundNumber -= uint64(math.Ceil(float64(nightHoursPerDay) / 2 * roundsPerHour)) // Subtract half a night

	}

	return roundNumber
}

func MonthName(month int) string {
	names := activeCalendar[`default`].monthNames
	if len(names) == 0 {
		return ``
	}
	month--
	return names[month%len(names)]
}

func GetZodiac(year int) string {
	z := activeCalendar[`default`].zodiacList
	if len(z) == 0 {
		return ``
	}
	return z[year%len(z)]
}
