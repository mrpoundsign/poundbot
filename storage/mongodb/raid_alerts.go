package mongodb

import (
	"errors"
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/pkg/models"
)

type userGetter interface {
	GetByPlayerID(models.PlayerID) (models.User, error)
}

// A RaidAlerts implements storage.RaidAlertsStore
type RaidAlerts struct {
	collection *mgo.Collection
	users      userGetter
}

// AddInfo implements storage.RaidAlertsStore.AddInfo
func (r RaidAlerts) AddInfo(alertIn, invalidIn time.Duration, ed models.EntityDeath) error {
	for _, pid := range ed.OwnerIDs {
		// Checking if the user exists, just bail if not
		_, err := r.users.GetByPlayerID(models.PlayerID(pid))
		if err != nil {
			continue
		}

		validUntil := time.Now().UTC().Add(invalidIn)

		_, err = r.collection.Upsert(
			bson.M{
				"playerid":   pid,
				"serverkey":  ed.ServerKey,
				"validuntil": bson.M{"$gt": time.Now().UTC()},
			},
			bson.M{
				"$setOnInsert": bson.M{
					"alertat":     time.Now().UTC().Add(alertIn),
					"servername":  ed.ServerName,
					"serverkey":   ed.ServerKey,
					"notifycount": 0,
				},
				"$set": bson.M{
					"validuntil": validUntil,
				},
				"$inc": bson.M{
					fmt.Sprintf("items.%s", ed.Name): 1,
				},
				"$addToSet": bson.M{
					"gridpositions": ed.GridPos,
				},
			},
		)

		if err != nil {
			log.WithError(err).Error("could not add info")
		}
	}
	return nil
}

// GetReady implements storage.RaidAlertsStore.GetReady
func (r RaidAlerts) GetReady() ([]models.RaidAlert, error) {
	var alerts []models.RaidAlert

	err := r.collection.Find(
		bson.M{
			"alertat": bson.M{
				"$lte": time.Now().UTC(),
			},
		},
	).All(&alerts)

	if err == mgo.ErrNotFound {
		err = nil
	}
	return alerts, err
}

// Remove implements storage.RaidAlertsStore.Remove
func (r RaidAlerts) Remove(alert models.RaidAlert) error {
	return r.collection.Remove(bson.M{"_id": alert.ID})
}

func (r RaidAlerts) IncrementNotifyCount(ra models.RaidAlert) error {
	icount := ra.ItemCount()

	if icount == ra.NotifyCount {
		return errors.New("notification count does not need to be incremented")
	}

	return r.collection.Update(
		bson.M{
			"_id":         ra.ID,
			"notifycount": ra.NotifyCount,
		},
		bson.M{
			"$set": bson.M{
				"notifycount": icount,
			},
		},
	)
}

func (r RaidAlerts) SetMessageID(ra models.RaidAlert, messageID string) error {
	return r.collection.Update(
		bson.M{"_id": ra.ID},
		bson.M{
			"$set": bson.M{
				"messageid": messageID,
			},
		},
	)
}
