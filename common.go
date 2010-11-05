package main

import (
	"strconv"
	"bytes"
	"container/vector"
)

type Byter interface {
	Byte() []byte
}

type AddPlaceArgs struct {
	Auth
	Name string
}

func (a *AddPlaceArgs) Byte() []byte {
	b := bytes.NewBufferString(a.Name)
	return b.Bytes()
}


type UIntArgs struct {
	Auth
	Num uint
}

func (a *UIntArgs) Byte() []byte {
	b := bytes.NewBufferString(strconv.Uitoa(a.Num))
	return b.Bytes()
}

type EmptyArgs struct {
	Auth
}

func (a *EmptyArgs) Byte() []byte {
	return make([]byte, 1)
}

type Person struct {
	CanDrive        bool
	Name            string
	NumSeats        uint
	NominationsLeft uint
}

func (p *Person) String() string {
	str := p.Name
	if p.CanDrive {
		str += " [" + strconv.Uitoa(p.NumSeats) + " seats]"
	}
	return str
}

type Place struct {
	Id        uint
	Name      string
	Votes     uint
	People    *vector.Vector
	Nominator *Person
}

type Auth struct {
	Name                        string
	Mac, CChallenge, SChallenge []byte
}

func (p *Place) String() string {
	str := strconv.Uitoa(p.Id) + ") " + p.Name + " : " + p.Nominator.Name + " [" + strconv.Uitoa(p.Votes) + " votes]"
	for _, person := range p.People {
		str += "\n  - " + person.String()
	}
	return str
}

func (p *Place) removePerson(name string) bool {
	for i, e := range p.People {
		person, ok := e.(*Person)
		if ok && person.Name == name {
			p.People.Delete(i)
			return true
		}	
	}
	return false
}
