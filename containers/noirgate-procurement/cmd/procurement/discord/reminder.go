// functions for working with shell expiration reminders
package discord

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) saveReminder(principal string, alert *time.Timer) {
	if _, ok := h.reminders[principal]; ok {
		return
	}
	h.reminders[principal] = alert
}

func (h *Handler) removeReminder(principal string) {
	if alert, ok := h.reminders[principal]; ok {
		alert.Stop()
		delete(h.reminders, principal)
	}
}

func (h *Handler) scheduleReminder(s *discordgo.Session, i *discordgo.InteractionCreate, after time.Duration) {
	userId := i.Member.User.ID
	alert := time.AfterFunc(after, func() {
		defer h.removeReminder(userId)
		dm, err := s.UserChannelCreate(userId)
		if err != nil {
			h.logger.Sugar().Errorf("Error getting DM channel, %s", err.Error())
			return
		}
		_, err = s.ChannelMessageSend(dm.ID, "Your shell will expire soon.")
		if err != nil {
			h.logger.Sugar().Errorf("Error sending warning DM, %s", err.Error())
			return
		}
	})
	h.saveReminder(userId, alert)
}
