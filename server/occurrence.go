package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/mlog"
	"time"
	"encoding/json"
	"strings"
	"errors"
	"strconv"
	"github.com/google/uuid"
)

const (
	DEFAULT_TIME = "9:00AM"
)

type Occurrence struct {
	Id string

	Username string

	ReminderId string

	Occurrence time.Time

	Snoozed time.Time

	Repeat string
}

func (p *Plugin) CreateOccurrences(request *ReminderRequest) error {

	p.API.LogDebug("CreateOccurrences", request.Reminder.When)

	if strings.HasPrefix(request.Reminder.When, "in") {
		occurrences, inErr := p.in(request.Reminder.When)
		if inErr != nil {
			return inErr
		}

		return p.addOccurrences(request, occurrences)
	}

	if strings.HasPrefix(request.Reminder.When, "at") {
		occurrences, inErr := p.at(request.Reminder.When)
		if inErr != nil {
			return inErr
		}

		return p.addOccurrences(request, occurrences)
	}

	if strings.HasPrefix(request.Reminder.When, "on") {
		occurrences, inErr := p.on(request.Reminder.When)
		if inErr != nil {
			return inErr
		}

		return p.addOccurrences(request, occurrences)
	}

	if strings.HasPrefix(request.Reminder.When, "every") {
		occurrences, inErr := p.every(request.Reminder.When)
		if inErr != nil {
			return inErr
		}

		return p.addOccurrences(request, occurrences)
	}

	occurrences, freeErr := p.freeForm(request.Reminder.When)
	if freeErr != nil {
		return freeErr
	}

	return p.addOccurrences(request, occurrences)
}

func (p *Plugin) addOccurrences(request *ReminderRequest, occurrences []time.Time) error {

	loc, _ := time.LoadLocation("Europe/Warsaw")
	for _, o := range occurrences {

		repeat := ""

		if p.isRepeating(request) {
			repeat = request.Reminder.When
		}

		guid, gErr := uuid.NewRandom()
		if gErr != nil {
			p.API.LogError("failed to generate guid")
			return gErr
		}

		occurrence := Occurrence{guid.String(), request.Username, request.Reminder.Id, o.In(loc), time.Time{}, repeat}

		p.upsertOccurrence(occurrence)
		request.Reminder.Occurrences = append(request.Reminder.Occurrences, occurrence)

	}

	return nil
}

func (p *Plugin) isRepeating(request *ReminderRequest) bool {

	return strings.Contains(request.Reminder.When, "every") ||
		strings.Contains(request.Reminder.When, "sundays") ||
		strings.Contains(request.Reminder.When, "mondays") ||
		strings.Contains(request.Reminder.When, "tuesdays") ||
		strings.Contains(request.Reminder.When, "wednesdays") ||
		strings.Contains(request.Reminder.When, "thursdays") ||
		strings.Contains(request.Reminder.When, "fridays") ||
		strings.Contains(request.Reminder.When, "saturdays")

}


func (p *Plugin) upsertOccurrence(reminderOccurrence Occurrence) {

	bytes, err := p.API.KVGet(string(fmt.Sprintf("%v", reminderOccurrence.Occurrence)))
	if err != nil {
		p.API.LogError("failed KVGet %s", err)
		return
	}

	var reminderOccurrences []Occurrence
	roErr := json.Unmarshal(bytes, &reminderOccurrences)
	if roErr != nil {
		p.API.LogDebug("new occurrence " + string(fmt.Sprintf("%v", reminderOccurrence.Occurrence)))
	} else {
		p.API.LogDebug("existing " + fmt.Sprintf("%v", reminderOccurrences))
	}

	reminderOccurrences = append(reminderOccurrences, reminderOccurrence)
	ro, __ := json.Marshal(reminderOccurrences)

	if __ != nil {
		p.API.LogError("failed to marshal reminderOccurrences %s", reminderOccurrence.Id)
		return
	}

	key := string(fmt.Sprintf("%v", reminderOccurrence.Occurrence))
	p.API.LogDebug("Set occurences " + key)
	p.API.KVSet(key, ro)

}


