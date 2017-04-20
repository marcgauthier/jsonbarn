package models

import (
	"encoding/json"
	"time"

	"github.com/antigloss/go/logger"
	"github.com/robarchibald/calendar"
)

/*Recurrentdate is a type of data structure require for each JSON object that need to have recurrence.

The Recurrence struct is modeled after the recurring schedule data model used by both Microsoft Outlook and Google Calendar for recurring appointments. Just like Outlook, you can pick from Daily ("D"), Weekly ("W"), Monthly ("M") and Yearly ("Y") recurrence pattern codes. Each of those recurrence patterns then require the corresponding information to be filled in.

All recurrences:

StartDateTime - start time of the appointment. Should be set to the first desired occurence of the recurring appointment
RecurrencePatternCode - D: daily, W: weekly, M: monthly or Y: yearly
RecurEvery - number defining how many days, weeks, months or years to wait between recurrences
EndByDate (optional) - date by which recurrences must be done by
NumberOfOccurrences (optional) - data for UI which can be used to store the number of recurrences. Has no effect in calculations though. EndByDate must be calculated based on NumberOfOccurrences
Recurrence Pattern Code D (daily)

DailyIsOnlyWeekday (optional) - ensure that daily occurrences only fall on weekdays (M, T, W, Th, F)
Recurrence Pattern Code W (weekly)

WeeklyDaysIncluded - binary value (converted to int16) to indicate days included (e.g. 0101010 or decimal 42 would be MWF). Each of the individual days are bitwise AND'd together to get the value.
Sunday - 64 (1000000)
Monday - 32 (0100000)
Tuesday - 16 (0010000)
Wednesday - 8 (0001000)
Thursday - 4 (0000100)
Friday - 2 (0000010)
Saturday - 1 (0000001)
Recurrence Pattern Code M (monthly)

MonthlyWeekOfMonth - week of the month to recur on. e.g. Thanksgiving is always on the 4th week of the month. Must be used together with MonthlyDayOfWeek
MonthlyDayOfWeek - day of the week to recur on (0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday). Must be used together with MonthlyWeekOfMonth OR
MonthlyDay - day of the month to recur on. e.g. 5 would recur on the 5th of every month
Recurrence Pattern Code Y (yearly)

YearlyMonth - month of the year to recur on (1=January, 2=February, 3=March, 4=April, 5=May, 6=June, 7=July)
MonthlyWeekOfMonth - week of the month to recur on. e.g. Thanksgiving is always on the 4th week of the month. Must be used together with MonthlyDayOfWeek
MonthlyDayOfWeek - day of the week to recur on (0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday). Must be used together with MonthlyWeekOfMonth OR
MonthlyDay - day of the month to recur on. e.g. 5 would recur on the 5th of every month

*/
type Recurrentdate struct {
	StartDate             uint64  `json:"startdate"`             // Date to start Recurrence. Note that time and time zone information is NOT used in calculations
	Duration              int     `json:"duration"`              // in seconds
	RecurrencePatternCode string  `json:"recurrencepatterncode"` // D for daily, W for weekly, M for monthly or Y for yearly
	RecurEvery            int16   `json:"recurevery"`            // number of days, weeks, months or years between occurrences
	YearlyMonth           *int16  `json:"yearlymonth"`           // month of the year to recur (applies only to RecurrencePatternCode: Y)
	MonthlyWeekOfMonth    *int16  `json:"monthlyweekofmonth"`    // week of the month to recur. used together with MonthlyDayOfWeek (applies only to RecurrencePatternCode: M or Y)
	MonthlyDayOfWeek      *int16  `json:"monthlydayofweek"`      // day of the week to recur. used together with MonthlyWeekOfMonth (applies only to RecurrencePatternCode: M or Y)
	MonthlyDay            *int16  `json:"monthlyday"`            // day of the month to recur (applies only to RecurrencePatternCode: M or Y)
	WeeklyDaysIncluded    *int16  `json:"weeklydaysincluded"`    // integer representing binary values AND'd together for 1000000-64 (Sun), 0100000-32 (Mon), 0010000-16 (Tu), 0001000-8 (W), 0000100-4 (Th), 0000010-2 (F), 0000001-1 (Sat). (applies only to RecurrencePatternCode: M or Y)
	DailyIsOnlyWeekday    *bool   `json:"dailyisonlyweekday"`    // indicator that daily recurrences should only be on weekdays (applies only to RecurrencePatternCode: D)
	EndByDate             *uint64 `json:"endbydate"`             // date by which all occurrences must end by. Note that time and time zone information is NOT used in calculations
}

/*GetNextDatePeriod return time next start
  get the date a CFSR will start on restart for recurrent.

  	// test time

	rr := `{"startdate":` + strconv.FormatUint(uint64(models.UnixUTCnano()), 10) + `, "duration":60, "recurrencepatterncode":"D", "recurevery":1 }`
	start, end := models.GetNextDatePeriod(rr)

	fmt.Println("next recurent start" + time.Unix(0, int64(start)).String())
	fmt.Println("next recurent end  " + time.Unix(0, int64(end)).String())
	panic("end")

	logger.Trace("Openning Database")


*/
func GetNextDatePeriod(RecurrentInfo string) (uint64, uint64) {

	r := Recurrentdate{}
	err := json.Unmarshal([]byte(RecurrentInfo), &r)
	if err != nil {
		logger.Error("GetNextDatePeridod:" + err.Error())
		return 0, 0
	}

	// time must be pass in seconds

	startDate := time.Unix(int64(r.StartDate), 0).UTC()

	var endDate *time.Time
	var t time.Time

	if r.EndByDate == nil {
		endDate = nil
	} else {
		t = time.Unix(int64(*r.EndByDate), 0)
		endDate = &t
	}

	recurrences := calendar.Recurrence{
		StartDate:             startDate,
		RecurrencePatternCode: r.RecurrencePatternCode,
		RecurEvery:            r.RecurEvery,
		YearlyMonth:           r.YearlyMonth,
		MonthlyWeekOfMonth:    r.MonthlyWeekOfMonth,
		MonthlyDayOfWeek:      r.MonthlyDayOfWeek,
		MonthlyDay:            r.MonthlyDay,
		WeeklyDaysIncluded:    r.WeeklyDaysIncluded,
		DailyIsOnlyWeekday:    r.DailyIsOnlyWeekday,
		EndByDate:             endDate,
	}

	if endDate == nil {
		t := time.Now().Add(time.Duration(30*24*60*60) * time.Second) // add 30 days
		endDate = &t
	}

	occurrences := recurrences.GetOccurrences(startDate, *endDate)

	now := int64(UnixUTCSecs())

	for i := 0; i < len(occurrences); i++ {

		if i < 10 {
			logger.Trace("now: " + time.Unix(now, 0).String() + " occurence: " + occurrences[i].String())
		}

		if occurrences[i].UTC().Unix() > now {

			/* occurence don't contain time... */

			t := time.Date(occurrences[i].Year(), occurrences[i].Month(), occurrences[i].Day(), startDate.Hour(), startDate.Minute(), startDate.Second(), 0, time.UTC)
			start := t.UTC().Unix()
			end := t.Add(time.Duration(r.Duration) * time.Second).UTC().Unix()

			logger.Trace(" ")
			logger.Trace("Recurent request " + RecurrentInfo)
			logger.Trace("New StartDate = " + t.String())
			logger.Trace("New End  Date = " + t.Add(time.Duration(r.Duration)*time.Second).UTC().String())
			logger.Trace(" ")

			return uint64(start), uint64(end)
		}

	}

	return 0, 0

}
