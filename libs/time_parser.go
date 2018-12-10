package libs

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	Range = iota
	Separate
	EverySeparate
)

var Year = time.Now().Year()

type (
	TaskTime struct {
		Second, Minute *TimeTicket
		Hour, Day, Month, Week *TimeTicket
		NextExecTime time.Time
		SecondInterval int			// 针对 */n 格式的动态计算标志位
		MinuteInterval int
		HourInterval int
		DayInterval int
		MonthInterval int
		minuteStart, hourStart int  // 针对 */n 格式的重置位
		secondStart, dayStart int
		monthStart int
	}

	TimeTicket struct {
		Ticket []int
		Cursor int
	}
)

func TimeParse(express string, now time.Time) (*TaskTime, error) {
	re := regexp.MustCompile("\\s+")
	tsList := re.Split(express, -1)
	if len(tsList) != 6 {
		return nil, errors.New(fmt.Sprintf("illegal express: %s", express))
	}

	t :=  new(TaskTime)

	err := t.ParseSecond(tsList, now)
	if err != nil {
		return nil, err
	}

	err = t.ParseMinute(tsList, now)
	if err != nil {
		return nil, err
	}

	err = t.ParseHour(tsList, now)
	if err != nil {
		return nil, err
	}

	err = t.ParseDay(tsList, now)
	if err != nil {
		return nil, err
	}

	err = t.ParseMonth(tsList, now)
	if err != nil {
		return nil, err
	}

	err = t.ParseWeek(tsList, now)
	if err != nil {
		return nil, err
	}

	t.NextExecTime = now
	t.ComputeNextExecTime()
	return t, nil
}

func parse(timeString, parseType string, minTime, maxTime int, startTime time.Time) ([]int, int, error) {
	var (
		tmpStringSlice = make([]string, 0, 4)
		result []int
		times int
		flag = -1
		zeroInterval int
	)

	result = make([]int, 0, 60)
	L: for _, s := range timeString {
		v := string(s)
		switch v {
		case "-":
			times++

			// 符号出现超过1次则是非法表达式
			if times > 1 {
				return nil, zeroInterval, errors.New("illegal express, parse - ")
			}

			flag = Range
			intString := strings.Join(tmpStringSlice, "")
			// clean  slice
			tmpStringSlice = make([]string, 0, 4)
			tmpStringSlice = append(tmpStringSlice, intString)
			continue
		case ",":
			times++
			flag = Separate
			tmpStringSlice = strings.Split(timeString, ",")
			break L
		case "/":
			times++
			if times > 1 {
				return nil, zeroInterval, errors.New("illegal express, parse /")
			}

			flag = EverySeparate
			if len(tmpStringSlice) != 1 && tmpStringSlice[0] != "*" {
				return nil, zeroInterval, errors.New("illegal express, parse /")
			}
			continue
		}
		tmpStringSlice = append(tmpStringSlice, v)
	}

	// 处理 10-30  3,4,5  */12 等情况
	switch flag {
	case Range:
		start, err := strconv.Atoi(tmpStringSlice[0])
		if err != nil {
			return nil, zeroInterval, err
		}
		stop, err := strconv.Atoi(strings.Join(tmpStringSlice[1:], ""))
		if err != nil {
			return nil, zeroInterval, err
		}

		// validate
		if start < minTime || start > maxTime || stop < minTime || stop > maxTime {
			return nil, zeroInterval, errors.New("illegal express, generate range")
		}

		for i := start; i <= stop; i++ {
			result = append(result, i)
		}
		return result, zeroInterval, nil
	case Separate:
		for _, v := range tmpStringSlice {
			i, err := strconv.Atoi(v)
			if err != nil{
				return nil, zeroInterval, err
			}

			if i < minTime || i > maxTime {
				return nil, zeroInterval, errors.New("illegal express, generate separate")
			}

			result = append(result, i)
		}
		return result, zeroInterval, nil
	case EverySeparate:
		interval, err := strconv.Atoi(strings.Join(tmpStringSlice[1:], ""))
		if err != nil {
			return nil, zeroInterval, err
		}

		// 如果day 是 */3 的形式，则返回interval动态计算
		if parseType == "day" {
			return nil, interval, nil
		}

		if interval < minTime || interval > maxTime {
			return nil, zeroInterval, errors.New("illegal express, generate everySeparate")
		}

		var t int
		switch parseType {
		case "second":
			t = int(startTime.Second())
		case "minute":
			t = int(startTime.Minute())
		case "hour":
			t = int(startTime.Hour())
		case "month":
			t = int(startTime.Month())
		case "week":
			t = int(startTime.Weekday())
		}

		period := maxTime - minTime + 1

		// 非闭环时间，动态计算
		if period % interval != 0 {
			return nil, interval, nil
		}

		// 闭环时间, 静态计算
		length := int(period / interval)
		tmpResult := make(sort.IntSlice, 0, length)
		for i := 0; i < length; i++{
			tmpResult = append(tmpResult, t % period)
			t += interval
		}
		tmpResult.Sort()
		result = []int(tmpResult)
		return result, zeroInterval, nil
	}

	// 处理 * 的情况
	if len(tmpStringSlice) == 1 && tmpStringSlice[0] == "*" {

		// 如果day 是 * 形式，返回interval 动态计算
		if parseType == "day" {
			return nil, 1, nil
		}

		for i := minTime; i <= maxTime; i++ {
			result = append(result, i)
		}
		return result, zeroInterval, nil
	}

	// 处理常数情况
	sc, err := strconv.Atoi(strings.Join(tmpStringSlice, ""))
	if err != nil {
		return nil, zeroInterval, err
	}

	if sc < minTime || sc > maxTime {
		return nil, zeroInterval, errors.New("illegal express, generate constant")
	}
	result = append(result, sc)
	return result, zeroInterval, nil
}

