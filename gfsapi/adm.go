package gfsapi

import (
	"sync/atomic"

	"strings"

	"github.com/Centny/gfs/gfsdb"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/util"
)

var CFG *util.Fcfg = nil
var VerifyRunning int32

func AdmVerify(hs *routing.HTTPSession) routing.HResult {
	if CFG == nil {
		return hs.MsgResE3(1, "srv-err", "the configure is not inital")
	}
	if !atomic.CompareAndSwapInt32(&VerifyRunning, 0, 1) {
		return hs.MsgResE3(2, "srv-err", "having task running")
	}
	var ids []string
	var err = hs.ValidCheckVal(`
        ids,O|S,L:0;
    `, &ids)
	if err != nil {
		return hs.MsgResErr2(2, "srv-err", err)
	}
	go runVerify(ids)
	return hs.MsgRes("OK")
}

func runVerify(ids []string) {
	defer func() {
		err := recover()
		if err != nil {
			log.E("RunVerify panic with %v:\n%v", err, util.CallStatck())
		}
	}()
	var supported = CFG.Val("supported_v")
	if len(supported) < 1 {
		log.W("RunVerify the supported_v is empty")
		return
	}
	total, fail, err := gfsdb.VerifyVideo(CFG.Val("w_dir_i"), CFG.Val("w_dir_o"), strings.Split(supported, ","), ids, nil)
	if err == nil {
		log.D("RunVerify done success with total(%v),fail(%v)", total, fail)
	} else {
		log.E("RunVerify done fail with total(%v),fail(%v),err(%v)", total, fail, err)
	}
	atomic.StoreInt32(&VerifyRunning, 0)
}

func RedoTask(hs *routing.HTTPSession) routing.HResult {
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
	//log.D("FSH query file info by fid(%v)/sha(%v)/md5(%v)/mark(%v) success", fid, sha, md5, mark)
	if file.Exec == gfsdb.ES_RUNNING {
		return hs.MsgResE3(3, "arg-err", "the file exec is running")
	}
	_, err = gfsdb.DoAddTask(file)
	if err == nil {
		return hs.MsgRes("OK")
	}
	return hs.MsgResErr2(5, "srv-err", err)
}
