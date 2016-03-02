package gfsdb

import (
	"testing"
)

func TestFile(t *testing.T) {
	var file = &File{
		Fid:   "xxx",
		Oid:   "1",
		Owner: "USR",
	}
	var updated, err = FOI_File(file)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if updated < 1 {
		t.Error("error")
		return
	}
	updated, err = FOI_File(file)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if updated > 0 {
		t.Error("error")
		return
	}
	tc, err := CountFile()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if tc != 1 {
		t.Error("error")
		return
	}
	_, err = FOI_File(&File{})
	if err == nil {
		t.Error("error")
		return
	}
}