func (p *Plugin) at(when string) (times []time.Time, err error) {
	loc, _ := time.LoadLocation("Europe/Warsaw")
	whenTrim := strings.Trim(when, " ")
	whenSplit := strings.Split(whenTrim, " ")
	normalizedWhen := strings.ToLower(whenSplit[1])

	if strings.Contains(when, "every") {

		dateTimeSplit := strings.Split(when, " "+"every"+" ")
		return p.every("every"+" "+dateTimeSplit[1]+" "+dateTimeSplit[0])

	} else if len(whenSplit) >= 3 &&
		(strings.EqualFold(whenSplit[2], "pm") ||
			strings.EqualFold(whenSplit[2], "am")) {

		if !strings.Contains(normalizedWhen, ":") {
			if len(normalizedWhen) >= 3 {
				hrs := string(normalizedWhen[:len(normalizedWhen)-2])
				mins := string(normalizedWhen[len(normalizedWhen)-2:])
				normalizedWhen = hrs + ":" + mins
			} else {
				normalizedWhen = normalizedWhen + ":00"
			}
		}
		t, pErr := time.ParseInLocation(time.Kitchen, normalizedWhen+strings.ToUpper(whenSplit[2]), loc)
		if pErr != nil {
			mlog.Error(fmt.Sprintf("%v", pErr))
		}

		now := time.Now().In(loc).Round(time.Hour * time.Duration(24))
		occurrence := t.AddDate(now.Year(), int(now.Month())-1, now.Day()-1)
		return []time.Time{p.chooseClosest(&occurrence, true)}, nil

	} else if strings.HasSuffix(normalizedWhen, "pm") ||
		strings.HasSuffix(normalizedWhen, "am") {

		if !strings.Contains(normalizedWhen, ":") {
			var s string
			var s2 string
			if len(normalizedWhen) == 3 {
				s = normalizedWhen[:len(normalizedWhen)-2]
				s2 = normalizedWhen[len(normalizedWhen)-2:]
			} else if len(normalizedWhen) >= 4 {
				s = normalizedWhen[:len(normalizedWhen)-4]
				s2 = normalizedWhen[len(normalizedWhen)-4:]
			}

			if len(s2) > 2 {
				normalizedWhen = s + ":" + s2
			} else {
				normalizedWhen = s + ":00" + s2
			}

		}
		t, pErr := time.ParseInLocation(time.Kitchen, strings.ToUpper(normalizedWhen), loc)
		if pErr != nil {
			mlog.Error(fmt.Sprintf("%v", pErr))
		}

		now := time.Now().In(loc).Round(time.Hour * time.Duration(24))
		occurrence := t.AddDate(now.Year(), int(now.Month())-1, now.Day()-1)
		return []time.Time{p.chooseClosest(&occurrence, true)}, nil

	}

	switch normalizedWhen {

	case "noon":

		now := time.Now().In(loc)

		noon, pErr := time.ParseInLocation(time.Kitchen, "12:00PM", loc)
		if pErr != nil {
			mlog.Error(fmt.Sprintf("%v", pErr))
			return []time.Time{}, pErr
		}

		noon = noon.AddDate(now.Year(), int(now.Month())-1, now.Day()-1)
		return []time.Time{p.chooseClosest(&noon, true)}, nil

	case "midnight":

		now := time.Now().In(loc)

		midnight, pErr := time.ParseInLocation(time.Kitchen, "12:00AM", loc)
		if pErr != nil {
			mlog.Error(fmt.Sprintf("%v", pErr))
			return []time.Time{}, pErr
		}

		midnight = midnight.AddDate(now.Year(), int(now.Month())-1, now.Day()-1)
		return []time.Time{p.chooseClosest(&midnight, true)}, nil

	case "one",
		"two",
		"three",
		"four",
		"five",
		"six",
		"seven",
		"eight",
		"nine",
		"ten",
		"eleven",
		"twelve":

		nowkit := time.Now().In(loc).Format(time.Kitchen)
		ampm := string(nowkit[len(nowkit)-2:])

		num, wErr := p.wordToNumber(normalizedWhen)
		if wErr != nil {
			return []time.Time{}, wErr
		}

		wordTime, _ := time.ParseInLocation(time.Kitchen, strconv.Itoa(num)+":00"+ampm, loc)
		return []time.Time{p.chooseClosest(&wordTime, false)}, nil

	case "0",
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
		"10",
		"11",
		"12":

		nowkit := time.Now().In(loc).Format(time.Kitchen)
		ampm := string(nowkit[len(nowkit)-2:])

		num, wErr := strconv.Atoi(normalizedWhen)
		if wErr != nil {
			return []time.Time{}, wErr
		}

		wordTime, _ := time.ParseInLocation(time.Kitchen, strconv.Itoa(num)+":00"+ampm, loc)
		return []time.Time{p.chooseClosest(&wordTime, false)}, nil

	default:

		if !strings.Contains(normalizedWhen, ":") && len(normalizedWhen) >= 3 {
			s := normalizedWhen[:len(normalizedWhen)-2]
			normalizedWhen = s + ":" + normalizedWhen[len(normalizedWhen)-2:]
		}

		timeSplit := strings.Split(normalizedWhen, ":")
		hr, _ := strconv.Atoi(timeSplit[0])
		ampm := "am"
		dayInterval := false

		if hr > 11 {
			ampm = "pm"
		}
		if hr > 12 {
			hr -= 12
			dayInterval = true
			timeSplit[0] = strconv.Itoa(hr)
			normalizedWhen = strings.Join(timeSplit, ":")
		}

		t, pErr := time.ParseInLocation(time.Kitchen, strings.ToUpper(normalizedWhen+ampm), loc)
		if pErr != nil {
			return []time.Time{}, pErr
		}

		now := time.Now().In(loc).Round(time.Hour * time.Duration(24))
		occurrence := t.AddDate(now.Year(), int(now.Month())-1, now.Day()-1)
		return []time.Time{p.chooseClosest(&occurrence, dayInterval)}, nil

	}

	return []time.Time{}, errors.New("could not format 'at'")
}

