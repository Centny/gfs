package gfsapi

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/util"
	"io/ioutil"
)

var SrvAddr = func() string {
	panic("the gfs server address is not initial")
}
var SrvArgs = func() string {
	return ""
}

func DoUpF(file, name, mark, tags, folder, desc string, pub int) (util.Map, error) {
	var url = fmt.Sprintf(
		"%v/usr/api/uload?name=%v&mark=%v&tags=%v&folder=%v&desc=%v&pub=%v&%v",
		SrvAddr(), name, mark, tags, folder, desc, pub, SrvArgs())
	log.D("DoUpF upload file to %v", url)
	var res, err = util.HPostF2(url, nil, "file", file)
	if err != nil {
		return nil, err
	}
	fmt.Println(res)
	if res.IntVal("code") == 0 {
		return res.MapVal("data"), nil
	} else {
		return nil, util.Err(
			"upload file by file(%v)name(%v),mark(%v),tags(%v),folder(%v),desc(%v),pub(%v) error->%v",
			file, name, mark, tags, folder, desc, pub, util.S2Json(res))
	}
}

func DoUpBase64(buf, ctype, name, mark, tags, folder, desc string, pub int) (util.Map, error) {
	var base64 = bytes.NewBufferString(buf)
	var _, res, err = util.HPostN2(fmt.Sprintf(
		"%v/usr/api/uload?name=%v&mark=%v&tags=%v&folder=%v&desc=%v&pub=%v&base64=1&%v",
		SrvAddr(), name, mark, tags, folder, desc, pub, SrvArgs()), ctype, base64)
	if err != nil {
		return nil, err
	}
	if res.IntVal("code") == 0 {
		return res.MapVal("data"), nil
	} else {
		return nil, util.Err(
			"upload file by type(%v),name(%v),mark(%v),tags(%v),folder(%v),desc(%v),pub(%v) error->%v",
			ctype, name, mark, tags, folder, desc, pub, util.S2Json(res))
	}
}

func DoInfo(fid, sha, md5, mark string) (util.Map, error) {
	var res, err = util.HGet2(
		"%v/pub/api/info?fid=%v&sha=%v&md5=%v&mark=%v&%v",
		SrvAddr(), fid, sha, md5, mark, SrvArgs())
	if err != nil {
		return nil, err
	}
	if res.IntVal("code") == 0 {
		return res.MapVal("data"), nil
	} else {
		return nil, util.Err(
			"query file info by fid(%v),sha(%v),md5(%v),mark(%v) error->%v",
			fid, sha, md5, mark, util.S2Json(res))
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
		return base64.RawStdEncoding.EncodeToString(bys), nil
	} else {
		return "", err
	}
}
