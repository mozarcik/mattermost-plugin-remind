package main

import (
	"errors"
	"fmt"
	"github.com/mattermost/mattermost-server/mlog"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (p *Plugin) ParseRequest(request ReminderRequest) (string, string, string, error) {

	commandSplit := strings.Split(request.Payload, " ")

	p.API.LogDebug("parseRequest " + fmt.Sprintf("%v", request))
	p.API.LogDebug(request.Payload)

	if strings.HasPrefix(request.Payload, "me") ||
		strings.HasPrefix(request.Payload, "~") ||
		strings.HasPrefix(request.Payload, "@") {

		p.API.LogDebug("found target")

		firstIndex := strings.Index(request.Payload, "\"")
		lastIndex := strings.LastIndex(request.Payload, "\"")

		if firstIndex > -1 && lastIndex > -1 && firstIndex != lastIndex { // has quotes

			message := request.Payload[firstIndex:lastIndex+1]

			p.API.LogDebug(message)

			when := strings.Replace(request.Payload, message, "", -1)
			when = strings.Replace(when, commandSplit[0], "", -1)
			when = strings.Trim(when," ")

			p.API.LogDebug("quotes when (" + fmt.Sprintf("%v", firstIndex) + " " + fmt.Sprintf("%v", lastIndex) + ") " + when)

			message = strings.Replace(message,"\"","",-1)

			return commandSplit[0], when, message, nil
		}

		p.API.LogDebug("no quotes when " + fmt.Sprintf("%v", firstIndex) + " " + fmt.Sprintf("%v", lastIndex))

		if wErr := p.findWhen(&request); wErr != nil {
			return "", "", "", wErr
		}


		message := strings.Replace(request.Payload, request.Reminder.When, "", -1)
		message = strings.Replace(message, commandSplit[0], "", -1)
		message = strings.Trim(message, " \"")

		return commandSplit[0], request.Reminder.When, message, nil
	}

	return "", "", "", errors.New("unrecognized Target")
}

func (p *Plugin) findWhen(request *ReminderRequest) error {
	inIndex := strings.Index(request.Payload, " in ")
	if inIndex > -1 {
		request.Reminder.When = strings.Trim(request.Payload[inIndex:], " ")
		return nil
	}

	everyIndex := strings.Index(request.Payload, " every ")
	atIndex := strings.Index(request.Payload, " at ")
	if (everyIndex > -1 && atIndex == -1) || (atIndex > everyIndex) && everyIndex != -1 {
		request.Reminder.When = strings.Trim(request.Payload[everyIndex:], " ")
		return nil
	}

	onIndex := strings.Index(request.Payload, " on ")
	if onIndex > -1 {
		request.Reminder.When = strings.Trim(request.Payload[onIndex:], " ")
		return nil
	}

	everydayIndex := strings.Index(request.Payload, " everyday ")
	atIndex = strings.Index(request.Payload, " at ")
	if (everydayIndex > -1 && atIndex >= -1) && (atIndex > everydayIndex) {
		request.Reminder.When = strings.Trim(request.Payload[everydayIndex:], " ")
		return nil
	}

	todayIndex := strings.Index(request.Payload, " today ")
	atIndex = strings.Index(request.Payload, " at ")
	if (todayIndex > -1 && atIndex >= -1) && (atIndex > todayIndex) {
		request.Reminder.When = strings.Trim(request.Payload[todayIndex:], " ")
		return nil
	}

	tomorrowIndex := strings.Index(request.Payload, " tomorrow ")
	atIndex = strings.Index(request.Payload, " at ")
	if (tomorrowIndex > -1 && atIndex >= -1) && (atIndex > tomorrowIndex) {
		request.Reminder.When = strings.Trim(request.Payload[tomorrowIndex:], " ")
		return nil
	}

	mondayIndex := strings.Index(request.Payload, " monday ")
	atIndex = strings.Index(request.Payload, " at ")
	if (mondayIndex > -1 && atIndex >= -1) && (atIndex > mondayIndex) {
		request.Reminder.When = strings.Trim(request.Payload[mondayIndex:], " ")
		return nil
	}

	tuesdayIndex := strings.Index(request.Payload, " tuesday ")
	atIndex = strings.Index(request.Payload, " at ")
	if (tuesdayIndex > -1 && atIndex >= -1) && (atIndex > tuesdayIndex) {
		request.Reminder.When = strings.Trim(request.Payload[tuesdayIndex:], " ")
		return nil
	}

	wednesdayIndex := strings.Index(request.Payload, " wednesday ")
	atIndex = strings.Index(request.Payload, " at ")
	if (wednesdayIndex > -1 && atIndex >= -1) && (atIndex > wednesdayIndex) {
		request.Reminder.When = strings.Trim(request.Payload[wednesdayIndex:], " ")
		return nil
	}

	thursdayIndex := strings.Index(request.Payload, " thursday ")
	atIndex = strings.Index(request.Payload, " at ")
	if (thursdayIndex > -1 && atIndex >= -1) && (atIndex > thursdayIndex) {
		request.Reminder.When = strings.Trim(request.Payload[thursdayIndex:], " ")
		return nil
	}

	fridayIndex := strings.Index(request.Payload, " friday ")
	atIndex = strings.Index(request.Payload, " at ")
	if (fridayIndex > -1 && atIndex >= -1) && (atIndex > fridayIndex) {
		request.Reminder.When = strings.Trim(request.Payload[fridayIndex:], " ")
		return nil
	}

	saturdayIndex := strings.Index(request.Payload, " saturday ")
	atIndex = strings.Index(request.Payload, " at ")
	if (saturdayIndex > -1 && atIndex >= -1) && (atIndex > saturdayIndex) {
		request.Reminder.When = strings.Trim(request.Payload[saturdayIndex:], " ")
		return nil
	}

	sundayIndex := strings.Index(request.Payload, " sunday ")
	atIndex = strings.Index(request.Payload, " at ")
	if (sundayIndex > -1 && atIndex >= -1) && (atIndex > sundayIndex) {
		request.Reminder.When = strings.Trim(request.Payload[sundayIndex:], " ")
		return nil
	}

	atIndex = strings.Index(request.Payload, " at ")
	everyIndex = strings.Index(request.Payload, " every ")
	if (atIndex > -1 && everyIndex >= -1) || (everyIndex > atIndex) && atIndex != -1 {
		request.Reminder.When = strings.Trim(request.Payload[atIndex:], " ")
		return nil
	}

	textSplit := strings.Split(request.Payload, " ")

	if len(textSplit) == 1 {
		request.Reminder.When = textSplit[0]
		return nil
	}

	lastWord := textSplit[len(textSplit)-2] + " " + textSplit[len(textSplit)-1]
	_, dErr := p.normalizeDate(lastWord)
	if dErr == nil {
		request.Reminder.When = lastWord
		return nil
	} else {
		lastWord = textSplit[len(textSplit)-1]

		switch lastWord {
		case "tomorrow":
			request.Reminder.When = lastWord
			return nil
		case "everyday",
			"mondays",
			"tuesdays",
			"wednesdays",
			"thursdays",
			"fridays",
			"saturdays",
			"sundays":
			request.Reminder.When = lastWord
		default:
			break
		}

		_, dErr = p.normalizeDate(lastWord)
		if dErr == nil {
			request.Reminder.When = lastWord
			return nil
		} else {
			if len(textSplit) < 3 {
				return errors.New("unable to find when")
			}
			var firstWord string
			switch textSplit[1] {
			case "at":
				firstWord = textSplit[2]
				request.Reminder.When = textSplit[1] + " " + firstWord
				return nil
			case "in",
				"on":
				if len(textSplit) < 4 {
					return errors.New("unable to find when")
				}
				firstWord = textSplit[2] + " " + textSplit[3]
				request.Reminder.When = textSplit[1] + " " + firstWord
				return nil
			case "tomorrow",
				"monday",
				"tuesday",
				"wednesday",
				"thursday",
				"friday",
				"saturday",
				"sunday":
				firstWord = textSplit[1]
				request.Reminder.When = firstWord
				return nil
			default:
				break
			}
		}

	}

	return errors.New("unable to find when")
}


func (p *Plugin) normalizeDate(text string) (string, error) {
	loc, _ := time.LoadLocation("Europe/Warsaw")

	date := strings.ToLower(text)
	if strings.EqualFold("day", date) {
		return date, nil
	} else if strings.EqualFold("today", date) {
		return date, nil
	} else if strings.EqualFold("everyday", date) {
		return date, nil
	} else if strings.EqualFold("tomorrow", date) {
		return date, nil
	}

	switch date {
	case "mon",
		"monday":
		return "monday", nil
	case "tues",
		"tuesday":
		return "tuesday", nil
	case "wed",
		"wednes",
		"wednesday":
		return "wednesday", nil
	case "thur",
		"thursday":
		return "thursday", nil
	case "fri",
		"friday":
		return "friday", nil
	case "sat",
		"satur",
		"saturday":
		return "saturday", nil
	case "sun",
		"sunday":
		return "sunday", nil
	case "mondays",
		"tuesdays",
		"wednesdays",
		"thursdays",
		"fridays",
		"saturdays",
		"sundays":
		return date, nil
	}

	if strings.Contains(date, "jan") ||
		strings.Contains(date, "january") ||
		strings.Contains(date, "feb") ||
		strings.Contains(date, "february") ||
		strings.Contains(date, "mar") ||
		strings.Contains(date, "march") ||
		strings.Contains(date, "apr") ||
		strings.Contains(date, "april") ||
		strings.Contains(date, "may") ||
		strings.Contains(date, "june") ||
		strings.Contains(date, "july") ||
		strings.Contains(date, "aug") ||
		strings.Contains(date, "august") ||
		strings.Contains(date, "sept") ||
		strings.Contains(date, "september") ||
		strings.Contains(date, "oct") ||
		strings.Contains(date, "october") ||
		strings.Contains(date, "nov") ||
		strings.Contains(date, "november") ||
		strings.Contains(date, "dec") ||
		strings.Contains(date, "december") {

		date = strings.Replace(date, ",", "", -1)
		parts := strings.Split(date, " ")

		switch len(parts) {
		case 1:
			break
		case 2:
			if len(parts[1]) > 2 {
				parts[1] = p.daySuffix(parts[1])
			}
			if _, err := strconv.Atoi(parts[1]); err != nil {
				if wn, wErr := p.wordToNumber(parts[1]); wErr == nil {
					parts[1] = strconv.Itoa(wn)
				}
			}

			parts = append(parts, fmt.Sprintf("%v", time.Now().In(loc).Year()))

			break
		case 3:
			if len(parts[1]) > 2 {
				parts[1] = p.daySuffix(parts[1])
			}

			if _, err := strconv.Atoi(parts[1]); err != nil {
				if wn, wErr := p.wordToNumber(parts[1]); wErr == nil {
					parts[1] = strconv.Itoa(wn)
				} else {
					mlog.Error(wErr.Error())
				}

				if _, pErr := strconv.Atoi(parts[2]); pErr != nil {
					return "", pErr
				}
			}

			break
		default:
			return "", errors.New("unrecognized date format")
		}

		switch parts[0] {
		case "jan",
			"january":
			parts[0] = "01"
			break
		case "feb",
			"february":
			parts[0] = "02"
			break
		case "mar",
			"march":
			parts[0] = "03"
			break
		case "apr",
			"april":
			parts[0] = "04"
			break
		case "may":
			parts[0] = "05"
			break
		case "june":
			parts[0] = "06"
			break
		case "july":
			parts[0] = "07"
			break
		case "aug",
			"august":
			parts[0] = "08"
			break
		case "sept",
			"september":
			parts[0] = "09"
			break
		case "oct",
			"october":
			parts[0] = "10"
			break
		case "nov",
			"november":
			parts[0] = "11"
			break
		case "dec",
			"december":
			parts[0] = "12"
			break
		default:
			return "", errors.New("month not found")
		}

		if len(parts[1]) < 2 {
			parts[1] = "0" + parts[1]
		}
		return parts[2] + "-" + parts[0] + "-" + parts[1] + "T00:00:00Z", nil

	} else if match, _ := regexp.MatchString("^(([0-9]{2}|[0-9]{1})(-|/)([0-9]{2}|[0-9]{1})((-|/)([0-9]{4}|[0-9]{2}))?)", date); match {

		date := p.regSplit(date, "-|/")

		switch len(date) {
		case 2:
			year := time.Now().In(loc).Year()
			month, mErr := strconv.Atoi(date[0])
			if mErr != nil {
				return "", mErr
			}
			day, dErr := strconv.Atoi(date[1])
			if dErr != nil {
				return "", dErr
			}

			return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc).Format(time.RFC3339), nil

		case 3:
			if len(date[2]) == 2 {
				date[2] = "20" + date[2]
			}
			year, yErr := strconv.Atoi(date[2])
			if yErr != nil {
				return "", yErr
			}
			month, mErr := strconv.Atoi(date[0])
			if mErr != nil {
				return "", mErr
			}
			day, dErr := strconv.Atoi(date[1])
			if dErr != nil {
				return "", dErr
			}

			return time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc).Format(time.RFC3339), nil

		default:
			return "", errors.New("unrecognized date")
		}

	} else { //single number day

		var dayInt int
		day := p.daySuffix(date)

		if d, nErr := strconv.Atoi(day); nErr != nil {
			if wordNum, wErr := p.wordToNumber(date); wErr != nil {
				return "", wErr
			} else {
				day = strconv.Itoa(wordNum)
				dayInt = wordNum
			}
		} else {
			dayInt = d
		}

		month := time.Now().In(loc).Month()
		year := time.Now().In(loc).Year()

		var t time.Time
		t = time.Date(year, month, dayInt, 0, 0, 0, 0, loc)
		if t.Before(time.Now().In(loc)) {
			t = t.AddDate(0, 1, 0)
		}

		return t.Format(time.RFC3339), nil

	}

	return "", errors.New("unrecognized time")
}