func (p *Plugin) in(when string) (times []time.Time, err error) {
	loc, _ := time.LoadLocation("Europe/Warsaw")

	whenSplit := strings.Split(when, " ")
	value := whenSplit[1]
	units := whenSplit[len(whenSplit)-1]

	switch units {
	case "seconds",
		"second",
		"secs",
		"sec",
		"s":

		i, e := strconv.Atoi(value)

		if e != nil {
			num, wErr := p.wordToNumber(value)
			if wErr != nil {
				mlog.Error(fmt.Sprintf("%v", wErr))
				return []time.Time{}, wErr
			}
			i = num
		}

		times = append(times, time.Now().In(loc).Round(time.Second).Add(time.Second*time.Duration(i)))

		return times, nil

	case "minutes",
		"minute",
		"min":

		i, e := strconv.Atoi(value)

		if e != nil {
			num, wErr := p.wordToNumber(value)
			if wErr != nil {
				mlog.Error(fmt.Sprintf("%v", wErr))
				return []time.Time{}, wErr
			}
			i = num
		}

		times = append(times, time.Now().In(loc).Round(time.Second).Add(time.Minute*time.Duration(i)))
		return times, nil

	case "hours",
		"hour",
		"hrs",
		"hr":

		i, e := strconv.Atoi(value)

		if e != nil {
			num, wErr := p.wordToNumber(value)
			if wErr != nil {
				mlog.Error(fmt.Sprintf("%v", wErr))
				return []time.Time{}, wErr
			}
			i = num
		}

		times = append(times, time.Now().In(loc).Round(time.Second).Add(time.Hour*time.Duration(i)))

		return times, nil

	case "days",
		"day",
		"d":

		i, e := strconv.Atoi(value)

		if e != nil {
			num, wErr := p.wordToNumber(value)
			if wErr != nil {
				mlog.Error(fmt.Sprintf("%v", wErr))
				return []time.Time{}, wErr
			}
			i = num
		}

		times = append(times, time.Now().In(loc).Round(time.Second).Add(time.Hour*24*time.Duration(i)))

		return times, nil

	case "weeks",
		"week",
		"wks",
		"wk":

		i, e := strconv.Atoi(value)

		if e != nil {
			num, wErr := p.wordToNumber(value)
			if wErr != nil {
				mlog.Error(fmt.Sprintf("%v", wErr))
				return []time.Time{}, wErr
			}
			i = num
		}

		times = append(times, time.Now().In(loc).Round(time.Second).Add(time.Hour*24*7*time.Duration(i)))

		return times, nil

	case "months",
		"month",
		"m":

		i, e := strconv.Atoi(value)

		if e != nil {
			num, wErr := p.wordToNumber(value)
			if wErr != nil {
				mlog.Error(fmt.Sprintf("%v", wErr))
				return []time.Time{}, wErr
			}
			i = num
		}

		times = append(times, time.Now().In(loc).Round(time.Second).Add(time.Hour*24*30*time.Duration(i)))

		return times, nil

	case "years",
		"year",
		"yr",
		"y":

		i, e := strconv.Atoi(value)

		if e != nil {
			num, wErr := p.wordToNumber(value)
			if wErr != nil {
				mlog.Error(fmt.Sprintf("%v", wErr))
				return []time.Time{}, wErr
			}
			i = num
		}

		times = append(times, time.Now().In(loc).Round(time.Second).Add(time.Hour*24*365*time.Duration(i)))

		return times, nil

	default:
		return nil, errors.New("could not format 'in'")
	}

	return nil, errors.New("could not format 'in'")
}

