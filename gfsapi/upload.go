package gfsapi

import (
	"encoding/base64"
	"fmt"
	"github.com/Centny/gfs/gfsdb"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/util"
	"mime/multipart"
	"path/filepath"
	"strings"
)

func (f *FSH) Up(hs *routing.HTTPSession) routing.HResult {
	var (
		pub, base64                    int = 0, 0
		name, mark, tags, desc, folder string
	)
	fmt.Println("up.....")
	err := hs.ValidCheckValN(`
		pub,O|I,O:0~1;
		base64,O|I,O:0~1;
		name,O|S,L:0~255;
		mark,O|S,L:0~255;
		tags,O|S,L:0~255;
		desc,O|S,L:0~255;
		folder,O|S,L:0~255;
		`, &pub, &base64, &name, &mark, &tags, &desc, &folder)
	if err != nil {
		log.D("FSH upload receive bad arguments error:%v", err.Error())
		return hs.MsgResErr2(-1, "arg-err", err)
	}
	log.D("FSH receive file upload by pub(%v),base64(%v),name(%v),mark(%v),tags(%v),folder(%v)",
		pub, base64, name, mark, tags, folder)
	var (
		pub_url string
		rf      = &gfsdb.F{}
	)
	if base64 == 0 {
		err = f.do_file(hs, rf, name)
	} else {
		err = f.do_base64(hs, rf, name)
	}
	if err != nil {
		return hs.MsgResErr2(-2, "srv-err", err)
	}
	rf.Time, rf.Status = util.Now(), "N"
	if pub > 0 {
		rf.Pub = "F/" + util.ShortLink(rf.SHA+rf.MD5+rf.Path+rf.EXT)
		pub_url = fmt.Sprintf("%v/%v", f.Host, rf.Pub)
	}
	_, err = gfsdb.FOI_F(rf)
	if err != nil {
		err = util.Err("FSH find or insert really file by (%v) error->%v", util.S2Json(rf), err)
		log.E("%v", err)
		return hs.MsgResErr2(-3, "srv-err", err)
	}
	if len(mark) > 0 {
		rf, err = gfsdb.AddMarkF(rf.Id, strings.Split(mark, ","))
		if err != nil {
			err = util.Err("FSH adding mark to really file by id(%v),mark(%v) error->%v", rf.Id, mark, err)
			log.E("%v", err)
			return hs.MsgResErr2(-4, "srv-err", err)
		}
	}
	if len(folder) > 0 {
		_, err = gfsdb.FindFolder(folder)
		if err != nil {
			err = util.Err("FSH check folder exist by id(%v) error->%v", folder, err)
			log.E("%v", err)
			return hs.MsgResErr2(-4, "srv-err", err)
		}
	}
	var file = &gfsdb.File{}
	file.Fid, file.Name = rf.Id, rf.Name
	file.Oid, file.Owner = hs.StrVal("uid"), "USR"
	if len(tags) > 0 {
		file.Tags = strings.Split(tags, ",")
	}
	file.Desc, file.Folder = desc, folder
	file.Time, file.Status = util.Now(), "N"
	updated, err := gfsdb.FOI_File(file)
	if err == nil {
		log.D("FSH add file pub(%v),base64(%v),name(%v),mark(%v),tags(%v),folder(%v) success",
			pub, base64, name, mark, tags, folder)
		return hs.MsgRes(util.Map{
			"url":   pub_url,
			"base":  rf,
			"file":  file,
			"added": updated,
		})
	} else {
		err = util.Err("FSH find or insert user file by (%v) error->%v", util.S2Json(file), err)
		log.E("%v", err)
		return hs.MsgResErr2(-5, "srv-err", err)
	}
}

func (f *FSH) do_file(hs *routing.HTTPSession, rf *gfsdb.F, name string) error {
	var _, size, sha_, md5_, err = hs.RecFvV2(f.Key, func(part *multipart.Part) string {
		rf.Filename = part.FileName()
		rf.EXT = strings.ToLower(filepath.Ext(rf.Filename))
		_, rf.Path = f.Base.NewFile(hs, rf.Filename)
		var spath = f.Base.AbsPath(hs, rf.Path)
		var ts = part.Header["Content-Type"]
		if len(ts) > 0 {
			rf.Type = ts[0]
		}
		return spath
	})
	if err == nil {
		rf.SHA, rf.MD5, rf.Size, rf.Name = sha_, md5_, size, name
		if len(rf.Name) < 1 {
			rf.Name = rf.Filename
		}
	} else {
		err = util.Err("FSH upload receive file by key(%v) error->%v", f.Key, err)
		log.E("%v", err)
	}
	return err
}

func (f *FSH) do_base64(hs *routing.HTTPSession, rf *gfsdb.F, name string) error {
	if len(name) < 1 {
		var err = util.Err("FSH do upload by base64 format fail with name is empty")
		log.E("%v", err)
		return err
	}
	var ts = hs.R.Header["Content-Type"]
	if len(ts) > 0 {
		rf.Type = ts[0]
	}
	if len(rf.Type) < 1 {
		var err = util.Err("FSH do upload by base64 format fail with Content-Type header is empty")
		log.E("%v", err)
		return err
	}
	_, rf.Path = f.Base.NewFile(hs, name)
	var spath = f.Base.AbsPath(hs, rf.Path)
	var reader = base64.NewDecoder(base64.StdEncoding, hs.R.Body)
	size_, sha_, md5_, err := util.Copyp3(spath, reader)
	if err == nil {
		rf.SHA, rf.MD5, rf.Size = sha_, md5_, size_
		rf.Filename, rf.Name = name, name
		rf.EXT = filepath.Ext(name)
	} else {
		err = util.Err("FSH do base64 fail with write data to path(%v) error->%v", spath, err)
		log.E("%v", err)
	}
	return err
}
