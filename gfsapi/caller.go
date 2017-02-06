package gfsapi

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/util"
)

var SrvAddr = func() string {
	panic("the gfs server address is not initial")
}
var SrvArgs = func() string {
	return ""
}

func DoUpF(file, name, mark, tags, folder, desc string, pub, recorded int) (util.Map, error) {
	var url = fmt.Sprintf(
		"%v/usr/api/uload?name=%v&mark=%v&tags=%v&folder=%v&desc=%v&pub=%v&recorded=%v&%v",
		SrvAddr(), name, mark, tags, folder, desc, pub, recorded, SrvArgs())
	log.D("DoUpF upload file(%v) to %v", file, url)
	var res, err = util.HPostF2(url, nil, "file", file)
	if err != nil {
		return nil, err
	}
	if res.IntVal("code") == 0 {
		return res, nil
	} else {
		return nil, util.Err(
			"upload file by file(%v)name(%v),mark(%v),tags(%v),folder(%v),desc(%v),pub(%v) error->%v",
			file, name, mark, tags, folder, desc, pub, util.S2Json(res))
	}
}

func DoUpBase64(buf, ctype, name, mark, tags, folder, desc string, pub, recorded int) (util.Map, error) {
	var base64 = bytes.NewBufferString(buf)
	var _, res, err = util.HPostN2(fmt.Sprintf(
		"%v/usr/api/uload?name=%v&mark=%v&tags=%v&folder=%v&desc=%v&pub=%v&base64=1&recorded=%v&%v",
		SrvAddr(), name, mark, tags, folder, desc, pub, recorded, SrvArgs()), ctype, base64)
	if err != nil {
		return nil, err
	}
	if res.IntVal("code") == 0 {
		return res, nil
	} else {
		return nil, util.Err(
			"upload file by type(%v),name(%v),mark(%v),tags(%v),folder(%v),desc(%v),pub(%v) error->%v",
			ctype, name, mark, tags, folder, desc, pub, util.S2Json(res))
	}
}

func DoInfo(fid, sha, md5, mark, pub string) (util.Map, error) {
	var res, err = util.HGet2(
		"%v/pub/api/info?fid=%v&sha=%v&md5=%v&mark=%v&pub=%v&%v",
		SrvAddr(), fid, sha, md5, mark, pub, SrvArgs())
	if err != nil {
		return nil, err
	}
	if res.IntVal("code") == 0 {
		return res.MapVal("data"), nil
	} else {
		return nil, util.Err(
			"query file info by fid(%v),sha(%v),md5(%v),mark(%v),pub(%v) error->%v",
			fid, sha, md5, mark, pub, util.S2Json(res))
	}
}
func DoRedoTask(fid, sha, md5, mark, pub string) error {
	var res, err = util.HGet2(
		"%v/usr/api/redoTask?fid=%v&sha=%v&md5=%v&mark=%v&pub=%v&%v",
		SrvAddr(), fid, sha, md5, mark, pub, SrvArgs())
	if err != nil {
		return err
	}
	if res.IntVal("code") == 0 {
		return nil
	} else {
		return util.Err(
			"redo file info by fid(%v),sha(%v),md5(%v),mark(%v),pub(%v) error->%v",
			fid, sha, md5, mark, pub, util.S2Json(res))
	}
}

func DoListInfo(fid, sha, md5, mark, pub []string) ([]util.Map, error) {
	var res, err = util.HGet2(
		"%v/pub/api/listInfo?fid=%v&sha=%v&md5=%v&mark=%v&pub=%v&%v",
		SrvAddr(), strings.Join(fid, ","), strings.Join(sha, ","),
		strings.Join(md5, ","), strings.Join(mark, ","), strings.Join(pub, ","), SrvArgs())
	if err != nil {
		return nil, err
	}
	if res.IntVal("code") == 0 {
		return res.AryMapVal("data"), nil
	} else {
		return nil, util.Err(
			"query file info by fid(%v),sha(%v),md5(%v),mark(%v),pub(%v) error->%v",
			fid, sha, md5, mark, pub, util.S2Json(res))
	}
}
func DoListInfoM(fid, sha, md5, mark, pub []string, mode string) (util.Map, error) {
	if mode != "fid" && mode != "sha" && mode != "md5" && mode != "mark" && mode != "pub" {
		return nil, util.Err("the mode must be one of fid/sha/md5/mark/pub, but %v found", mode)
	}
	var res, err = util.HGet2(
		"%v/pub/api/listInfo?fid=%v&sha=%v&md5=%v&mark=%v&pub=%v&mode=%v&%v",
		SrvAddr(), strings.Join(fid, ","), strings.Join(sha, ","),
		strings.Join(md5, ","), strings.Join(mark, ","), strings.Join(pub, ","), mode, SrvArgs())
	if err != nil {
		return nil, err
	}
	if res.IntVal("code") == 0 {
		return res.MapVal("data"), nil
	} else {
		return nil, util.Err(
			"query file info by fid(%v),sha(%v),md5(%v),mark(%v),pub(%v) error->%v",
			fid, sha, md5, mark, pub, util.S2Json(res))
	}
}