func (p *Plugin) normalizeTime(text string) (string, error) {
	loc, _ := time.LoadLocation("Europe/Warsaw")

	switch text {
	case "noon":
		return "12:00:00", nil
	case "midnight":
		return "00:00:00", nil
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

		num, wErr := p.wordToNumber(text)
		if wErr != nil {
			mlog.Error(fmt.Sprintf("%v", wErr))
			return "", wErr
		}

		wordTime := time.Now().In(loc).Round(time.Hour).Add(time.Hour * time.Duration(num+2))

		dateTimeSplit := p.regSplit(p.chooseClosest(&wordTime, false).Format(time.RFC3339), "T|Z")

		switch len(dateTimeSplit) {
		case 2:
			tzSplit := strings.Split(dateTimeSplit[1], "-")
			return tzSplit[0], nil
			break
		case 3:
			break
		default:
			return "", errors.New("unrecognized dateTime format")
		}

		return dateTimeSplit[1], nil

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
		"12",
		"13",
		"14",
		"15",
		"16",
		"17",
		"18",
		"19",
		"20",
		"21",
		"22",
		"23":

		num, nErr := strconv.Atoi(text)
		if nErr != nil {
			return "", nErr
		}

		numTime := time.Now().In(loc).Round(time.Hour).Add(time.Hour * time.Duration(num+2))
		dateTimeSplit := p.regSplit(p.chooseClosest(&numTime, false).Format(time.RFC3339), "T|Z")

		switch len(dateTimeSplit) {
		case 2:
			tzSplit := strings.Split(dateTimeSplit[1], "-")
			return tzSplit[0], nil
			break
		case 3:
			break
		default:
			return "", errors.New("unrecognized dateTime format")
		}

		return dateTimeSplit[1], nil

	default:
		break
	}

	t := text
	if match, _ := regexp.MatchString("(1[012]|[1-9]):[0-5][0-9](\\s)?(?i)(am|pm)", t); match { // 12:30PM, 12:30 pm

		t = strings.ToUpper(strings.Replace(t, " ", "", -1))
		test, tErr := time.ParseInLocation(time.Kitchen, t, loc)
		if tErr != nil {
			return "", tErr
		}

		dateTimeSplit := p.regSplit(test.Format(time.RFC3339), "T|Z")
		if len(dateTimeSplit) != 3 {
			return "", errors.New("unrecognized dateTime format")
		}

		return dateTimeSplit[1], nil
	} else if match, _ := regexp.MatchString("(1[012]|[1-9]):[0-5][0-9]", t); match { // 12:30

		nowkit := time.Now().In(loc).Format(time.Kitchen)
		ampm := string(nowkit[len(nowkit)-2:])
		timeUnitSplit := strings.Split(t, ":")
		hr, _ := strconv.Atoi(timeUnitSplit[0])

		if hr > 11 {
			ampm = strings.ToUpper("pm")
		}
		if hr > 12 {
			hr -= 12
			timeUnitSplit[0] = strconv.Itoa(hr)
		}

		t = timeUnitSplit[0] + ":" + timeUnitSplit[1] + ampm

		test, tErr := time.ParseInLocation(time.Kitchen, t, loc)
		if tErr != nil {
			return "", tErr
		}

		dateTimeSplit := p.regSplit(p.chooseClosest(&test, false).Format(time.RFC3339), "T|Z")
		if len(dateTimeSplit) != 3 {
			return "", errors.New("unrecognized dateTime format")
		}

		return dateTimeSplit[1], nil

	} else if match, _ := regexp.MatchString("(1[012]|[1-9])(\\s)?(?i)(am|pm)", t); match { // 5PM, 7 am

		nowkit := time.Now().In(loc).Format(time.Kitchen)
		ampm := string(nowkit[len(nowkit)-2:])

		timeSplit := p.regSplit(t, "(?i)(am|pm)")

		test, tErr := time.ParseInLocation(time.Kitchen, timeSplit[0]+":00"+ampm, loc)
		if tErr != nil {
			return "", tErr
		}

		dateTimeSplit := p.regSplit(p.chooseClosest(&test, false).Format(time.RFC3339), "T|Z")
		if len(dateTimeSplit) != 3 {
			return "", errors.New("unrecognized dateTime format")
		}

		return dateTimeSplit[1], nil
	} else if match, _ := regexp.MatchString("(1[012]|[1-9])[0-5][0-9]", t); match { // 1200

		return t[:len(t)-2] + ":" + t[len(t)-2:] + ":00", nil

	}

	return "", errors.New("unable to normalize time")
}

