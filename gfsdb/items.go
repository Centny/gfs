package gfsdb

import (
	"github.com/Centny/gwf/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	ES_RUNNING = "running"
	ES_DONE    = "done"
	ES_ERROR   = "error"
	ES_IGNORE  = "ignore"
	ES_NONE    = "none"

	//
	VS_VERIFIED = "verified"
	VS_ZERO     = "zero"
	VS_REDO     = "redo"
	VS_ERROR    = "error"
)

const (
	FS_N = "N" //normal
	FS_D = "D" //deleted
)
const (
	FT_FILE   = "file"
	FT_FOLDER = "folder"
)

type F struct {
	Id       string   `bson:"_id" json:"id"`
	Name     string   `bson:"name" json:"name"`
	Filename string   `bson:"filename" json:"filename"` //upload file name
	Pub      string   `bson:"pub" json:"pub"`           //public path.
	SHA      string   `bson:"sha" json:"sha"`           //file sha
	MD5      string   `bson:"md5" json:"md5"`           //file md5
	EXT      string   `bson:"ext" json:"ext"`           //file externd
	Size     int64    `bson:"size" json:"size"`         //file size.
	Type     string   `bson:"type" json:"type"`         //mimetype
	Path     string   `bson:"path" json:"-"`            //file save path.
	Exec     string   `bson:"exec" json:"exec"`         //the exec status
	Info     util.Map `bson:"info" json:"info"`         //the extern info.
	Status   string   `bson:"status" json:"status"`     //file status
	Time     int64    `bson:"time" json:"time"`         //upload time.
}

func (f *F) ToBsonM() bson.M {
	return bson.M{
		"_id":      f.Id,
		"name":     f.Name,
		"filename": f.Filename,
		"pub":      f.Pub,
		"sha":      f.SHA,
		"md5":      f.MD5,
		"ext":      f.EXT,
		"size":     f.Size,
		"type":     f.Type,
		"path":     f.Path,
		"exec":     f.Exec,
		"info":     f.Info,
		"status":   f.Status,
		"time":     f.Time,
	}
}

type Mark struct {
	Id  string `bson:"_id" json:"id"`
	Fid string `bson:"fid" json:"fid"`
}

// func (f *F) AddMark(mark []string) []string {
// 	var ms = map[string]int{}
// 	for _, v := range f.Mark {
// 		ms[v] = 1
// 	}
// 	var added = map[string]int{}
// 	var news = []string{}
// 	for _, v := range mark {
// 		if _, ok := ms[v]; ok {
// 			continue
// 		}
// 		if _, ok := added[v]; ok {
// 			continue
// 		}
// 		news = append(news, v)
// 		added[v] = 1
// 	}
// 	return news
// }

type File struct {
	Id     string   `bson:"_id" json:"id"`
	Fid    string   `bson:"fid" json:"fid"`
	Pid    string   `bson:"pid" json:"pid"`
	Oid    string   `bson:"oid" json:"oid"`
	Owner  string   `bson:"owner" json:"owner"`
	Name   string   `bson:"name" json:"name"`
	EXT    string   `bson:"ext" json:"ext"`   //file externd
	Type   string   `bson:"type" json:"type"` //type
	Tags   []string `bson:"tags" json:"tags"`
	Desc   string   `bson:"desc" json:"desc"`
	Status string   `bson:"status" json:"status"` //file status
	Time   int64    `bson:"time" json:"time"`     //upload time.
}

var Indexes = map[string]map[string]mgo.Index{
	CN_F: map[string]mgo.Index{
		"f_name": mgo.Index{
			Key: []string{"name"},
		},
		"f_filename": mgo.Index{
			Key: []string{"filename"},
		},
		"f_pub": mgo.Index{
			Key: []string{"pub"},
		},
		"f_sha": mgo.Index{
			Key: []string{"sha"},
		},
		"f_md5": mgo.Index{
			Key: []string{"md5"},
		},
		"f_ext": mgo.Index{
			Key: []string{"ext"},
		},
		"f_size": mgo.Index{
			Key: []string{"size"},
		},
		"f_type": mgo.Index{
			Key: []string{"type"},
		},
		"f_exec": mgo.Index{
			Key: []string{"exec"},
		},
		"f_status": mgo.Index{
			Key: []string{"status"},
		},
		"f_time": mgo.Index{
			Key: []string{"time"},
		},
	},
}