func (p *Plugin) on(when string) (times []time.Time, err error) {
	loc, _ := time.LoadLocation("Europe/Warsaw")

	whenTrim := strings.Trim(when, " ")
	whenSplit := strings.Split(whenTrim, " ")

	if len(whenSplit) < 2 {
		return []time.Time{}, errors.New("not enough arguments")
	}

	chronoUnit := strings.ToLower(strings.Join(whenSplit[1:], " "))
	dateTimeSplit := strings.Split(chronoUnit, " "+"at"+" ")
	chronoDate := dateTimeSplit[0]
	chronoTime := DEFAULT_TIME
	if len(dateTimeSplit) > 1 {
		chronoTime = dateTimeSplit[1]
	}

	dateUnit, ndErr := p.normalizeDate(chronoDate)
	if ndErr != nil {
		return []time.Time{}, ndErr
	}
	timeUnit, ntErr := p.normalizeTime(chronoTime)
	if ntErr != nil {
		return []time.Time{}, ntErr
	}

	switch dateUnit {
	case "sunday",
		"monday",
		"tuesday",
		"wednesday",
		"thursday",
		"friday",
		"saturday":

		todayWeekDayNum := int(time.Now().In(loc).Weekday())
		weekDayNum := p.weekDayNumber(dateUnit)
		day := 0

		if weekDayNum < todayWeekDayNum {
			day = 7 - (todayWeekDayNum - weekDayNum)
		} else if weekDayNum > todayWeekDayNum {
			day = weekDayNum - todayWeekDayNum
		} else {
			day = 7
		}

		timeUnitSplit := strings.Split(timeUnit, ":")
		hr, _ := strconv.Atoi(timeUnitSplit[0])
		ampm := strings.ToUpper("am")

		if hr > 11 {
			ampm = strings.ToUpper("pm")
		}
		if hr > 12 {
			hr -= 12
			timeUnitSplit[0] = strconv.Itoa(hr)
		}

		timeUnit = timeUnitSplit[0] + ":" + timeUnitSplit[1] + ampm
		wallClock, pErr := time.ParseInLocation(time.Kitchen, timeUnit, loc)
		if pErr != nil {
			return []time.Time{}, pErr
		}

		nextDay := time.Now().In(loc).AddDate(0, 0, day)
		occurrence := wallClock.AddDate(nextDay.Year(), int(nextDay.Month())-1, nextDay.Day()-1)

		return []time.Time{p.chooseClosest(&occurrence, false)}, nil

		break
	case "mondays",
		"tuesdays",
		"wednesdays",
		"thursdays",
		"fridays",
		"saturdays",
		"sundays":

		return p.every(
			"every"+" "+
				dateUnit[:len(dateUnit)-1]+" "+
				"at"+" "+
				timeUnit[:len(timeUnit)-3])

		break
	}

	dateSplit := p.regSplit(dateUnit, "T|Z")

	if len(dateSplit) < 3 {
		timeSplit := strings.Split(dateSplit[1], "-")
		t, tErr := time.ParseInLocation(time.RFC3339, dateSplit[0]+"T"+timeUnit+"-"+timeSplit[1], loc)
		if tErr != nil {
			return []time.Time{}, tErr
		}
		return []time.Time{t}, nil
	} else {
		t, tErr := time.ParseInLocation(time.RFC3339, dateSplit[0]+"T"+timeUnit+"Z"+dateSplit[2], loc)
		if tErr != nil {
			return []time.Time{}, tErr
		}
		return []time.Time{t}, nil
	}

}