func (p *Plugin) daySuffix(day string) string {

	daySuffixes := []string{
		"0th",
		"1st",
		"2nd",
		"3rd",
		"4th",
		"5th",
		"6th",
		"7th",
		"8th",
		"9th",
		"10th",
		"11th",
		"12th",
		"13th",
		"14th",
		"15th",
		"16th",
		"17th",
		"18th",
		"19th",
		"20th",
		"21st",
		"22nd",
		"23rd",
		"24th",
		"25th",
		"26th",
		"27th",
		"28th",
		"29th",
		"30th",
		"31st",
	}
	for _, suffix := range daySuffixes {
		if suffix == day {
			day = day[:len(day)-2]
			break
		}
	}
	return day
}

func (p *Plugin) wordToNumber(word string) (int, error) {

	var sum int
	var temp int
	var previous int

	numbers := make(map[string]int)
	onumbers := make(map[string]int)
	tnumbers := make(map[string]int)

	numbers["zero"] = 0
	numbers["one"] = 1
	numbers["two"] = 2
	numbers["three"] = 3
	numbers["four"] = 4
	numbers["five"] = 5
	numbers["six"] = 6
	numbers["seven"] = 7
	numbers["eight"] = 8
	numbers["nine"] = 9
	numbers["ten"] = 10
	numbers["eleven"] = 11
	numbers["twelve"] = 12
	numbers["thirteen"] = 13
	numbers["fourteen"] = 14
	numbers["fifteen"] = 15
	numbers["sixteen"] = 16
	numbers["seventeen"] = 17
	numbers["eighteen"] = 18
	numbers["nineteen"] = 19

	tnumbers["twenty"] = 20
	tnumbers["thirty"] = 30
	tnumbers["forty"] = 40
	tnumbers["fifty"] = 50
	tnumbers["sixty"] = 60
	tnumbers["seventy"] = 70
	tnumbers["eighty"] = 80
	tnumbers["ninety"] = 90

	onumbers["hundred"] = 100
	onumbers["thousand"] = 1000
	onumbers["million"] = 1000000
	onumbers["billion"] = 1000000000

	numbers["first"] = 1
	numbers["second"] = 2
	numbers["third"] = 3
	numbers["fourth"] = 4
	numbers["fifth"] = 5
	numbers["sixth"] = 6
	numbers["seventh"] = 7
	numbers["eighth"] = 8
	numbers["nineth"] = 9
	numbers["tenth"] = 10
	numbers["eleventh"] = 11
	numbers["twelveth"] = 12
	numbers["thirteenth"] = 13
	numbers["fourteenth"] = 14
	numbers["fifteenth"] = 15
	numbers["sixteenth"] = 16
	numbers["seventeenth"] = 17
	numbers["eighteenth"] = 18
	numbers["nineteenth"] = 19

	tnumbers["twenteth"] = 20
	tnumbers["twentyfirst"] = 21
	tnumbers["twentysecond"] = 22
	tnumbers["twentythird"] = 23
	tnumbers["twentyfourth"] = 24
	tnumbers["twentyfifth"] = 25
	tnumbers["twentysixth"] = 26
	tnumbers["twentyseventh"] = 27
	tnumbers["twentyeight"] = 28
	tnumbers["twentynineth"] = 29
	tnumbers["thirteth"] = 30
	tnumbers["thirtyfirst"] = 31

	splitted := strings.Split(strings.ToLower(word), " ")

	for _, split := range splitted {
		if numbers[split] != 0 {
			temp = numbers[split]
			sum = sum + temp
			previous = previous + temp
		} else if onumbers[split] != 0 {
			if sum != 0 {
				sum = sum - previous
			}
			sum = sum + previous*onumbers[split]
			temp = 0
			previous = 0
		} else if tnumbers[split] != 0 {
			temp = tnumbers[split]
			sum = sum + temp
		}
	}

	if sum == 0 {
		return 0, errors.New("couldn't format number")
	}

	return sum, nil
}

func (p *Plugin) weekDayNumber(day string) int {
	switch day {
	case "sunday":
		return 0
	case "monday":
		return 1
	case "tuesday":
		return 2
	case "wednesday":
		return 3
	case "thursday":
		return 4
	case "friday":
		return 5
	case "saturday":
		return 6
	default:
		return -1
	}
}

func (p *Plugin) regSplit(text string, delimeter string) []string {

	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:]
	return result
}

func (p *Plugin) chooseClosest(chosen *time.Time, dayInterval bool) time.Time {
	loc, _ := time.LoadLocation("Europe/Warsaw")

	if dayInterval &&  chosen.Before(time.Now().In(loc)) {
			return chosen.Round(time.Second).Add(time.Hour * 24 * time.Duration(1))
	}
	if dayInterval {
		return *chosen
	}

	if !chosen.Before(time.Now().In(loc)) {
		return *chosen
	}

	if chosen.Add(time.Hour * 12 * time.Duration(1)).Before(time.Now().In(loc)) {
		return chosen.Round(time.Second).Add(time.Hour * 24 * time.Duration(1))
	}

	return chosen.Round(time.Second).Add(time.Hour * 12 * time.Duration(1))
}
