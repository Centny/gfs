package gfsapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Centny/ffcm"
	"github.com/Centny/gfs/gfsdb"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/util"
)

const (
	OWN_USR = "USR"
)

var ShowLog bool = false

func slog(format string, args ...interface{}) {
	if ShowLog {
		log.D_(1, format, args...)
	}
}

type FBase interface {
	NewFile(hs *routing.HTTPSession, filename string) (string, string)
	AbsPath(hs *routing.HTTPSession, path string) string
}

type FBaseImpl struct {
	Base string
	Pre  string
	Subs map[string]string
}

func NewFBaseImpl(base, pre string) *FBaseImpl {
	return &FBaseImpl{
		Base: base,
		Pre:  pre,
		Subs: map[string]string{},
	}
}
func (f *FBaseImpl) NewFile(hs *routing.HTTPSession, filename string) (string, string) {
	var sub = hs.CheckVal("sub")
	var ext = strings.ToLower(filepath.Ext(filename))
	if len(sub) < 1 || len(f.Subs[sub]) < 1 {
		sub = ext
	}
	var pre = f.Pre
	if strings.Contains(pre, "%v") {
		pre = fmt.Sprintf(pre, time.Now().Format("2006-01-02"))
	}
	return f.Base, fmt.Sprintf("%v%v%v%v", pre, util.UUID(), f.Subs[sub], ext)
}
func (f *FBaseImpl) AbsPath(hs *routing.HTTPSession, path string) string {
	return filepath.Join(f.Base, path)
}
func (f *FBaseImpl) ParseSubs(fcfg *util.Fcfg, sec string) int {
	var count = 0
	fcfg.Range(sec, func(key string, val interface{}) {
		var vals = strings.Split(fmt.Sprintf("%v", val), ",")
		for _, val := range vals {
			f.Subs[val] = key
			count += 1
		}
	})
	return count
}

func FilterTaskInfo(fs []*gfsdb.F) {
	for _, f := range fs {
		if f.Info == nil {
			return
		}
		for key, _ := range f.Info {
			var mv = f.Info.MapVal(key)
			if mv == nil {
				continue
			}
			delete(mv, "files")
			delete(mv, "src")
			f.Info.SetVal(key, mv)
		}
	}
}

type FSH struct {
	Base    FBase
	Key     string
	Host    string
	Mode    os.FileMode
	SenderL map[string]FSedner
}

func NewFSH(base FBase) *FSH {
	return &FSH{
		Base:    base,
		Key:     "file",
		Mode:    os.ModePerm,
		SenderL: map[string]FSedner{},
	}
}

func NewFSH2(fcfg *util.Fcfg) (*FSH, error) {
	var base = fcfg.Val2("w_dir", ".")
	var pre_u = fcfg.Val2("pre_u", "")
	var host = fcfg.Val2("host", "")
	var subs = fcfg.Val2("subs", "Subs")
	var impl = NewFBaseImpl(base, pre_u)
	var subs_c = impl.ParseSubs(fcfg, subs)
	var fsh = NewFSH(impl)
	fsh.Host = host
	fsh.Mode = fcfg.FileModeV("mode", os.ModePerm)
	var sender_l = fcfg.Val2("sender_l", "")
	if len(sender_l) > 0 {
		var ss, err = ParseSenderL(fcfg, strings.Split(sender_l, ","))
		if err != nil {
			return nil, err
		}
		fsh.SenderL = ss
	}
	log.D("create FSH success with subs(%v), w_dir(%v), pre_u(%v), host(%v), sender(%v)",
		subs_c, base, pre_u, host, len(fsh.SenderL))
	return fsh, nil
}

func (f *FSH) AddSender(alias string, s FSedner) {
	f.SenderL[alias] = s
}

