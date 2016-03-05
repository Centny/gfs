package gfsdb

import (
	"github.com/Centny/gwf/util"
)

const (
	ES_RUNNING = "running"
	ES_DONE    = "done"
	ES_ERROR   = "error"
	ES_NONE    = "none"
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
	Path     string   `bson:"path" json:"path"`         //file save path.
	Mark     []string `bson:"mark" json:"mark"`         //file marks
	Exec     string   `bson:"exec" json:"exec"`         //the exec status
	Info     util.Map `bson:"info" json:"info"`         //the extern info.
	Status   string   `bson:"status" json:"status"`     //file status
	Time     int64    `bson:"time" json:"time"`         //upload time.
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
	Oid    string   `bson:"oid" json:"oid"`
	Owner  string   `bson:"owner" json:"owner"`
	Name   string   `bson:"name" json:"name"`
	Tags   []string `bson:"tags" json:"tags"`
	Desc   string   `bson:"desc" json:"desc"`
	Folder string   `bson:"folder" json:"floder"`
	Status string   `bson:"status" json:"status"` //file status
	Time   int64    `bson:"time" json:"time"`     //upload time.
}

type Floder struct {
	Id     string   `bson:"_id" json:"id"`
	Oid    string   `bson:"oid" json:"oid"`
	Owner  string   `bson:"owner" json:"owner"`
	Pid    string   `bson:"pid" json:"pid"`
	Name   string   `bson:"name" json:"name"`
	Type   string   `bson:"type" json:"type"`     //type
	Tags   []string `bson:"tags" json:"tags"`     //the tags
	Subs   []string `bson:"subs" json:"subs"`     //the tags
	Status string   `bson:"status" json:"status"` //file status
	Time   int64    `bson:"time" json:"time"`     //create time.
}
