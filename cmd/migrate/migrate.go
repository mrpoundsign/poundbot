package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	migrate "github.com/eminetto/mongo-migrate"
	"github.com/globalsign/mgo"
	_ "github.com/poundbot/poundbot/migrations"
	"github.com/spf13/viper"
)

var (
	configFile = flag.String("c", "", "The config file")
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("Missing options: up or down")
	}
	option := os.Args[1]

	viper.SetDefault("mongo.dial", "mongodb://localhost:27017")
	viper.SetDefault("mongo.database", "poundbot")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if len(*configFile) == 0 {
		log.Info("using default config locations")
		viper.SetConfigName("poundbot")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/poundbot/")
		viper.AddConfigPath("$HOME/.poundbot/")
	} else {
		viper.SetConfigFile(*configFile)
	}

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Errorf("Error reading config file: %s,%s", reflect.TypeOf(err), err)
			os.Exit(1)
		}
		log.Info("No config file found. Using defaults and env only.")
	}

	var sErr error
	var sess *mgo.Session
	if viper.GetBool("mongo.ssl.enabled") {
		dialInfo, err := mgo.ParseURL(viper.GetString("mongo.dial-addr"))
		if err != nil {
			log.Println(err)
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: viper.GetBool("mongo.ssl.insecure"),
			}
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			if err != nil {
				log.Println(err)
			}
			return conn, err
		}
		sess, sErr = mgo.DialWithInfo(dialInfo)
	} else {
		sess, sErr = mgo.Dial(viper.GetString("mongo.dial-addr"))
	}
	if sErr != nil {
		panic(sErr)
	}

	// session, err := mgo.Dial(viper.GetString("mongo.dial-addr"))
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }
	// defer session.Close()
	// db := session.DB(viper.GetString("mongo.database"))
	migrate.SetDatabase(sess.DB(viper.GetString("mongo.database")))
	migrate.SetMigrationsCollection("migrations")
	// migrate.SetLogger(log.Info)
	switch option {
	case "new":
		if len(os.Args) != 3 {
			log.Fatal("Should be: new description-of-migration")
		}
		fName := fmt.Sprintf("./migrations/%s_%s.go", time.Now().Format("20060102150405"), os.Args[2])
		from, err := os.Open("./migrations/template.go")
		if err != nil {
			log.Fatal("Should be: new description-of-migration")
		}
		defer from.Close()

		to, err := os.OpenFile(fName, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err.Error())
		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("New migration created: %s\n", fName)
	case "up":
		err = migrate.Up(migrate.AllAvailable)
	case "down":
		err = migrate.Down(migrate.AllAvailable)
	}
	if err != nil {
		log.Fatal(err.Error())
	}
}