//File Info
//Get the file information by file id/mark/sha/md5,
//the file informantion container file normal info like SHA1,MD5,size and the extern file info.
//
//@url,normal http get request
//	~/pub/api/info?fid=xxx		GET
//@arg,the normal query arguments, at least one arguments is setted on fid/mark/sha/md5
//	fid		O	the file id
//	mark	O	the file mark, it is specified when file is uploaded
//	sha		O	the file SHA.
//	md5		O	the file MD5.
//	pub		O	the file pub.
//	~/pub/api/info?fid=xxx
//@ret,code/data return
//	base			O	the file base information
//	exec			O	the converter executing information, not this field when the task is not started/done or get task status fail.
//	err				S	the error inforamtion when get converter task status fail.
//	exec.total		F	the converter task process rate
//	exec.detail		O	the sub task process rate
//	base.filename 	S	the uploaed file name
//	base.id			S	the file id
//	base.mark		A	the file mark list.
//	base.md5		S	the file MD5 hash
//	base.name		S	the special name
//	base.sha		S	the file SHA1 hash.
//	base.size		I	the file size.
//	base.time		I	the file upload time
//	base.type		S	the file mime type.
//	base.pub		S	the file public path.
/*	the example
	{
		"code": 0,
		"data": {
			"exec": {
				"detail": {
					"V_json": 0,
					"V_pc": 0
				},
				"total": 0
			},
			"base": {
				"exec": "running",
				"ext": ".mp4",
				"filename": "../../ffcm/xx.mp4",
				"id": "56d9a4eec3666e4e02af307f",
				"info": {},
				"mark": ["xxa"],
				"md5": "52757d83284ca0967bc0c9e2be342c13",
				"name": "../../ffcm/xx.mp4",
				"path": "www/u_56d9a4eec3666e4e02000001.mp4",
				"pub": "F/bDRYOA==",
				"sha": "226cf3e82860ea778ccae40a9e424be5700249e1",
				"size": 431684,
				"status": "N",
				"time": 1.457104110367e+12,
				"type": "application/octet-stream"
			}
		}
	}
*/
//@tag,file,info
//@author,cny,2016-03-05
//@case,File System
func (f *FSH) Info(hs *routing.HTTPSession) routing.HResult {
	var err error
	var fid, sha, md5, mark, pub string
	hs.ValidCheckVal(`
		fid,O|S,L:0;
		sha,O|S,L:0;
		md5,O|S,L:0;
		mark,O|S,L:0;
		pub,O|S,L:0;
		`, &fid, &sha, &md5, &mark, &pub)
	var file *gfsdb.F
	if len(pub) > 0 {
		file, err = gfsdb.FindPubF(pub)
	} else if len(fid) > 0 {
		file, err = gfsdb.FindF(fid)
	} else if len(sha) > 0 || len(md5) > 0 {
		file, err = gfsdb.FindHashF(sha, md5)
	} else if len(mark) > 0 {
		file, err = gfsdb.FindMarkF(mark)
	} else {
		return hs.MsgResE3(2, "arg-err", "at least one argments must be setted on fid/sha/md5/mark")
	}
	if err != nil {
		err = util.Err("FSH find file by fid(%v),sha(%v),md5(%v),mark(%v) error->%v", fid, sha, md5, mark, err)
		log.E("%v", err)
		return hs.MsgResErr2(1, "srv-err", err)
	}
	FilterTaskInfo([]*gfsdb.F{file})
	log.D("FSH query file info by fid(%v)/sha(%v)/md5(%v)/mark(%v) success", fid, sha, md5, mark)
	var mres = util.Map{
		"base": file,
	}
	if len(file.Pub) > 0 {
		mres["url"] = fmt.Sprintf("%v/%v", f.Host, file.Pub)
	}
	if file.Exec != gfsdb.ES_RUNNING || ffcm.SRV == nil {
		return hs.MsgRes(mres)
	}
	log.D("FSH query file convert info by fid(%v)", fid)
	task, err := ffcm.SRV.Db.Find(file.Id)
	if err == nil {
		mres["task"] = util.Map{
			"code":   task.Code,
			"status": task.Status,
		}
	} else {
		mres["task"] = util.Map{
			"err": err.Error(),
		}
		return hs.MsgRes(mres)
	}
	total, res, err := ffcm.SRV.TaskRate(file.Id)
	if err == nil {
		mres["exec"] = util.Map{
			"total":  total,
			"detail": res,
		}
	} else {
		mres["exec"] = util.Map{
			"err": err.Error(),
		}
	}
	return hs.MsgRes(mres)
}

