package gfsdb

import (
	"path/filepath"
	"strings"

	"gopkg.in/mgo.v2"
)

const (
	CN_F      = "c_f"
	CN_MARK   = "c_mark"
	CN_FILE   = "c_file"
	CN_FOLDER = "c_folder"
)

var ShowLog = 0
var OutPathPre = ""
var CreateOutPath = func(f *F) string {
	if len(f.Path) < 1 {
		return ""
	}
	var out = OutPathPre + strings.TrimSuffix(f.Path, filepath.Ext(f.Path))
	if f.Path == out {
		return out + "_"
	} else {
		return out
	}
}

var C = func(n string) *mgo.Collection {
	panic("the gfs database caller is not initial")
}
