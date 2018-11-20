package main

import (
	"fmt"
	"time"
	"strings"
	"encoding/json"

	"github.com/mattermost/mattermost-server/model"
)

type Reminder struct {
	TeamId string

	Id string

	Username string

	Target string

	Message string

	When string

	Occurrences []Occurrence

	Completed time.Time
}

type ReminderRequest struct {
	TeamId string

	Username string

	Payload string

	Reminder Reminder
}

func (p *Plugin) GetReminders(username string) ([]Reminder) {

	user, uErr := p.API.GetUserByUsername(username)

	if uErr != nil {
		p.API.LogError("failed to query user %s", username)
		return []Reminder{}
	}

	bytes, bErr := p.API.KVGet(user.Username)
	if bErr != nil {
		p.API.LogError("failed KVGet %s", bErr)
		return []Reminder{}
	}

	var reminders []Reminder
	err := json.Unmarshal(bytes, &reminders)

	if err != nil {
		p.API.LogError("new reminder " + user.Username)
	} else {
		p.API.LogDebug("existing " + fmt.Sprintf("%v", reminders))
	}

	return reminders
}

func (p *Plugin) UpsertReminder(request ReminderRequest) {

	user, uErr := p.API.GetUserByUsername(request.Username)

	if uErr != nil {
		p.API.LogError("failed to query user %s", request.Username)
		return
	}

	bytes, bErr := p.API.KVGet(user.Username)
	if bErr != nil {
		p.API.LogError("failed KVGet %s", bErr)
		return
	}

	var reminders []Reminder
	err := json.Unmarshal(bytes, &reminders)

	if err != nil {
		p.API.LogDebug("new reminder " + user.Username)
	} else {
		p.API.LogDebug("existing " + fmt.Sprintf("%v", reminders))
	}

	reminders = append(reminders, request.Reminder)
	ro, rErr := json.Marshal(reminders)

	if rErr != nil {
		p.API.LogError("failed to marshal reminders %s", user.Username)
		return
	}

	p.API.KVSet(user.Username, ro)
}

func (p *Plugin) TriggerReminders() {

	bytes, err := p.API.KVGet(string(fmt.Sprintf("%v", time.Now().Round(time.Second))))

	p.API.LogDebug("*")

	if err != nil {
		p.API.LogError("failed KVGet %s", err)
		return
	}
	if string(bytes[:]) == "" {
		return
	}

	p.API.LogDebug(string(bytes[:]))

	var reminderOccurrences []Occurrence
	roErr := json.Unmarshal(bytes, &reminderOccurrences)
	if roErr != nil {
		p.API.LogError("Failed to unmarshal reminder occurrences " + fmt.Sprintf("%v", roErr))
		return
	}

	p.API.LogDebug("Trigger", fmt.Sprintf("%v", reminderOccurrences))

	for _, ReminderOccurrence := range reminderOccurrences {

		user, err := p.API.GetUserByUsername(ReminderOccurrence.Username)

		if err != nil {
			p.API.LogError("failed to query user %s", user.Id)
			continue
		}

		bytes, b_err := p.API.KVGet(user.Username)
		if b_err != nil {
			p.API.LogError("failed KVGet %s", b_err)
			return
		}

		var reminders []Reminder
		uerr := json.Unmarshal(bytes, &reminders)

		if uerr != nil {
			continue
		}

		reminder := p.findReminder(reminders, ReminderOccurrence)

		p.API.LogDebug(fmt.Sprintf("%v", reminder))

		if strings.HasPrefix(reminder.Target, "@") || strings.HasPrefix(reminder.Target, "me") {
			p.triggerToUser(user, reminder)
			continue
		}
		// channel
		p.triggerToChannel(user, reminder)
	}
}

func (p *Plugin) getUser(user *model.User, reminder Reminder) *model.User {
	p.API.LogDebug("find user 1", user.Id, reminder.Target)
	if !strings.HasPrefix(reminder.Target, "@") {
		return user
	}

	targetUsername := strings.TrimLeft(reminder.Target, "@")
	targetUser, err := p.API.GetUserByUsername(targetUsername)
	p.API.LogDebug("find user 2", targetUsername, targetUser.Id)

	if err == nil {
		return targetUser
	}

	p.API.LogError("failed to query target user", targetUsername, err.Error())
	return user
}

func (p *Plugin) triggerToUser(user *model.User, reminder Reminder) {
	targetUser := p.getUser(user, reminder)

	p.API.LogDebug("DM: "+fmt.Sprintf("%v", p.remindUserId) + "__" + fmt.Sprintf("%v", targetUser.Id))
	channel, cErr := p.API.GetDirectChannel(p.remindUserId, targetUser.Id)

	if cErr != nil {
		p.API.LogError("fail to get direct channel ", fmt.Sprintf("%v", cErr))
		return
	}
	p.API.LogDebug("got direct channel " + fmt.Sprintf("%v", channel))

	var finalTarget string
	finalTarget = reminder.Target
	if finalTarget == "me" {
		finalTarget = "You"
	} else {
		finalTarget = "@" + user.Username
	}

	_, err := p.API.CreatePost(&model.Post{
		UserId:    p.remindUserId,
		ChannelId: channel.Id,
		Message:   fmt.Sprintf(finalTarget + " asked me to remind you \"" + reminder.Message + "\"."),
	})
	if err == nil {
		return
	}
	p.API.LogError(
		"failed to post DM message",
		"user_id", user.Id,
		"error", err.Error(),
	)
}

func (p *Plugin) triggerToChannel(user *model.User, reminder Reminder) {
	//channel, cErr := p.API.GetChannelByName(reminder.TeamId, strings.Replace(reminder.Target, "~", "", -1), false)
	channel, cErr := p.API.GetChannelByName(reminder.TeamId, strings.Replace(reminder.Target, "~", "", -1))

	if cErr != nil {
		p.API.LogError("fail to get channel " + fmt.Sprintf("%v", cErr))
		return
	}

		p.API.LogDebug("got channel " + fmt.Sprintf("%v", channel))

	_, err := p.API.CreatePost(&model.Post{
		UserId:    p.remindUserId,
		ChannelId: channel.Id,
		Message:   fmt.Sprintf("@" + user.Username + " asked me to remind you \"" + reminder.Message + "\"."),
	})
	if err == nil {
		return
	}

	p.API.LogError(
		"failed to post DM message",
		"user_id", user.Id,
		"error", err.Error(),
	)
}


func (p *Plugin) findReminder(reminders []Reminder, reminderOccurrence Occurrence) (Reminder) {
	for _, reminder := range reminders {
		if reminder.Id == reminderOccurrence.ReminderId {
			return reminder
		}
	}
	return Reminder{}
}
