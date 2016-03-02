package gfsapi

import (
	"fmt"
	"github.com/Centny/ffcm"
	"github.com/Centny/gfs/gfsdb"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/netw/dtm"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/util"
	"path/filepath"
	"strings"
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
	return f.Base, fmt.Sprintf("%v%v%v%v", f.Pre, util.UUID(), f.Subs[sub], ext)
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

type FSH struct {
	Base    FBase
	Key     string
	Host    string
	SenderL map[string]FSedner
}

func NewFSH(base FBase) *FSH {
	return &FSH{
		Base:    base,
		Key:     "file",
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
func (f *FSH) Info(hs *routing.HTTPSession) routing.HResult {
	var err error
	var fid, sha, md5, mark string
	hs.ValidCheckVal(`
		fid,O|S,L:0;
		sha,O|S,L:0;
		md5,O|S,L:0;
		mark,O|S,L:0;
		`, &fid, &sha, &md5, &mark)
	var file *gfsdb.F
	if len(fid) > 0 {
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
	if ffcm.SRV == nil {
		return hs.MsgRes(util.Map{
			"file": file,
		})
	}
	var task *dtm.Task
	task, err = ffcm.SRV.Db.Find(file.Id)
	if err == nil {
		return hs.MsgRes(util.Map{
			"file": file,
			"task": task,
		})
	} else {
		err = util.Err("FSH find task by id (%v) error->%v", file.Id, err)
		log.E("%v", err)
		return hs.MsgRes(util.Map{
			"file": file,
			"err":  err,
		})
	}
}

func (f *FSH) Hand(pre string, mux *routing.SessionMux) {
	mux.HFunc("^"+pre+"/pub/api/info(\\?.*)?", f.Info)
	mux.HFunc("^"+pre+"/usr/api/uload(\\?.*)?", f.Up)
	mux.HFunc("^"+pre+"/usr/api/dload(\\?.*)?", f.Down)
	mux.HFunc("^"+pre+"/.*$", f.Pub)
}