func (t *TaskTime) ParseSecond(tsList []string, now time.Time) error {
	startTime, currentSecond := now, now.Second()
	if tsList[1] != "*" {
		 startTime = time.Date(
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), 0, 0, now.Location(),
		)
		currentSecond = 0
		t.secondStart = 1
	}

	tl, interval, err := parse(tsList[0], "second", 0, 59, startTime)
	if err != nil {
		return err
	}

	if tl != nil {
		t.Second = &TimeTicket{Ticket: tl,  Cursor: getCursor(tl, currentSecond)}
	}

	t.SecondInterval = interval
	return nil
}

func (t *TaskTime) ParseMinute(tsList []string, now time.Time) error  {
	startTime, currentMinute := now, now.Minute()
	if tsList[2] != "*" {
		startTime = time.Date(
			now.Year(), now.Month(), now.Day(),
			now.Hour(), 0, 0, 0, now.Location(),
		)
		currentMinute = 0
		t.minuteStart = 1
	}
	tl, interval, err := parse(tsList[1], "minute", 0, 59, startTime)
	if err != nil {
		return err
	}

	if tl != nil {
		t.Minute = &TimeTicket{Ticket: tl, Cursor: getCursor(tl, currentMinute)}
	}

	t.MinuteInterval = interval
	return nil
}

func (t *TaskTime) ParseHour(tsList []string, now time.Time) error  {
	startTime, currentHour := now, now.Hour()
	if tsList[3] != "*" || tsList[5] != "*" {
		startTime = time.Date(
			now.Year(), now.Month(), now.Day(),
			0, 0, 0, 0, now.Location(),
		)
		currentHour = 0
		t.hourStart = 1
	}
	tl, interval, err := parse(tsList[2], "hour", 0, 23, startTime)
	if err != nil {
		return err
	}

	if tl != nil {
		t.Hour =  &TimeTicket{Ticket: tl, Cursor: getCursor(tl, currentHour)}
	}
	t.HourInterval = interval
	return nil
}