func (p *Plugin) every(when string) (times []time.Time, err error) {
	loc, _ := time.LoadLocation("Europe/Warsaw")

	whenTrim := strings.Trim(when, " ")
	whenSplit := strings.Split(whenTrim, " ")

	if len(whenSplit) < 2 {
		return []time.Time{}, errors.New("not enough arguments")
	}

	var everyOther bool
	chronoUnit := strings.ToLower(strings.Join(whenSplit[1:], " "))
	otherSplit := strings.Split(chronoUnit, "other")
	if len(otherSplit) == 2 {
		chronoUnit = strings.Trim(otherSplit[1], " ")
		everyOther = true
	}
	dateTimeSplit := strings.Split(chronoUnit, " "+"at"+" ")
	chronoDate := dateTimeSplit[0]
	chronoTime := DEFAULT_TIME
	if len(dateTimeSplit) > 1 {
		chronoTime = strings.Trim(dateTimeSplit[1], " ")
	}

	days := p.regSplit(chronoDate, "(and)|(,)")

	for _, chrono := range days {

		dateUnit, ndErr := p.normalizeDate(strings.Trim(chrono, " "))
		if ndErr != nil {
			return []time.Time{}, ndErr
		}
		timeUnit, ntErr := p.normalizeTime(chronoTime,)
		if ntErr != nil {
			return []time.Time{}, ntErr
		}

		switch dateUnit {
		case "day":
			d := 1
			if everyOther {
				d = 2
			}

			timeUnitSplit := strings.Split(timeUnit, ":")
			hr, _ := strconv.Atoi(timeUnitSplit[0])
			ampm := strings.ToUpper("am")

			if hr > 11 {
				ampm = strings.ToUpper("pm")
			}
			if hr > 12 {
				hr -= 12
				timeUnitSplit[0] = strconv.Itoa(hr)
			}

			timeUnit = timeUnitSplit[0] + ":" + timeUnitSplit[1] + ampm
			wallClock, pErr := time.ParseInLocation(time.Kitchen, timeUnit, loc)
			if pErr != nil {
				return []time.Time{}, pErr
			}

			nextDay := time.Now().In(loc).AddDate(0, 0, d)
			occurrence := wallClock.AddDate(nextDay.Year(), int(nextDay.Month())-1, nextDay.Day()-1)
			times = append(times, p.chooseClosest(&occurrence, false))

			break
		case "sunday",
			"monday",
			"tuesday",
			"wednesday",
			"thursday",
			"friday",
			"saturday":

			todayWeekDayNum := int(time.Now().In(loc).Weekday())
			weekDayNum := p.weekDayNumber(dateUnit)
			day := 0

			if weekDayNum < todayWeekDayNum {
				day = 7 - (todayWeekDayNum - weekDayNum)
			} else if weekDayNum >= todayWeekDayNum {
				day = weekDayNum - todayWeekDayNum
			} else {
				day = 7
			}

			timeUnitSplit := strings.Split(timeUnit, ":")
			hr, _ := strconv.Atoi(timeUnitSplit[0])
			ampm := strings.ToUpper("am")

			if hr > 11 {
				ampm = strings.ToUpper("pm")
			}
			if hr > 12 {
				hr -= 12
				timeUnitSplit[0] = strconv.Itoa(hr)
			}

			timeUnit = timeUnitSplit[0] + ":" + timeUnitSplit[1] + ampm
			wallClock, pErr := time.ParseInLocation(time.Kitchen, timeUnit, loc)
			if pErr != nil {
				return []time.Time{}, pErr
			}

			nextDay := time.Now().In(loc).AddDate(0, 0, day)
			occurrence := wallClock.AddDate(nextDay.Year(), int(nextDay.Month())-1, nextDay.Day()-1)
			times = append(times, p.chooseClosest(&occurrence, false))
			break
		default:

			dateSplit := p.regSplit(dateUnit, "T|Z")

			if len(dateSplit) < 3 {
				timeSplit := strings.Split(dateSplit[1], "-")
				t, tErr := time.ParseInLocation(time.RFC3339, dateSplit[0]+"T"+timeUnit+"-"+timeSplit[1], loc)
				if tErr != nil {
					return []time.Time{}, tErr
				}
				times = append(times, t)
			} else {
				t, tErr := time.ParseInLocation(time.RFC3339, dateSplit[0]+"T"+timeUnit+"Z"+dateSplit[2], loc)
				if tErr != nil {
					return []time.Time{}, tErr
				}
				times = append(times, t)
			}

		}

	}

	return times, nil

}

