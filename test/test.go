package test

import (
	"github.com/Centny/dbm/mgo"
)

func init() {
	mgo.AddDefault("cny:123@loc.m:27017/cny", "cny")
	mgo.C("c_f").RemoveAll(nil)
	mgo.C("c_file").RemoveAll(nil)
}