func (t *TaskTime) ParseDay(tsList []string, now time.Time) error {
	month := int(now.Month())
	startTime, currentDay := now, now.Day()
	if tsList[4] != "*" {
		startTime = time.Date(
			now.Year(), now.Month(), 1,
			0, 0, 0, 0, now.Location(),
		)
		currentDay = 1
		t.dayStart = 1
	}
	tl, interval, err := parse(tsList[3], "day", 1, GetMaxDay(month), startTime)
	if err != nil {
		return  err
	}

	if tl != nil {
		t.Day =  &TimeTicket{Ticket: tl, Cursor: getCursor(tl, currentDay)}
	}

	t.DayInterval = interval
	if tsList[5] != "*" {
		t.DayInterval = 0
	}
	return nil
}

func (t *TaskTime) ParseMonth(tsList []string, now time.Time) error  {
	startTime, currentMonth := now, int(now.Month())
	tl, interval, err := parse(tsList[4], "month", 1, 12, startTime)
	if err != nil {
		return err
	}

	if tl != nil {
		t.Month =  &TimeTicket{Ticket: tl, Cursor: getCursor(tl, currentMonth)}
	}

	t.MonthInterval = interval
	return nil
}

func (t *TaskTime) ParseWeek(tsList []string, now time.Time) error  {
	startTime, currentWeek := now, int(now.Weekday())
	if tsList[4] != "*" {
		startTime = time.Date(
			now.Year(), now.Month(), 1,
			0, 0, 0, 0, now.Location(),
		)
		currentWeek = -1
		t.dayStart = 1
	}
	tl, _, err := parse(tsList[5], "week",0, 6, startTime)
	if err != nil {
		return err
	}

	if tl == nil {
		return errors.New("week format error, format not supported: */3")
	}
	t.Week = &TimeTicket{Ticket: tl, Cursor: getCursor(tl, currentWeek)}
	return nil
}

