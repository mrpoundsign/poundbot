package rustconn

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	SourceRust    = "rust"
	SourceDiscord = "discord"
)

type MongoConfig struct {
	DialAddress string
	Database    string
}

type ServerConfig struct {
	BindAddr    string
	Port        int
	MongoConfig MongoConfig
}

type Server struct {
	sc              *ServerConfig
	mongoDB         *mgo.Database
	userColl        *mgo.Collection
	discordAuthColl *mgo.Collection
	raidAlertColl   *mgo.Collection
	chatCollection  *mgo.Collection
	RaidNotify      chan RaidNotification
	DiscordAuth     chan DiscordAuth
	AuthSuccess     chan DiscordAuth
	ChatChan        chan string
	ChatOutChan     chan ChatMessage
}

func NewServer(sc *ServerConfig, rch chan RaidNotification, dach, asch chan DiscordAuth, cch chan string, coch chan ChatMessage) *Server {
	return &Server{
		sc:          sc,
		RaidNotify:  rch,
		DiscordAuth: dach,
		AuthSuccess: asch,
		ChatChan:    cch,
		ChatOutChan: coch,
	}
}

func (s *Server) Serve() {
	session, err := mgo.Dial(s.sc.MongoConfig.DialAddress)
	if err != nil {
		log.Fatal(err)
	}
	s.mongoDB = session.DB(s.sc.MongoConfig.Database)
	s.userColl = s.mongoDB.C("users")
	s.discordAuthColl = s.mongoDB.C("discord_auths")
	s.raidAlertColl = s.mongoDB.C("raid_alerts")
	s.chatCollection = s.mongoDB.C("chats")

	index := mgo.Index{
		Key:        []string{"steam_id"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	s.userColl.EnsureIndex(index)
	s.discordAuthColl.EnsureIndex(index)

	// u := User{DiscordInfo: DiscordInfo{DiscordID: "MrPoundsign#2364"}, SteamInfo: SteamInfo{SteamID: 76561197960794006}, CreatedAt: time.Now().UTC()}
	// s.userColl.Upsert(u.UpsertID(), u)
	go s.authHandler()
	go s.raidAlerter()
	fmt.Printf("Starting HTTP Server on %s:%d\n", s.sc.BindAddr, s.sc.Port)
	http.HandleFunc("/entity_death", s.entityDeathHandler)
	http.HandleFunc("/discord_auth", s.discordAuthHandler)
	http.HandleFunc("/chat", s.chatHandler)
	go log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", s.sc.BindAddr, s.sc.Port), nil))
}

func (s *Server) raidAlerter() {
	for {
		time.Sleep(10)
		var results []RaidNotification
		err := s.raidAlertColl.Find(bson.M{"alert_at": bson.M{"$lte": time.Now().UTC()}}).All(&results)
		if err != nil {
			log.Printf("Error getting raid notifications! %s\n", err)
		}
		if len(results) > 0 {
			for _, result := range results {
				s.RaidNotify <- result
				s.raidAlertColl.Remove(result.DiscordInfo)
			}
		}
	}
}

func (s *Server) authHandler() {
	for {
		as := <-s.AuthSuccess
		u := User{DiscordInfo: as.DiscordInfo, SteamInfo: as.SteamInfo, CreatedAt: time.Now().UTC()}
		_, err := s.userColl.Upsert(u.UpsertID(), u)
		if err == nil {
			s.discordAuthColl.Remove(as.SteamInfo)
			if as.Ack != nil {
				as.Ack(true)
			}
		} else {
			if as.Ack != nil {
				as.Ack(false)
			}
		}
	}
}

func (s *Server) chatHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		decoder := json.NewDecoder(r.Body)
		var t ChatMessage
		err := decoder.Decode(&t)
		if err != nil {
			log.Println(err)
			return
		}

		t.Source = SourceRust
		s.chatCollection.Insert(t)
		go func(t ChatMessage, c chan string) {
			c <- fmt.Sprintf("☢️ **%s**: %s", t.Username, t.Message)
		}(t, s.ChatChan)
	case http.MethodGet:
		select {
		case res := <-s.ChatOutChan:
			b, err := json.Marshal(res)
			if err != nil {
				fmt.Println(err)
				return
			}
			s.chatCollection.Insert(res)

			w.Write(b)
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusNoContent)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Method %s not allowed\"}", r.Method)))
	}
}

func (s *Server) entityDeathHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t EntityDeath
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(err)
		return
	}
	for _, steamID := range t.Owners {
		u, err := s.findUser(steamID)
		if err == nil {
			s.raidAlertColl.Upsert(
				u.DiscordInfo,
				bson.M{
					"$setOnInsert": bson.M{
						"alert_at": time.Now().UTC().Add(5 * time.Minute),
					},
					"$inc": bson.M{
						fmt.Sprintf("items.%s", t.Name): 1,
					},
					"$addToSet": bson.M{
						"grid_positions": t.GridPos,
					},
				},
			)
		}
	}
}

func (s *Server) discordAuthHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t DiscordAuth
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("User Auth Request: %v from %v\n", t, r.Body)
	u, err := s.findUser(t.SteamID)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s is linked to you.\"}", u.DiscordID)))
		return
	} else if t.DiscordID == "check" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"Account is not linked to discord.\"}"))
		return
	}

	_, err = s.discordAuthColl.Upsert(t.UpsertID(), t)
	if err == nil {
		s.DiscordAuth <- t
	} else {
		log.Println(err)
	}
}

func (s *Server) findUser(steamid uint64) (u *User, err error) {
	err = s.userColl.Find(SteamInfo{SteamID: steamid}).One(&u)
	if err != nil {
		return nil, err
	}
	return
}

func (s *Server) removeUser(steamid uint64) (err error) {
	err = s.userColl.Remove(SteamInfo{SteamID: steamid})
	return
}