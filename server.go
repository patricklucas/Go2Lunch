package main

import (
	"rpc"
	"log"
	"http"
	"net"
	"os"
	"crypto/hmac"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/rand"
	"io/ioutil"
	"sync"
	"json"
	"flag"
	"strconv"
)

var port = flag.Uint("p", 1234, "Specifies the port to listen on.")
var configFile = flag.String("c", "config.json", "Specify a config file.")
var displayHelp = flag.Bool("help", false, "Displays this help message.")

type ServerConfig struct {
	Sekritz map[string]string
}

var userMap map[string]*Auth
var config *ServerConfig
var cMutex sync.Mutex

func loadUsersFromFile() (err os.Error) {
	tempConfig := &ServerConfig{}
	read, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(read, tempConfig)
	if err != nil {
		return err
	}

	cMutex.Lock()
	config = tempConfig
	cMutex.Unlock()
	return nil
}

func Challenge(name *string, challenge *[]byte) os.Error {
	_, valid := userMap[(*name)]
	if !valid {
		valid = checkUser(*name)
		if !valid {
			return nil
		}
	}

	*challenge = make([]byte, 512)
	n, err := rand.Read(*challenge)

	if err != nil || n != 512 {
		panic("Challenge Generation Failed")
	}

	userMap[*name].SChallenge = *challenge
	return nil
}

func checkUser(name string) bool {
	loadUsersFromFile()
	cMutex.Lock()
	_, valid := config.Sekritz[name]
	cMutex.Unlock()
	if valid {
		userMap[name] = &Auth{Name: name, SChallenge: make([]byte, 512)}
	}

	return valid
}


func verify(a *Auth, d Byter) (bool, os.Error) {
	cMutex.Lock()
	key, ok := config.Sekritz[(*a).Name]
	cMutex.Unlock()

	if !ok {
		return false, os.NewError("Unknown User")
	}

	mac := hmac.New(sha512.New, []byte(key))

	mac.Write([]byte((*a).Name))
	mac.Write((*a).CChallenge)
	mac.Write(d.Byte())
	mac.Write(userMap[(*a).Name].SChallenge)
	if subtle.ConstantTimeCompare(mac.Sum(), (*a).Mac) == 1 {
		return true, nil
	}
	return false, os.NewError("Authentication Failed")
}


func main() {
	flag.Parse()

	if *displayHelp {
		flag.PrintDefaults()
		return
	}
	userMap = make(map[string]*Auth)
	err := loadUsersFromFile()
	if err != nil {
		log.Exit("Error reading config file. Have you created it?\nCoused By: ", err)
	}

	t := &LunchTracker{newPollChan()}
	rpc.Register(t)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+strconv.Uitoa(*port))
	if e != nil {
		log.Exit("listen error:", e)
	}
	http.Serve(l, nil)
}

func newPollChan() chan *LunchPoll {
	ch := make(chan *LunchPoll)
	poll := NewPoll()
	go func() {
		for {
			ch <- poll
			poll <- ch
		}
	}()
}