func (t *TaskTime) ComputeNextExecTime() {
	var (
		second, minute, hour int
		day, month int
		reset, dayIncrease bool
		minuteInc, hourInc, monthInc bool
	)

	// 根据上一次时间来计算下一个时间, 不根据当前时间计算
	preTime := t.NextExecTime
	// second
	if t.SecondInterval == 0 && t.Second != nil {
		if t.Second.Cursor >= len(t.Second.Ticket) || t.Second.Ticket[t.Second.Cursor] < preTime.Second() {
			minuteInc = true
			t.incMinute()
			t.Second.Cursor = 0
		}
		second = t.Second.Ticket[t.Second.Cursor]

		//
		t.Second.Cursor++
	} else {
		second = preTime.Second() + t.SecondInterval
		if t.secondStart > 0 {
			second = 0
			t.secondStart = 0
		}
		if second >= 60 {
			minuteInc = true
			second -= 60
			if t.incMinute() {
				second = 0
			}
		}
	}

	// minute
	if t.MinuteInterval == 0 && t.Minute != nil {
		if t.Minute.Cursor >= len(t.Minute.Ticket) || t.Minute.Ticket[t.Minute.Cursor] < preTime.Minute() {
			hourInc = true
			t.incHour()
			t.Minute.Cursor = 0
		}
		minute = t.Minute.Ticket[t.Minute.Cursor]
	} else {
		minute = preTime.Minute()
		if t.minuteStart > 0 {
			minute = 0
			t.minuteStart = 0
		}
		if minuteInc {
			minute += t.MinuteInterval
			if minute >= 60 {
				hourInc = true
				minute -= 60
				if t.incHour() {
					minute = 0
				}
			}
		}
	}
	// hour
	if t.HourInterval == 0 && t.Hour != nil {
		if t.Hour.Cursor >= len(t.Hour.Ticket) || t.Hour.Ticket[t.Hour.Cursor] < preTime.Hour() {
			dayIncrease = true
			t.incDay(preTime)
			t.incWeek(preTime)
			t.Hour.Cursor = 0
		}
		hour = t.Hour.Ticket[t.Hour.Cursor]
	} else {
		hour = preTime.Hour()
		if t.hourStart > 0 {
			hour = 0
			t.hourStart = 0
		}
		if hourInc {
			hour += t.HourInterval
			if hour >=  24 {
				dayIncrease = true
				hour -= 24

				// 立即reset
				if t.incDay(preTime) || t.incWeek(preTime) {
					hour = 0
				}
			}
		}
	}

	// process day & week
	if t.DayInterval == 0 {
		// 获取静态设定的时间，如 1,3,4  3-30 格式
		if t.Day != nil {
			if t.Day.Cursor >= len(t.Day.Ticket) || t.Day.Ticket[t.Day.Cursor] < preTime.Day() {
				monthInc = true
				t.incMonth()
				t.Day.Cursor = 0
			}
			day = t.Day.Ticket[t.Day.Cursor]
		} else {
			// 从周获取天
			if t.Week.Cursor >= len(t.Week.Ticket) {
				t.Week.Cursor = 0
			}

			month, day = GetMonthDayFromWeek(t.Week.Ticket[t.Week.Cursor], preTime)
			// 匹配月份
			if t.Month != nil {
				if month != t.Month.Ticket[t.Month.Cursor % len(t.Month.Ticket)] {
					t.Month.Cursor++
					if t.Month.Cursor >= len(t.Month.Ticket) || t.Month.Ticket[t.Month.Cursor] < month {
						t.resetMonth()
						reset = true
					}
					day = GetFirstWeekDay(t.Month.Ticket[t.Month.Cursor], t.Week.Ticket[t.Week.Cursor])
				}
			} else {
				month = int(preTime.Month()) + t.MonthInterval
				if month > 12 {
					month -= 12
					t.resetMonth()
					monthInc = true
				}

				day = GetFirstWeekDay(month, t.Week.Cursor)
			}
		}
	} else {
		// 动态计算天，处理如 */3, * 格式
		day = preTime.Day()
		if dayIncrease {
			maxDay := GetMaxDay(int(preTime.Month()))
			nextDay := preTime.Day() + t.DayInterval
			// 时间到下个月
			if nextDay > maxDay {
				monthInc = true
				t.incMonth()
				// 下个月第n天
				day = nextDay - maxDay

				// 如果下个月不在月份表里, 则置天数为1
				if t.Month != nil {
					if int(preTime.Month()) + 1 != t.Month.Ticket[t.Month.Cursor % len(t.Month.Ticket)] {
						day = 1
					}
				} else {
					if t.MonthInterval != 1 {
						day = 1
					}
				}
			} else {
				day = nextDay
			}
		}
	}

	// month
	if t.MonthInterval == 0 && t.Month != nil {
		if t.Month.Cursor >= len(t.Month.Ticket) ||
			(t.Month.Ticket[t.Month.Cursor] < int(preTime.Month()) && !reset) {
			t.resetMonth()
		}
		month = t.Month.Ticket[t.Month.Cursor]
	} else {
		month = int(preTime.Month())
		if monthInc {
			month += t.MonthInterval
			if month > 12 {
				month -= 12
				t.resetMonth()
			}
		}
	}

	t.NextExecTime = time.Date(
		Year, time.Month(month), day, hour, minute, second, 0, preTime.Location(),
	)
}

func (t *TaskTime) incMinute() bool {
	if t.MinuteInterval == 0 && t.Minute != nil {
		minLen := len(t.Minute.Ticket)
		t.Minute.Cursor++

		// 高位时间不连续 则重置低位时间, return true 则立即重置此次的minute值
		diff := t.Minute.Ticket[t.Minute.Cursor % minLen] - t.Minute.Ticket[(t.Minute.Cursor-1) % minLen]
		if diff != 1 && diff - 1 != -60 {
			t.resetSecond()
			return true
		}
		return false
	}
	if t.MinuteInterval > 1 {
		t.resetSecond()
		return true
	}
	return false
}

func (t *TaskTime) incHour() bool {
	if t.HourInterval == 0 && t.Hour != nil {
		hourLen := len(t.Hour.Ticket)
		t.Hour.Cursor++

		diff := t.Hour.Ticket[t.Hour.Cursor % hourLen] - t.Hour.Ticket[(t.Hour.Cursor-1) % hourLen]
		if diff != 1 && diff - 1 != -24 {
			t.resetMinute()
			return true
		}
		return false
	}
	if t.HourInterval > 1 {
		t.resetMinute()
		return true
	}
	return false
}

