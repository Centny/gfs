package gfsdb

import (
	"github.com/Centny/gwf/util"
	tmgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func FOI_File(file *File) (int, error) {
	file.Id = bson.NewObjectId().Hex()
	if len(file.Type) < 1 {
		return 0, util.Err("the type must be setted")
	}
	var query bson.M
	if file.Type == FT_FILE {
		if len(file.Fid) < 1 || len(file.Oid) < 1 || len(file.Owner) < 1 {
			return 0, util.Err("the fid/oid/owner must be setted")
		}
		query = bson.M{
			"fid":   file.Fid,
			"oid":   file.Oid,
			"owner": file.Owner,
			"type":  file.Type,
		}
	} else {
		if len(file.Oid) < 1 || len(file.Owner) < 1 || len(file.Name) < 1 {
			return 0, util.Err("the oid/owner/name must be setted")
		}
		query = bson.M{
			"pid":   file.Pid,
			"oid":   file.Oid,
			"owner": file.Owner,
			"name":  file.Name,
			"type":  file.Type,
		}
	}
	var res, err = C(CN_FILE).Find(query).Apply(tmgo.Change{
		Update: bson.M{
			"$setOnInsert": file,
		},
		Upsert:    true,
		ReturnNew: true,
	}, file)
	var updated = 0
	if err == nil && res.UpsertedId != nil {
		updated = 1
	}
	return updated, err
}

func UpdateFile(file *File) error {
	var update = bson.M{}
	if len(file.Name) > 0 {
		update["name"] = file.Name
	}
	if len(file.Tags) > 0 {
		update["tags"] = file.Tags
	}
	if len(file.Desc) > 0 {
		update["desc"] = file.Desc
	}
	if len(file.Pid) > 0 {
		update["pid"] = file.Pid
	}
	update["time"] = util.Now()
	return C(CN_FILE).Update(bson.M{"_id": file.Id}, update)
}

func CountFile() (int, error) {
	return C(CN_FILE).Find(bson.M{"type": FT_FILE}).Count()
}

func FindFile(id string) (*File, error) {
	var file = &File{}
	var err = C(CN_FILE).FindId(id).One(&file)
	return file, err
}

func ListFile(oid, owner, name, typ string, pid, tags, status []string) ([]*File, error) {
	var query = bson.M{}
	if len(pid) > 0 {
		query["pid"] = bson.M{
			"$in": pid,
		}
	}
	if len(oid) > 0 {
		query["oid"] = oid
	}
	if len(owner) > 0 {
		query["owner"] = owner
	}
	if len(name) > 0 {
		query["name"] = bson.M{
			"$regex":   ".*" + name + ".*",
			"$options": "mi",
		}
	}
	if len(typ) > 0 {
		query["type"] = typ
	}
	if len(tags) > 0 {
		query["tags"] = bson.M{
			"$elemMatch": bson.M{
				"$in": tags,
			},
		}
	}
	if len(status) > 0 {
		query["status"] = bson.M{
			"$in": status,
		}
	}
	// fmt.Println(util.S2Json(query))
	var fs = []*File{}
	var err = C(CN_FILE).Find(query).All(&fs)
	return fs, err

}
