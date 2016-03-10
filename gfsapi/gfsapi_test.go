package gfsapi

import (
	"bytes"
	"fmt"
	"github.com/Centny/dbm/mgo"
	"github.com/Centny/ffcm"
	"github.com/Centny/gfs/gfsdb"
	_ "github.com/Centny/gfs/test"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/routing/filter"
	"github.com/Centny/gwf/routing/httptest"
	"github.com/Centny/gwf/util"
	tmgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init() {
	mgo.C(gfsdb.CN_FOLDER).RemoveAll(nil)
	mgo.C(gfsdb.CN_F).RemoveAll(nil)
	mgo.C(gfsdb.CN_MARK).RemoveAll(nil)
	func() {
		defer func() {
			recover()
		}()
		SrvAddr()
	}()
	gfsdb.C = mgo.C
}

func TestUpDown(t *testing.T) {
	ShowLog = true
	runtime.GOMAXPROCS(util.CPU())
	os.RemoveAll("www")
	os.RemoveAll("out")
	os.RemoveAll("tmp")
	var folder = &gfsdb.Floder{
		Status: "N",
	}
	gfsdb.AddFolder(folder)
	ffcm.StartTest2("../gfs_s.properties", "../gfs_c.properties", gfsdb.NewFFCM_H())
	time.Sleep(2 * time.Second)
	if ffcm.SRV == nil {
		t.Error("initial error")
		return
	}
	var fcfg = util.NewFcfg3()
	var err = fcfg.InitWithFilePath2("../gfs_s.properties", false)
	if err != nil {
		t.Error(err.Error())
		return
	}
	var ts = httptest.NewMuxServer()
	tsh, err := NewFSH2(fcfg)
	if err != nil {
		t.Error(err.Error())
		return
	}
	ts.Mux.HFilterFunc("^/usr/api/uload(\\?.*)?$", filter.ParseQuery)
	ts.Mux.HFilterFunc("^.*$", func(hs *routing.HTTPSession) routing.HResult {
		hs.SetVal("uid", "123")
		return routing.HRES_CONTINUE
	})
	tsh.Hand("", ts.Mux)
	tsh.Host = ts.URL
	SrvAddr = func() string {
		return ts.URL
	}
	//
	//test upload file
	res, err := DoUpF("../../ffcm/xx.mp4", "", "xxa", "x,y,z", folder.Id, "desc", 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(util.S2Json(res))
	// fmt.Println(util.S2Json(res))
	var fid = res.StrValP("/base/id")
	var sha = res.StrValP("/base/sha")
	var pub = res.StrValP("/base/pub")
	var url = res.StrValP("/url")
	var path = res.StrValP("/base/path")
	if len(fid) < 1 || len(sha) < 1 || len(pub) < 1 {
		t.Error("fid is empty")
		return
	}
	//
	//test file id
	time.Sleep(1 * time.Second)
	var ff_srv = ffcm.SRV
	ffcm.SRV = nil
	res, err = DoInfo(fid, "", "", "")
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(util.S2Json(res))
	if res.Val("exec") != nil {
		t.Error("error")
		return
	}
	ffcm.SRV = ff_srv
	//
	//test file id
	res, err = DoInfo(fid, "", "", "")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if fid != res.StrValP("/base/id") {
		t.Error("error")
		return
	}
	//
	//test file mark
	res, err = DoInfo("", "", "", "xxa")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if fid != res.StrValP("/base/id") {
		t.Error("error")
		return
	}
	fmt.Println(util.S2Json(res))
	//
	//test file hash
	res, err = DoInfo("", sha, "", "")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if fid != res.StrValP("/base/id") {
		t.Error("error")
		return
	}
	fmt.Println(util.S2Json(res))
	//
	//wait task done..
	time.Sleep(1 * time.Second)
	var check_c = 0
	for {
		res, err = DoInfo(fid, "", "", "")
		if err != nil {
			t.Error(err.Error())
			return
		}
		if fid != res.StrValP("/base/id") {
			t.Error("error")
			return
		}
		fmt.Println(util.S2Json(res))
		if res.MapVal("exec") == nil {
			var info = res.MapValP("/base/info")
			if info == nil || len(info) < 1 { //convert fail
				t.Error("error")
				return
			}
			break
		}
		check_c += 1
		time.Sleep(time.Second)
	}
	if check_c < 1 {
		t.Error("error")
		return
	}
	// if true {
	// 	return
	// }

	//upload sample file
	res_2, err := DoUpF("../../ffcm/xx.mp4", "", "xxa", "x,y,z", folder.Id, "desc", 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if res_2.IntVal("added") != 0 {
		t.Error(util.S2Json(res_2))
		return
	}
	if res_2.StrValP("/base/id") != fid {
		fmt.Println(util.S2Json(res_2), fid)
		t.Error("error")
		return
	}
	if res_2.StrValP("/base/sha") != sha {
		t.Error("error")
		return
	}
	if res_2.StrValP("/base/pub") != pub {
		t.Error("error")
		return
	}
	if res_2.StrValP("/url") != url {
		t.Error("error")
		return
	}
	//
	//test download file
	err = DoPubDown(pub, "tmp/xxk.123")
	if err != nil {
		t.Error(err.Error())
		return
	}
	tsha, err := util.Sha1("tmp/xxk.123")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if sha != tsha {
		t.Error("error")
		return
	}
	//
	//test download extern file
	err = DoPubDown(pub+"/V_pc/0.mp4", "tmp/xxk.124")
	if err != nil {
		t.Error(err.Error())
		return
	}
	tsha1, err := util.Sha1("out/" + strings.Replace(path, ".", "_pc.", -1))
	if err != nil {
		t.Error(err.Error())
		return
	}
	tsha2, err := util.Sha1("tmp/xxk.124")
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(tsha, tsha1, tsha2)
	if tsha1 != tsha2 {
		t.Error("error")
		return
	}
	//
	//test download extern file 2
	err = DoPubDown(pub+"/V_json/0.mp4", "tmp/xxk.124_0")
	if err != nil {
		t.Error(err.Error())
		return
	}
	tsha1, err = util.Sha1("out/" + strings.Replace(path, ".", "_js.", -1))
	if err != nil {
		t.Error(err.Error())
		return
	}
	tsha2, err = util.Sha1("tmp/xxk.124_0")
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(tsha, tsha1, tsha2)
	if tsha1 != tsha2 {
		t.Error("error")
		return
	}
	//
	//test download file by fid
	err = DoFileDown(fid, "", "", 0, "tmp/xxk.125")
	if err != nil {
		t.Error(err.Error())
		return
	}
	tsha, err = util.Sha1("tmp/xxk.125")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if sha != tsha {
		t.Error("error")
		return
	}
	//
	//test donwload file by mark
	err = DoFileDown("", "xxa", "", 0, "tmp/xxk.126")
	if err != nil {
		t.Error(err.Error())
		return
	}
	tsha, err = util.Sha1("tmp/xxk.126")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if sha != tsha {
		t.Error("error")
		return
	}
	//
	//
	//test upload by base64
	data, err := ReadBase64("run_ff.sh")
	if err != nil {
		t.Error(err.Error())
		return
	}
	res, err = DoUpBase64(data, "text/plain", "run_ff.sh", "123", "", "", "", 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	var pub2 = res.StrValP("/base/pub")
	sha, _ = util.Sha1("run_ff.sh")
	if sha != res.StrValP("/base/sha") {
		t.Error("error")
		return
	}
	//
	//test not task
	res, err = DoInfo(fid, "", "", "")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if fid != res.StrValP("/base/id") {
		t.Error("error")
		return
	}
	fmt.Println(util.S2Json(res))
	if res.Val("task") != nil {
		t.Error("error")
		return
	}
	//
	//upload fail with mark exist
	_, err = DoUpF("run_ff.sh", "", "xxa", "x,y,z", folder.Id, "desc", 1)
	if err == nil {
		t.Error("error")
		return
	}
	//
	//
	//test error
	//
	//test upload error
	res, _ = util.HPostF2(fmt.Sprintf("%v/usr/api/uload?pub=sds&%v", SrvAddr(), SrvArgs()), nil, "file", "run_ff.sh")
	if res.IntVal("code") == 0 {
		t.Error("error")
		return
	}
	// fmt.Println(util.S2Json(res))
	res, _ = util.HPostF2(fmt.Sprintf("%v/usr/api/uload?%v", SrvAddr(), SrvArgs()), nil, "filex", "run_ff.sh")
	if res.IntVal("code") == 0 {
		t.Error("error")
		return
	}
	// fmt.Println(util.S2Json(res))
	_, res, _ = util.HPostN2(fmt.Sprintf("%v/usr/api/uload?base64=1&%v", SrvAddr(), SrvArgs()), "text/plain", bytes.NewBufferString("kkjfs"))
	if res.IntVal("code") == 0 {
		t.Error("error")
		return
	}
	//
	_, res, _ = util.HPostN2(fmt.Sprintf("%v/usr/api/uload?base64=1&name=xx.txt&%v", SrvAddr(), SrvArgs()), "", bytes.NewBufferString("kkjfs"))
	if res.IntVal("code") == 0 {
		t.Error("error")
		return
	}
	//
	tsh.Base.(*FBaseImpl).Base = "/xxx/"
	_, res, _ = util.HPostN2(fmt.Sprintf("%v/usr/api/uload?base64=1&name=xx.txt&%v", SrvAddr(), SrvArgs()), "text/plain", bytes.NewBufferString("kkjfs"))
	if res.IntVal("code") == 0 {
		t.Error("error")
		return
	}
	tsh.Base.(*FBaseImpl).Base = "."
	//
	// ffcm.SRV.Db.(*dtm.MemH).Errs["Find"] = util.Err("mock error")
	// res, _ = DoInfo(fid, "", "", "")
	// if res.Val("err") == nil {
	// 	t.Error("error")
	// 	return
	// }
	// ffcm.SRV.Db.(*dtm.MemH).Errs["Find"] = nil
	//
	res, err = DoInfo("", "", "", "")
	if err == nil {
		t.Error("error")
		return
	}
	//
	res, err = DoInfo("xdsds", "", "", "")
	if err == nil {
		t.Error("error")
		return
	}
	//
	gfsdb.UpdateF(fid, bson.M{"exec": gfsdb.ES_RUNNING})
	res, err = DoInfo(fid, "", "", "")
	if err != nil {
		t.Error("error")
		return
	}
	if res.Val("err") == nil {
		t.Error("error")
		return
	}
	gfsdb.UpdateF(fid, bson.M{"exec": gfsdb.ES_DONE})
	//
	err = DoFileDown("", "", "", 0, "kksfsd.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = DoFileDown("", "", "", -100, "kksfsd.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = DoFileDown("xds", "", "", 1, "kksfsd.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = DoPubDown("", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = DoPubDown("/kfsd/sd", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = DoPubDown("/kfsd", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = DoPubDown(pub+"/sfsd/xx.mp4", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = DoPubDown(pub+"/sfsd", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = DoPubDown(pub2+"/sfsd/", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = DoPubDown(pub+"/V_pc/100000", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = DoPubDown(pub+"/V_json/100000", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	//
	tsh.AddSender("xxx", tsh.SenderL["V_pc"])
	err = DoPubDown(pub+"/xxx/100000", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	delete(tsh.SenderL, "xxx")
	//
	tsh.AddSender("xxx", tsh.SenderL["V_json"])
	err = DoPubDown(pub+"/xxx/100000", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	delete(tsh.SenderL, "xxx")
	//
	tmgo.Mock = true
	//
	tmgo.SetMckC("Query-Apply", 0)
	_, err = DoUpF("../../ffcm/xx.mp4", "xx.mp4", "xxa", "x,y,z", folder.Id, "desc", 1)
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	tmgo.SetMckC("Query-Apply", 0)
	_, err = DoUpBase64(data, "text/plain", "run_ff.sh", "123", "", "", "", 1)
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	tmgo.SetMckC("Query-Apply", 1)
	_, err = DoUpF("../../ffcm/xx.mp4", "xx.mp4", "xxa", "x,y,z", folder.Id, "desc", 1)
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	tmgo.SetMckC("Query-One", 0)
	_, err = DoUpF("../../ffcm/xx.mp4", "xx.mp4", "xxa", "x,y,z", folder.Id, "desc", 1)
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	tmgo.SetMckC("Query-Apply", 2)
	_, err = DoUpF("../../ffcm/xx.mp4", "xx.mp4", "xxa", "x,y,z", folder.Id, "desc", 1)
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	tmgo.SetMckC("Query-One", 0)
	err = DoPubDown("/kfsd/sd", "kjdsf.xx")
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	//test address error
	SrvAddr = func() string {
		return "http://127.0.0.1:2334"
	}
	DoInfo("fid", "sha", "md5", "mark")
	DoFileDown("fid", "mark", "etype", 0, "path")
	DoPubDown("pub", "path")
	DoUpBase64("nil", "ctype", "name", "mark", "tags", "folder", "desc", 1)
	DoUpF("file", "name", "mark", "tags", "folder", "desc", 1)
	//
	fmt.Println("test done...")
}

func TestFBaseImpl(t *testing.T) {
	var fb = NewFBaseImpl(".", "")
	var fcfg = util.NewFcfg3()
	fcfg.InitWithFilePath2("../gfs_s.properties", true)
	if fb.ParseSubs(fcfg, "Subs") != 4 {
		t.Error("error")
		return
	}
	var ts = httptest.NewServer(func(hs *routing.HTTPSession) routing.HResult {
		var file = ""
		//
		_, file = fb.NewFile(hs, "xxx.doc")
		if !strings.HasSuffix(file, "_doc.doc") {
			t.Errorf("error-%v", file)
			return routing.HRES_RETURN
		}
		//
		hs.SetVal("sub", "abc")
		_, file = fb.NewFile(hs, "xxx.doc")
		if !strings.HasSuffix(file, "_xyz.doc") {
			t.Errorf("error-%v", file)
			return routing.HRES_RETURN
		}
		return routing.HRES_RETURN
	})
	ts.G("")

}

func TestUpDownErr(t *testing.T) {
	//
	//
	var fcfg = util.NewFcfg3()
	fcfg.InitWithData(`
[loc]
sender_l=Default
[Default]
sender=
		`)
	var _, err = NewFSH2(fcfg)
	if err == nil {
		t.Error("error")
		return
	}
	//
	//
	fcfg = util.NewFcfg3()
	fcfg.InitWithData(`
[loc]
sender_l=Default
[Default]
sender=dekk
		`)
	_, err = NewFSH2(fcfg)
	if err == nil {
		t.Error("error")
		return
	}
	//
	//
	_, err = ReadBase64("kkss")
	if err == nil {
		t.Error("error")
		return
	}
}

func TestXX(t *testing.T) {
	fmt.Println(strconv.ParseUint("0777", 8, 32))
	fmt.Println(uint32(os.ModePerm))
}
