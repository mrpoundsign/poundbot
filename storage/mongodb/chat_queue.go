package mongodb

import (
	"log"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/types"
)

type ChatQueue struct {
	collection *mgo.Collection
}

func (cq ChatQueue) InsertMessage(m types.ChatMessage) error {
	return cq.collection.Insert(m)
}

func (cq ChatQueue) GetGameServerMessage(sk string, to time.Duration) (types.ChatMessage, bool) {
	sess := cq.collection.Database.Session.Copy()
	defer sess.Close()

	iter := cq.collection.With(sess).Find(
		bson.M{
			"serverkey":  sk,
			"senttouser": false,
		},
	).Tail(to)
	defer iter.Close()

	var cm types.ChatMessage
	for iter.Next(&cm) {
		err := cq.collection.Update(bson.M{"_id": cm.ID, "senttouser": false}, bson.M{"$set": bson.M{"senttouser": true}})
		if err != nil {
			if err != mgo.ErrNotFound {
				log.Printf("MongoDB Error updating message: %v", err)
			}
			return cm, false
		}
		return cm, true
	}

	if iter.Err() != nil {
		log.Printf("MongoDB: error getting message: %v", iter.Err())
		return cm, false
	}

	return cm, false
}