func (p *Plugin) freeForm(when string) (times []time.Time, err error) {
	loc, _ := time.LoadLocation("Europe/Warsaw")

	whenTrim := strings.Trim(when, " ")
	chronoUnit := strings.ToLower(whenTrim)
	dateTimeSplit := strings.Split(chronoUnit, " "+"at"+" ")
	chronoTime := DEFAULT_TIME
	chronoDate := dateTimeSplit[0]

	if len(dateTimeSplit) > 1 {
		chronoTime = dateTimeSplit[1]
	}
	dateUnit, ndErr := p.normalizeDate(chronoDate)
	if ndErr != nil {
		return []time.Time{}, ndErr
	}
	timeUnit, ntErr := p.normalizeTime(chronoTime)
	if ntErr != nil {
		return []time.Time{}, ntErr
	}
	timeUnit = chronoTime

	switch dateUnit {
	case "today":
		return p.at("at"+" "+timeUnit)
	case "tomorrow":
		return p.on("on"+" "+ time.Now().In(loc).Add(time.Hour*24).Weekday().String()+" at "+timeUnit)
	case "everyday":
		return p.every("every day at " + timeUnit)
	case "mondays",
		"tuesdays",
		"wednesdays",
		"thursdays",
		"fridays",
		"saturdays",
		"sundays":
		return p.every(
			"every"+" "+
				dateUnit[:len(dateUnit)-1]+" "+
				"at"+" "+
				timeUnit)
	case "monday",
		"tuesday",
		"wednesday",
		"thursday",
		"friday",
		"saturday",
		"sunday":
		return p.on(
			"on"+" "+
				dateUnit+" "+
				"at"+" "+
				timeUnit)
	default:
		return p.on(
			"on"+" "+
				dateUnit[:len(dateUnit)-1]+" "+
				"at"+" "+
				timeUnit)
	}

	return []time.Time{}, nil
}
