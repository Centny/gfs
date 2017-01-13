package gfsapi

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Centny/gfs/gfsdb"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/util"
	tmgo "gopkg.in/mgo.v2"
)

//File Download(Private)
//Download file by file id or file mark, It always is used to download private file.
//it can be intercepted by filter.AttrFilter to do access control.
//
//
//@url,normal http get request
//	~/usr/api/dload?fid=xxx		GET
//@arg,the normal query arguments, at least one arguments is setted on fid/mark
//	fid		O	the file id
//	mark	O	the file mark, it is specified when file is uploaded
//	dl		O	force download file, not open in browser, default is 0, 1 is forced.
//	type	O	the extern file type, it always is used to download extern file which is created by convert task.
//	idx		O	the extern file index on file list, default is 0.
//	~/usr/api/dload?fid=xxx&type=Abc&idx=1&dl=1
//@ret,normal http file stream return.
//	not example.
//@tag,file,download,private
//@author,cny,2016-03-04
//@case,File System
func (f *FSH) Down(hs *routing.HTTPSession) routing.HResult {
	var etype, mark, fid string
	var dl, idx int = 0, 0
	var err = hs.ValidCheckVal(`
		type,O|S,L:0;
		dl,O|I,O:0~1;
		idx,O|I,R:-1;
		mark,O|S,L:0;
		fid,O|S,L:0;
		`, &etype, &dl, &idx, &mark, &fid)
	if err != nil {
		hs.W.WriteHeader(404)
		fmt.Fprintf(hs.W, "the file path must be setted")
		return routing.HRES_RETURN
	}
	var rf *gfsdb.F
	if len(fid) > 0 {
		rf, err = gfsdb.FindF(fid)
	} else if len(mark) > 0 {
		rf, err = gfsdb.FindMarkF(mark)
	} else {
		hs.W.WriteHeader(400)
		fmt.Fprintf(hs.W, "at least one arguments must be setted on fid/mark")
		return routing.HRES_RETURN
	}
	if err != nil {
		hs.W.WriteHeader(500)
		fmt.Fprintf(hs.W, "find file by fid(%v)/mark(%v) error->%v", fid, mark, err)
		return routing.HRES_RETURN
	}
	return f.DoSend(hs, rf, etype, dl == 1, idx)
}

//File Download(Public)
//Download file by file public path, It always is used to download public file.
//
//
//@url,normal http get request
//	~/<public path>/<extern type>/<file index>		GET
//@arg,the normal query path,
//	dl		O	force download file, not open in browser, default is 0, 1 is forced.
//	~/F/F/bDRYOA==.jpg?dl=1
//@ret,normal http file stream return.
//	not example.
//@tag,file,download,public
//@author,cny,2016-03-04
//@case,File System
func (f *FSH) Pub(hs *routing.HTTPSession) routing.HResult {
	var path = strings.Trim(hs.R.URL.Path, "/ \t")
	if len(path) < 1 {
		hs.W.WriteHeader(404)
		fmt.Fprintf(hs.W, "the file path must be setted")
		return routing.HRES_RETURN
	}
	path = strings.TrimSuffix(path, filepath.Ext(path))
	var paths = strings.Split(path, "/")
	// if len(paths) < 1 {
	// 	hs.W.WriteHeader(404)
	// 	fmt.Fprintf(hs.W, "the file path must be setted")
	// 	return routing.HRES_RETURN
	// }
	var pub = strings.TrimSuffix(paths[0], filepath.Ext(paths[0]))
	//
	var rf, err = gfsdb.FindPubF(pub)
	if err == tmgo.ErrNotFound {
		hs.W.WriteHeader(404)
		var msg = fmt.Sprintf("file not found by pub(%v)", pub)
		log.E("%v", msg)
		fmt.Fprintf(hs.W, "%v", msg)
		return routing.HRES_RETURN
	} else if err != nil {
		hs.W.WriteHeader(500)
		var msg = fmt.Sprintf("file file fail by pub(%v)->%v", pub, err)
		log.E("%v", msg)
		fmt.Fprintf(hs.W, "%v", msg)
		return routing.HRES_RETURN
	}
	var etype = ""
	if len(paths) > 1 {
		etype = paths[1]
	}
	var idx = 0
	if len(paths) > 2 {
		idx, err = util.ParseInt(paths[2])
		if err != nil {
			hs.W.WriteHeader(400)
			var msg = fmt.Sprintf("file index(%v) is invalid->%v", paths[2], err)
			log.E("%v", msg)
			fmt.Fprintf(hs.W, "%v", msg)
			return routing.HRES_RETURN
		}
	}
	var dl = hs.CheckVal("dl") == "1"
	slog("FSH do pub send file by file(%v),type(%v),dl(%v),idx(%v)", rf.Pub, etype, dl, idx)
	return f.DoSend(hs, rf, etype, dl, idx)
}

func (f *FSH) DoSend(hs *routing.HTTPSession, rf *gfsdb.F, etype string, dl bool, idx int) routing.HResult {
	var ttype = "Default"
	if len(etype) > 0 {
		// if rf.Info == nil || len(rf.Info) < 1 {
		// 	hs.W.WriteHeader(404)
		// 	var msg = fmt.Sprintf("file(%v,%v) /info attribute is not exist, the type/index operator is not supported", rf.Id, rf.Pub)
		// 	log.E("%v", msg)
		// 	fmt.Fprintf(hs.W, "%v", msg)
		// 	return routing.HRES_RETURN
		// }
		ttype = etype
	}
	var sender, ok = f.SenderL[ttype]
	if !ok {
		hs.W.WriteHeader(404)
		var msg = fmt.Sprintf("sender is not exist by alias(%v) on file(%v,%v)", ttype, rf.Id, rf.Pub)
		log.E("%v", msg)
		fmt.Fprintf(hs.W, "%v", msg)
		return routing.HRES_RETURN
	}
	log.D("FSH call sender(%v,%v) by file(%v,%v),type(%v),dl(%v),idx(%v)", ttype, sender, rf.Id, rf.Pub, etype, dl, idx)
	return sender.Send(hs, rf, etype, dl, idx)
}
