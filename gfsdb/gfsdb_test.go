package gfsdb

import (
	"github.com/Centny/dbm/mgo"
	"testing"
)

func init() {
	mgo.C(CN_F).RemoveAll(nil)
	mgo.C(CN_FILE).RemoveAll(nil)
	C = mgo.C
}

func TestCreateOutPath(t *testing.T) {
	CreateOutPath(&F{
		Path: "",
	})
}
