package gfsdb

import (
	"fmt"
	"testing"
)

func TestFile(t *testing.T) {
	var file = &File{
		Fid:    "xxx",
		Oid:    "1",
		Owner:  "USR",
		EXT:    ".txt",
		Type:   FT_FILE,
		Status: FS_N,
		Tags:   []string{"f0"},
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
	_, err = FOI_File(&File{
		Type: FT_FILE,
	})
	if err == nil {
		t.Error("error")
		return
	}

	//
	file = &File{
		Name:   "xkdd",
		Oid:    "1",
		Owner:  "USR",
		EXT:    ".txt",
		Type:   FT_FOLDER,
		Status: FS_N,
	}
	updated, err = FOI_File(file)
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
	_, err = FOI_File(&File{
		Type: FT_FOLDER,
	})
	if err == nil {
		t.Error("error")
		return
	}
	//
	//test list file
	var f = &File{
		Fid:    "xxx2",
		Oid:    "1",
		Owner:  "USR",
		EXT:    ".txt",
		Type:   FT_FILE,
		Pid:    file.Id,
		Status: FS_N,
		Tags:   []string{"f1"},
	}
	updated, err = FOI_File(f)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if updated < 1 {
		t.Error("error")
		return
	}
	var file2 = &File{
		Name:   "xkdd",
		Oid:    "1",
		Pid:    file.Id,
		Owner:  "USR",
		EXT:    ".txt",
		Type:   FT_FOLDER,
		Status: FS_N,
	}
	updated, err = FOI_File(file2)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if updated < 1 {
		t.Error("error")
		return
	}
	var f2 = &File{
		Fid:    "xd",
		Oid:    "1",
		Owner:  "USR",
		EXT:    ".txt",
		Type:   FT_FILE,
		Pid:    file2.Id,
		Status: FS_N,
		Tags:   []string{"f2"},
	}
	updated, err = FOI_File(f2)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if updated < 1 {
		t.Error("error")
		return
	}
	fs, err := ListFile("1", "USR", "", "", []string{""}, nil, nil, []string{FS_N})
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(fs) != 2 {
		fmt.Println(fs)
		t.Error("error")
		return
	}
	fs, err = ListFile("1", "USR", "", FT_FILE, []string{""}, []string{".txt"}, []string{"f0"}, []string{FS_N})
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(fs) != 1 {
		fmt.Println(fs)
		t.Error("error")
		return
	}
	fs, err = ListFile("1", "USR", "", "", []string{file2.Id}, nil, nil, []string{FS_N})
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(fs) != 1 {
		fmt.Println(fs)
		t.Error("error")
		return
	}
	fs, total, extCount, err := ListFilePaged("1", "USR", "", FT_FILE, []string{""}, []string{".txt"}, []string{"f0"}, []string{FS_N}, "", 0, 0, 0, 1, 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(fs) != 1 || total != 1 || len(extCount) != 1 || extCount[0].StrVal("ext") != ".txt" || extCount[0].IntVal("count") != 1 {
		fmt.Println(extCount)
		t.Error("error")
		return
	}
	//
	fs, _, _, err = ListFilePaged("1", "USR", "", FT_FILE, []string{""}, []string{".txt"}, []string{"f0"}, []string{FS_N}, "", 1, 0, 0, 1, 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(fs) != 0 {
		fmt.Println(fs)
		t.Error("error")
		return
	}
	//
	//test list file count
	f2 = &File{
		Fid:    "xdds",
		Oid:    "1",
		Owner:  "USR",
		EXT:    ".txt2",
		Type:   FT_FILE,
		Pid:    "",
		Status: FS_N,
		Tags:   []string{"f2"},
	}
	updated, err = FOI_File(f2)
	if err != nil || updated < 1 {
		t.Error("error")
		return
	}
	fs, total, extCount, err = ListFilePaged("1", "USR", "", FT_FILE, []string{""}, []string{".txt"}, []string{"f0"}, []string{FS_N}, "", 0, 0, 0, 1, 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(fs) != 1 || total != 1 || len(extCount) != 2 {
		fmt.Println(extCount)
		t.Error("error")
		return
	}
	//
	//test clear Tags
	var f3 = &File{
		Fid:    "xd22",
		Oid:    "1",
		Owner:  "USR",
		EXT:    ".txt",
		Type:   FT_FILE,
		Pid:    file2.Id,
		Status: FS_N,
		Tags:   []string{"f2"},
	}
	updated, err = FOI_File(f3)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if updated < 1 {
		t.Error("error")
		return
	}
	err = UpdateFile(&File{
		Id:   f3.Id,
		Tags: []string{"_NONE_"},
	})
	if err != nil {
		t.Error(err.Error())
		return
	}
	tf, err := FindFile(f3.Id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(tf.Tags) > 0 {
		t.Error("error")
		return
	}
}
