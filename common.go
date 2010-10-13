package main

import (
	"strconv"
	"bytes"
	"encoding/base64"
	"os"
	"json"
	"fmt"
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

func (p *Person) UnmarshalJSON(data []byte) os.Error {
	person := make(map[string]interface{})
	err := json.Unmarshal(data, &person)
	fmt.Println("JSON:", person)
	return err
}

type Place struct {
	Id        uint
	Name      string
	Votes     uint
	People    []*Person
	Nominator *Person
}

func (p *Place) UnmarshalJSON(data []byte) os.Error {
	place := make(map[string]interface{})
	err := json.Unmarshal(data, &place)
	fmt.Println("JSON:", place)
	return err
}

type Bin []byte

func (b Bin) MarshalJSON() ([]byte, os.Error) {
	encoded := make([]byte, 2+base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(encoded[1:], b)
	encoded[0] = '"'
	encoded[len(encoded)-1] = '"'
	return encoded, nil
}

func (b *Bin) UnmarshalJSON(val []byte) os.Error {
	data := val[1 : len(val)-1]
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(decoded, data)
	if err != nil {
		return err
	}
	*b = decoded[0:n]
	return nil
}

type Auth struct {
	Name                        string
	Mac, CChallenge, SChallenge *Bin
}

func (p *Place) String() string {
	str := strconv.Uitoa(p.Id) + ") " + p.Name + " nominated by " + p.Nominator.Name + " [" + strconv.Uitoa(p.Votes) + " votes]"
	for _, person := range p.People {
		str += "\n  - " + person.String()
	}
	return str
}