//List File Info
//Get the file information by file id/mark/sha/md5,
//the file informantion container file normal info like SHA1,MD5,size and the extern file info.
//
//@url,normal http get request
//	~/pub/api/listInfo?fid=xxx,xxk		GET
//@arg,the normal query arguments, at least one arguments is setted on fid/mark/sha/md5
//	fid		O	the file id split by comma
//	mark	O	the file mark split by comma, it is specified when file is uploaded
//	sha		O	the file SHA split by comma.
//	md5		O	the file MD5 split by comma.
//	pub		O	the file pub split by comma.
//	mode	O	the return data mode, default is list, it will return map value by special key value with setting mode is one of fid/mark/sha/md5/pub
//	~/pub/api/listInfo?fid=xxx,xx
//@ret,code/data return
//	base			O	the file base information
//	exec			O	the converter executing information, not this field when the task is not started/done or get task status fail.
//	err				S	the error inforamtion when get converter task status fail.
//	exec.total		F	the converter task process rate
//	exec.detail		O	the sub task process rate
//	base.filename 	S	the uploaed file name
//	base.id			S	the file id
//	base.mark		A	the file mark list.
//	base.md5		S	the file MD5 hash
//	base.name		S	the special name
//	base.sha		S	the file SHA1 hash.
//	base.size		I	the file size.
//	base.time		I	the file upload time
//	base.type		S	the file mime type.
//	base.pub		S	the file public path.
/*	the example
	{
		"code": 0,
		"data": [{
			"exec": {
				"detail": {
					"V_json": 0,
					"V_pc": 0
				},
				"total": 0
			},
			"base": {
				"exec": "running",
				"ext": ".mp4",
				"filename": "../../ffcm/xx.mp4",
				"id": "56d9a4eec3666e4e02af307f",
				"info": {},
				"mark": ["xxa"],
				"md5": "52757d83284ca0967bc0c9e2be342c13",
				"name": "../../ffcm/xx.mp4",
				"path": "www/u_56d9a4eec3666e4e02000001.mp4",
				"pub": "F/bDRYOA==",
				"sha": "226cf3e82860ea778ccae40a9e424be5700249e1",
				"size": 431684,
				"status": "N",
				"time": 1.457104110367e+12,
				"type": "application/octet-stream"
			}
		}]
	}
*/
//@tag,file,info,list
//@author,cny,2016-04-15
//@case,File System
func (f *FSH) ListInfo(hs *routing.HTTPSession) routing.HResult {
	var err error
	var fid, sha, md5, mark, pub string
	var mode = ""
	hs.ValidCheckVal(`
		fid,O|S,L:0;
		sha,O|S,L:0;
		md5,O|S,L:0;
		mark,O|S,L:0;
		pub,O|S,L:0;
		mode,O|S,L:0;
		`, &fid, &sha, &md5, &mark, &pub, &mode)
	var files []*gfsdb.F
	var mfids = map[string]string{}
	if len(pub) > 0 {
		files, err = gfsdb.ListPubF(strings.Split(pub, ","))
	} else if len(fid) > 0 {
		files, err = gfsdb.ListF(strings.Split(fid, ","))
	} else if len(sha) > 0 || len(md5) > 0 {
		var sha_a, md5_a []string
		if len(sha) > 0 {
			sha_a = strings.Split(sha, ",")
		}
		if len(md5) > 0 {
			md5_a = strings.Split(md5, ",")
		}
		files, err = gfsdb.ListHashF(sha_a, md5_a)
	} else if len(mark) > 0 {
		files, mfids, err = gfsdb.ListMarkFv(strings.Split(mark, ","))
	} else {
		return hs.MsgResE3(2, "arg-err", "at least one argments must be setted on fid/sha/md5/mark")
	}
	if err != nil {
		err = util.Err("FSH find file by fid(%v),sha(%v),md5(%v),mark(%v) error->%v", fid, sha, md5, mark, err)
		log.E("%v", err)
		return hs.MsgResErr2(1, "srv-err", err)
	}
	FilterTaskInfo(files)
	log.D("FSH query file info by fid(%v)/sha(%v)/md5(%v)/mark(%v)/pub(%v) success", fid, sha, md5, mark, pub)
	var fis = []util.Map{}
	var mfis = util.Map{}
	var addf = func(file *gfsdb.F, mv util.Map) {
		switch mode {
		case "fid":
			mfis[file.Id] = mv
		case "sha":
			mfis[file.SHA] = mv
		case "md5":
			mfis[file.MD5] = mv
		case "mark":
			for mark, fid := range mfids {
				if file.Id == fid {
					mfis[mark] = mv
				}
			}
		case "pub":
			mfis[file.Pub] = mv
		default:
			fis = append(fis, mv)
		}
	}
	for _, file := range files {
		if file.Exec != gfsdb.ES_RUNNING || ffcm.SRV == nil {
			addf(file, util.Map{
				"base": file,
			})
			continue
		}
		// log.D("FSH query file convert info by fid(%v)", fid)
		total, res, err := ffcm.SRV.TaskRate(file.Id)
		if err == nil {
			addf(file, util.Map{
				"base": file,
				"exec": util.Map{
					"total":  total,
					"detail": res,
				},
				"url": fmt.Sprintf("%v/%v", f.Host, file.Pub),
			})
		} else {
			addf(file, util.Map{
				"base": file,
				"err":  err,
			})
		}
	}
	if mode == "" {
		return hs.MsgRes(fis)
	} else {
		return hs.MsgRes(mfis)
	}
}

func SetMock(hs *routing.HTTPSession) routing.HResult {
	if hs.RVal("start_err") == "1" {
		gfsdb.MockStartTaskErr = 1
	} else {
		gfsdb.MockStartTaskErr = 0
	}
	return hs.MsgRes("OK")
}
