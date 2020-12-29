package raid

import (
	"time"

	"github.com/poundbot/poundbot/pkg/models"
)

type notifier interface {
	RaidNotify(models.RaidAlertWithMessageChannel)
}

// A raidStore stores raid information
type raidStore interface {
	GetReady() ([]models.RaidAlert, error)
	IncrementNotifyCount(models.RaidAlert) error
	Remove(models.RaidAlert) error
	messageIDSetter
}

type messageIDSetter interface {
	SetMessageID(models.RaidAlert, string) error
}

// A RaidAlerter sends notifications on raids
type alerter struct {
	rs        raidStore
	rn        notifier
	SleepTime time.Duration
	done      <-chan interface{}
	miu       func(ra models.RaidAlertWithMessageChannel, is messageIDSetter)
}

// NewAlerter constructs an Alerter
func NewAlerter(ral raidStore, rn notifier, done <-chan interface{}) *alerter {
	return &alerter{
		rs:        ral,
		rn:        rn,
		done:      done,
		SleepTime: 1 * time.Second,
		miu:       messageIDUpdate,
	}
}

func messageIDUpdate(ra models.RaidAlertWithMessageChannel, is messageIDSetter) {
	raLog := log.WithField("sys", "RALERT")
	newMessageID, ok := <-ra.MessageIDChannel
	if !ok {
		raLog.Trace("messageID channel close")
	}
	raLog.Tracef("New message ID is %s", newMessageID)
	if newMessageID != ra.MessageID {
		err := is.SetMessageID(ra.RaidAlert, newMessageID)
		if err != nil {
			raLog.WithError(err).Error("storage: Could not set message ID")
		}
	}
}

// Run checks for raids that need to be alerted and sends them
// out through the RaidNotify channel. It runs in a loop.
func (r *alerter) Run() {
	log.Info("Starting")
	for {
		select {
		case <-r.done:
			log.Warn("Shutting down")
			return
		case <-time.After(r.SleepTime):
			alerts, err := r.rs.GetReady()
			if err != nil {
				log.WithError(err).Error("could not get raid alert")
				continue
			}

			for _, alert := range alerts {
				shouldNotify := true
				log.Tracef("Processing alert %s, %d", alert.ID, alert.NotifyCount)

				// Increment notify count should ensure we're the node that should notify for this action.
				if err := r.rs.IncrementNotifyCount(alert); err != nil {
					log.WithError(err).Trace("could not increment")
					shouldNotify = false
				}

				if alert.ValidUntil.Before(time.Now()) {
					log.Trace("removing")
					if err := r.rs.Remove(alert); err != nil {
						log.Trace("could not remove")
						log.WithError(err).Error("storage: Could not remove alert")
						continue
					}
				}

				if shouldNotify {
					message := models.RaidAlertWithMessageChannel{
						RaidAlert:        alert,
						MessageIDChannel: make(chan string),
					}
					log.Trace("notifying")
					r.rn.RaidNotify(message)
					go r.miu(message, r.rs)
					if err := r.rs.Remove(alert); err != nil {
						log.Trace("could not remove")
						log.WithError(err).Error("storage: Could not remove alert")
						continue
					}
				}
			}
		}
	}
}
