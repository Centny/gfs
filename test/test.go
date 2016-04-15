package test

import (
	"github.com/Centny/dbm/mgo"
)

func init() {
	mgo.AddDefault("cny:123@loc.m:27017/cny", "cny")
}