func DoFileDown(fid, mark, etype string, idx int, path string) error {
	return util.DLoad(path,
		"%v/usr/api/dload?fid=%v&type=%v&mark=%v&idx=%v&dl=1&%v",
		SrvAddr(), fid, etype, mark, idx, SrvArgs())
}

func DoPubDown(pub, path string) error {
	return util.DLoad(path, "%v/%v?dl=1&%v", SrvAddr(), pub, SrvArgs())
}

func ReadBase64(path string) (string, error) {
	var bys, err = ioutil.ReadFile(path)
	if err == nil {
		return base64.StdEncoding.EncodeToString(bys), nil
	} else {
		return "", err
	}
}

func DoAdmStatus() (util.Map, error) {
	return util.HGet2("%v/adm/status?%v", SrvAddr(), SrvArgs())
}

func DoListFile(name, typ string, pid, ext, tags []string, reverseExt, pn, ps, retExtCount int) (util.Map, error) {
	var res, err = util.HGet2(
		"%v/usr/api/listFile?name=%v&type=%v&pid=%v&ext=%v&tags=%v&not_ext=%v&pn=%v&ps=%v&ret_ext_count=%v&%v",
		SrvAddr(), name, typ, strings.Join(pid, ","), strings.Join(ext, ","), strings.Join(tags, ","), reverseExt, pn, ps, retExtCount, SrvArgs())
	if err != nil {
		return nil, err
	}
	if res.IntVal("code") == 0 {
		return res.MapVal("data"), nil
	} else {
		return nil, util.Err("list file error->%v", util.S2Json(res))
	}
}

func DoUpdateFile(fid, name, desc string, tags []string) error {
	var res, err = util.HGet2(
		"%v/usr/api/updateFile?fid=%v&name=%v&desc=%v&tags=%v&%v",
		SrvAddr(), fid, name, desc, strings.Join(tags, ","), SrvArgs())
	if err != nil {
		return err
	}
	if res.IntVal("code") == 0 {
		return nil
	}
	return util.Err("list file error->%v", util.S2Json(res))
}

func DoUpdateFileParent(fids []string, pid string) error {
	var res, err = util.HGet2(
		"%v/usr/api/updateFileParent?fids=%v&pid=%v&%v",
		SrvAddr(), strings.Join(fids, ","), pid, SrvArgs())
	if err != nil {
		return err
	}
	if res.IntVal("code") == 0 {
		return nil
	}
	return util.Err("list file error->%v", util.S2Json(res))
}

func DoRemoveFile(fid string) (int, error) {
	var res, err = util.HGet2(
		"%v/usr/api/removeFile?fid=%v&%v",
		SrvAddr(), fid, SrvArgs())
	if err != nil {
		return 0, err
	}
	if res.Exist("code") && res.IntVal("code") == 0 {
		return int(res.IntVal("data")), nil
	}
	return 0, util.Err("remove file error->%v", util.S2Json(res))
}

func DoAddFolder(pid, name, desc string, tags []string) (util.Map, error) {
	var res, err = util.HGet2(
		"%v/usr/api/addFolder?pid=%v&name=%v&desc=%v&tags=%v&%v",
		SrvAddr(), pid, name, strings.Join(tags, ","), SrvArgs())
	if err != nil {
		return nil, err
	}
	if res.Exist("code") && res.IntVal("code") == 0 {
		return res.MapVal("data"), nil
	}
	return nil, util.Err("add folder error->%v", util.S2Json(res))
}

func DoAdmVerify() error {
	var res, err = util.HGet2(
		"%v/adm/verify?%v",
		SrvAddr(), SrvArgs())
	if err != nil {
		return err
	}
	if res.Exist("code") && res.IntVal("code") == 0 {
		return nil
	}
	return util.Err("adm verify error->%v", util.S2Json(res))
}
