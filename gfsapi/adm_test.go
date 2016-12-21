package gfsapi

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/Centny/gwf/util"
)

func TestVerify(t *testing.T) {
	ShowLog = true
	clearDb()
	runtime.GOMAXPROCS(util.CPU())
	os.RemoveAll("www")
	os.RemoveAll("out")
	os.RemoveAll("tmp")
	uid = "123"
	//
	//test upload file
	res, err := DoUpF("../../ffcm/xx.mp4", "", "xxa", "x,y,z", "", "desc", 1, 1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// fmt.Println(res)
	var fid = res.StrValP("/base/id")
	fmt.Println(fid)
	// //
	// //wait task done..
	var check_c = 0
	for {
		res, err = DoInfo(fid, "", "", "", "")
		if err != nil {
			t.Error(err.Error())
			return
		}
		// fmt.Println(util.S2Json(res))
		if res.StrValP("/base/exec") != "running" {
			var info = res.MapValP("/base/info")
			if info == nil || len(info) < 1 { //convert fail
				t.Error("error")
				fmt.Println(util.S2Json(res))
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
	fmt.Println("\n\n\n\n")
	//
	//test verify
	fmt.Println("test verify---->")
	os.RemoveAll("sdata_o")
	err = DoAdmVerify()
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)
	for {
		res, err = DoInfo(fid, "", "", "", "")
		if err != nil {
			t.Error(err.Error())
			return
		}
		// fmt.Println(util.S2Json(res))
		if res.StrValP("/base/exec") != "running" {
			var info = res.MapValP("/base/info")
			if info == nil || len(info) < 1 { //convert fail
				t.Error("error")
				fmt.Println(util.S2Json(res))
				return
			}
			break
		}
		check_c += 1
		time.Sleep(time.Second)
	}
	if check_c < 2 {
		t.Error("error")
		return
	}
	fmt.Println("\n\n\n\n")
	//
	//test redo
	fmt.Println("test redo---->")
	err = DoRedoTask(fid, "", "", "", "")
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)
	for {
		res, err = DoInfo(fid, "", "", "", "")
		if err != nil {
			t.Error(err.Error())
			return
		}
		// fmt.Println(util.S2Json(res))
		if res.StrValP("/base/exec") != "running" {
			var info = res.MapValP("/base/info")
			if info == nil || len(info) < 1 { //convert fail
				t.Error("error")
				fmt.Println(util.S2Json(res))
				return
			}
			break
		}
		check_c += 1
		time.Sleep(time.Second)
	}
	if check_c < 3 {
		t.Error("error")
		return
	}
}