func (t *TaskTime) incDay(now time.Time) bool {
	if t.DayInterval == 0 && t.Day != nil {
		t.Day.Cursor++
		dayLen := len(t.Day.Ticket)
		diff := t.Day.Ticket[t.Day.Cursor%dayLen]-t.Day.Ticket[(t.Day.Cursor-1)%dayLen]

		maxDay := GetMaxDay(int(now.Month()))
		if diff != 1 && diff - 1 != -maxDay {
			t.resetHour()
			return true
		}
		return false
	}

	if t.DayInterval > 1 {
		t.resetHour()
		return true
	}
	return false
}

func (t *TaskTime) incMonth() bool {
	if t.MonthInterval == 0 && t.Month != nil {
		t.Month.Cursor++
		monLen := len(t.Month.Ticket)
		diff := t.Month.Ticket[t.Month.Cursor % monLen] - t.Month.Ticket[(t.Month.Cursor-1) % monLen]
		if diff != 1 && diff - 1 != -12 {
			t.resetDay()
			t.resetWeek()
			return true
		}
	}
	if t.MonthInterval > 1 {
		t.resetDay()
		t.resetWeek()
		return true
	}

	return false
}

func (t *TaskTime) incWeek(now time.Time) bool {
	if t.Week.Cursor >= len(t.Week.Ticket) {
		t.Week.Cursor = 0
		return false
	}
	if t.Week.Ticket[t.Week.Cursor] <= int(now.Weekday()) {
		t.Week.Cursor++
		weekLen := len(t.Week.Ticket)

		if t.Day == nil &&
			t.Week.Ticket[t.Week.Cursor % weekLen] - t.Week.Ticket[(t.Week.Cursor-1) % weekLen] != 1 {
			t.resetHour()
			return true
		}
	}
	return false
}

func (t *TaskTime) resetSecond() {
	t.secondStart = 1
	if t.Second != nil {
		t.Second.Cursor = 0
	}
}

func (t *TaskTime) resetMinute() {
	t.minuteStart = 1
	if t.Minute != nil {
		t.Minute.Cursor = 0
	}
}

func (t *TaskTime) resetHour()  {
	t.hourStart = 1
	if t.Hour != nil {
		t.Hour.Cursor = 0
	}
}

func (t *TaskTime) resetDay() {
	t.dayStart = 1
	if  t.Day != nil {
		t.Day.Cursor = 0
	}
}

func (t *TaskTime) resetMonth() {
	t.monthStart = 1
	if t.Month != nil {
		t.Month.Cursor = 0
	}
	Year++
}

func (t *TaskTime) resetWeek() {
	t.Week.Cursor = 0
}

func GetMonthDayFromWeek(weekday int, t time.Time) (int, int) {
	//t := time.Now()
	currentWeekDay := int(t.Weekday())

	if weekday == 0 {
		weekday = 7
	}

	shiftDay :=  weekday - currentWeekDay
	// 下周
	if shiftDay < 0 {
		shiftDay = (7 - currentWeekDay) + weekday
	}

	t = t.AddDate(0, 0, shiftDay)
	return int(t.Month()), t.Day()
}

func GetFirstWeekDay(month, weekday int) int {
	maxDay := GetMaxDay(month)
	for i := 1; i <= maxDay; i++ {
		t := time.Date(Year, time.Month(month), i, 0, 0, 0, 0, time.UTC)
		if t.Weekday() == time.Weekday(weekday) {
			return t.Day()
		}
	}
	return 0
}


func GetMaxDay(month int) int {
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			return 30
		}
		return 31
	}

	// 闰年2月
	if (Year % 4 == 0 && Year % 100 != 0) || Year % 400 == 0 {
		return 29
	}
	return 28
}

func getCursor(tl []int, current int) int {
	for i, v := range tl {
		if v == current {
			return i
		}
	}
	return 0
}