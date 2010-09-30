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

var port = flag.Uint("p", 1234, "port")

type ServerConfig struct {
	Sekritz map[string]string
}

type LunchTracker struct {
	*LunchPoll
}

var userMap map[string]*Auth
var config *ServerConfig
var cMutex sync.Mutex

func init() {
	userMap = make(map[string]*Auth)
	err := loadUsersFromFile()
	if err != nil {
		panic(err)
	}
}

func loadUsersFromFile() (err os.Error) {
	tempConfig := &ServerConfig{}
	read, err := ioutil.ReadFile("config.json")
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

func (t *LunchTracker) AddPlace(args *AddPlaceArgs, place *uint) os.Error {
	valid, ive := verify(&args.Auth, args)
	if !valid {
		return ive
	}
	*place = t.LunchPoll.addPlace(args.Name, args.Auth.Name)
	return nil
}

func (t *LunchTracker) DelPlace(args *UIntArgs, success *bool) os.Error {
	valid, ive := verify(&args.Auth, args)
	if !valid {
		return ive
	}
	*success = t.LunchPoll.delPlace(args.Num)
	return nil
}

func (t *LunchTracker) Drive(args *UIntArgs, success *bool) os.Error {
	valid, ive := verify(&args.Auth, args)
	if !valid {
		return ive
	}
	*success = t.LunchPoll.drive(args.Auth.Name, args.Num)
	return nil
}

func (t *LunchTracker) UnDrive(args *EmptyArgs, success *bool) os.Error {
	valid, ive := verify(&args.Auth, args)
	if !valid {
		return ive
	}
	*success = t.LunchPoll.unDrive(args.Auth.Name)
	return nil
}

func (t *LunchTracker) Vote(args *UIntArgs, success *bool) os.Error {
	valid, ive := verify(&args.Auth, args)
	if !valid {
		return ive
	}

	*success = t.LunchPoll.vote(args.Auth.Name, args.Num)
	return nil
}

func (t *LunchTracker) UnVote(args *EmptyArgs, success *bool) os.Error {
	valid, ive := verify(&args.Auth, args)
	if !valid {
		return ive
	}
	*success = t.LunchPoll.unVote(args.Auth.Name)
	return nil
}

func (t *LunchTracker) DisplayPlaces(args *EmptyArgs, response *[]Place) os.Error {
	valid, ive := verify(&args.Auth, args)
	if !valid {
		return ive
	}
	*response = t.LunchPoll.displayPlaces()
	return nil
}


func (t *LunchTracker) Challenge(name *string, challenge *[]byte) os.Error {
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
	mac := hmac.New(sha512.New, []byte(config.Sekritz[(*a).Name]))
	cMutex.Unlock()

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
	t := &LunchTracker{NewPoll()}
	rpc.Register(t)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+strconv.Uitoa(*port))
	if e != nil {
		log.Exit("listen error:", e)
	}
	http.Serve(l, nil)
}
