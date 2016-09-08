package gfsapi

import (
	"github.com/Centny/gfs/gfsdb"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/util"
)

//List/Search User File/Folder
//List/Search login user file or folder
//
//@url,normal http get request
//	~/usr/api/listFile		GET
//@arg,the normal query arguments
//	name	O	the search key for file name
//	type	O	the type in file/folder to show the file or folder
//	pid		O	the parent folder id
//	tags	O	the file/folder tags to filter
/*
	//list user file or folder
	~/usr/api/listFile
	//list user file
	~/usr/api/listFile?type=file
	//searhc file
	~/usr/api/listFile?type=file&name=xx
*/
//@ret,code/data return
//	bases	O	the file base info, see upload api for deatail
//	files	A	the user file info.
/*	the example
	{
	    "code": 0,
	    "data": {
	        "bases": {
	            "57bd539dc3666e997e75f288": {
	                "exec": "done",
	                "ext": ".mp4",
	                "filename": "xx.mp4",
	                "id": "57bd539dc3666e997e75f288",
	                "info": {
	                    "V_json": {
	                        "count": 1,
	                        "files": ["www/2016-08-24/u_57bd539dc3666e997e000002_js.mp4"]
	                    },
	                    "V_pc": {
	                        "text": "www/2016-08-24/u_57bd539dc3666e997e000002_pc.mp4"
	                    },
	                    "code": 0
	                },
	                "md5": "52757d83284ca0967bc0c9e2be342c13",
	                "name": "xx.mp4",
	                "pub": "HI2hmt==",
	                "sha": "226cf3e82860ea778ccae40a9e424be5700249e1",
	                "size": 431684,
	                "status": "N",
	                "time": 1.472025501957e+12,
	                "type": "application/octet-stream"
	            }
	        },
	        "files": [{
	            "desc": "desc",
	            "fid": "57bd539dc3666e997e75f288",
	            "id": "57bd539dc3666e997e75f289",
	            "name": "xx.mp4",
	            "oid": "123",
	            "owner": "USR",
	            "pid": "57bd539ac3666e997e75f287",
	            "status": "N",
	            "tags": ["x", "y", "z"],
	            "time": 1.472025501961e+12,
	            "type": "file"
	        }]
	    }
	}
*/
//@tag,file,info,list
//@author,cny,2016-08-24
//@case,File System
func ListFile(hs *routing.HTTPSession) routing.HResult {
	var name, typ string
	var pid, tags []string
	var err = hs.ValidCheckVal(`
		name,O|S,L:0;
		type,O|S,O:file~folder;
		pid,O|S,L:0;
		tags,O|S,L:0;
		`, &name, &typ, &pid, &tags)
	if err != nil {
		return hs.MsgResErr2(1, "arg-err", err)
	}
	var uid = hs.StrVal("uid")
	fs, err := gfsdb.ListFile(uid, OWN_USR, name, typ, pid, tags, []string{gfsdb.FS_N})
	if err != nil {
		err = util.Err("ListFile list find by oid(%v),owner(%v),name(%v),type(%v),pid(%v),tags(%v) fail with error(%v)",
			uid, OWN_USR, name, typ, pid, tags, err)
		log.E("%v", err)
		return hs.MsgResErr2(2, "srv-err", err)
	}
	var ids = []string{}
	for _, f := range fs {
		ids = append(ids, f.Fid)
	}
	bases, err := gfsdb.ListF_m(ids)
	if err != nil {
		err = util.Err("ListFile list base file ids(%v) fail with error(%v)", ids, err)
		log.E("%v", err)
		return hs.MsgResErr2(3, "srv-err", err)
	}
	return hs.MsgRes(util.Map{
		"bases": bases,
		"files": fs,
	})
}

//Update user file or folder
//Update user file or foild by id
//
//@url,normal http get request
//	~/usr/api/updateFile?fid=xx		GET
//@arg,the normal query arguments
//	fid		R	the file/folder id
//	name	O	the file/folder name
//	desc	O	the file/folder desc
//	tags	O	the file/folder tags
/*
	//update file/folder name
	~/usr/api/updateFile?fid=xx&name=aaa
*/
//@ret,code/data return
//	code	I	the common code.
/*	the example
	{
	    "code": 0,
	    "data": "OK"
	}
*/
//@tag,file,info,update
//@author,cny,2016-08-24
//@case,File System
func UpdateFile(hs *routing.HTTPSession) routing.HResult {
	var file = &gfsdb.File{}
	var err = hs.ValidCheckVal(`
		fid,R|S,L:0;
		name,O|S,L:0;
		desc,O|S,L:0;
		tags,O|S,L:0;
		`, &file.Id, &file.Name, &file.Desc, &file.Tags)
	if err != nil {
		return hs.MsgResErr2(1, "arg-err", err)
	}
	err = gfsdb.UpdateFile(file)
	if err == nil {
		return hs.MsgRes("OK")
	} else {
		return hs.MsgResErr2(2, "srv-err", err)
	}
}
