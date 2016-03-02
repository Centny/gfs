package gfsdb

import (
	"github.com/Centny/gwf/util"
	tmgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func FOI_File(file *File) (int, error) {
	file.Id = bson.NewObjectId().Hex()
	if len(file.Fid) < 1 || len(file.Oid) < 1 || len(file.Owner) < 1 {
		return 0, util.Err("the fid/oid/owner must be setted")
	}
	var res, err = C(CN_FILE).Find(bson.M{
		"fid":   file.Fid,
		"oid":   file.Oid,
		"owner": file.Owner,
	}).Apply(tmgo.Change{
		Update: bson.M{
			"$setOnInsert": file,
		},
		Upsert:    true,
		ReturnNew: true,
	}, file)
	var updated int = 0
	if err == nil && res.UpsertedId != nil {
		updated = 1
	}
	return updated, err
}

func CountFile() (int, error) {
	return C(CN_FILE).Count()
}
