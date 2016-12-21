package gfsdb

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/Centny/dbm/mgo"
	"github.com/Centny/ffcm"
	_ "github.com/Centny/gfs/test"
	"github.com/Centny/gwf/netw/dtm"
	"github.com/Centny/gwf/tutil"
	"github.com/Centny/gwf/util"
	tmgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestF(t *testing.T) {
	runtime.GOMAXPROCS(util.CPU())
	mgo.C(CN_F).RemoveAll(nil)
	mgo.C(CN_MARK).RemoveAll(nil)
	var do_f = func(i int) int {
		var rt = &F{
			Path: "xxx",
			SHA:  "abc",
			MD5:  "xyz",
			Pub:  "/s",
		}
		var updated, err = FOI_F(rt)
		if err != nil {
			t.Error(err.Error())
			return 0
		}
		// fmt.Println(rt.Id)
		mk, err := FOI_Mark("jjk0", rt.Id)
		if err != nil {
			t.Error(err.Error())
			return 0
		}
		if mk.Fid != rt.Id {
			fmt.Println(mk.Fid, mk.Id)
			t.Error("error")
			return 0
		}
		return updated
	}
	var updated = do_f(0)
	if updated < 1 {
		t.Error("error")
		return
	}
	used, _ := tutil.DoPerf(100, "", func(i int) {
		do_f(i)
	})
	tc, err := CountF()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if tc != 1 {
		t.Error("error")
		return
	}
	fmt.Printf("done with used(%vms),per(%vms)\n", used, used/100)
	rt, err := FindHashF("abc", "xyz")
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = FindF(rt.Id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = FindMarkF("jjk0")
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = FindHashF("", "")
	if err == nil {
		t.Error("error")
		return
	}
	mk, err := FOI_Mark("jjk0", "xxds")
	if mk.Fid != rt.Id {
		t.Error("error")
		return
	}
	_, err = FOI_Mark("kjj", rt.Id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = FindMarkF("kjj")
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = FindPubF("/s")
	if err != nil {
		t.Error(err.Error())
		return
	}
	//
	fs, err := ListF([]string{rt.Id})
	if err != nil {
		t.Error(err)
		return
	}
	if len(fs) < 1 {
		t.Error("error")
		return
	}
	//
	fs, err = ListHashF([]string{rt.SHA}, []string{rt.MD5})
	if err != nil {
		t.Error(err)
		return
	}
	if len(fs) < 1 {
		t.Error("error")
		return
	}
	fs, err = ListHashF(nil, nil)
	if err == nil {
		t.Error("error")
		return
	}
	//
	fs, err = ListMarkF([]string{"jjk0"})
	if err != nil {
		t.Error(err)
		return
	}
	if len(fs) < 1 {
		t.Error("error")
		return
	}
	//
	fs, err = ListPubF([]string{"/s"})
	if err != nil {
		t.Error(err)
		return
	}
	if len(fs) < 1 {
		t.Error("error")
		return
	}
	//
	_, err = FOI_F(&F{})
	if err == nil {
		t.Error("error")
		return
	}
	//
	_, err = FOI_F(&F{
		Path: "sdfd",
	})
	if err == nil {
		t.Error("error")
		return
	}
	//
	tmgo.Mock = true
	tmgo.SetMckC("Query-Apply", 0)
	_, err = FOI_F(&F{
		Path: "sfkdf",
		SHA:  "abc",
		MD5:  "xyz",
	})
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	tmgo.SetMckC("Query-One", 0)
	_, err = FindMarkF("kjj")
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	tmgo.SetMckC("Query-All", 0)
	fs, err = ListPubF([]string{"/s"})
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	tmgo.SetMckC("Query-All", 0)
	fs, err = ListF([]string{rt.Id})
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	tmgo.SetMckC("Query-All", 0)
	fs, err = ListHashF([]string{rt.SHA}, []string{rt.MD5})
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
	tmgo.SetMckC("Query-All", 0)
	fs, err = ListMarkF([]string{"jjk0"})
	if err == nil {
		t.Error("error")
		return
	}
	tmgo.ClearMock()
	//
}

func test_img(t *testing.T) {
	var rt = &F{
		Path: "xx.jpg",
		SHA:  "abcsd",
		MD5:  "xydfsfz",
		EXT:  ".jpg",
	}
	_, err := FOI_F(rt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	for {
		rt, err = FindF(rt.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println("xxb->waiting result...")
		if len(rt.Info) > 0 {
			break
		}
		time.Sleep(time.Second)
	}
}
func TestFFCM(t *testing.T) {
	runtime.GOMAXPROCS(util.CPU())
	mgo.C(CN_F).RemoveAll(nil)
	ffcm.StartTest("../gfs_s.properties", "../gfs_c.properties", dtm.MemDbc, NewFFCM_H())
	time.Sleep(3 * time.Second)
	fmt.Println(ffcm.SRV)
	os.RemoveAll("tmp")
	///
	test_img(t)
	var rt = &F{
		Path: "xx.mp4",
		SHA:  "abc",
		MD5:  "xyz",
		EXT:  ".mp4",
	}
	ffcm.SRV.Db.(*dtm.MemH).Errs["Add"] = util.Err("mock error")
	var _, err = FOI_F(&F{
		Path: "xxkjk.mp4",
		SHA:  "abcsd",
		MD5:  "xyzfd",
	})
	if err != nil {
		t.Error(err)
		return
	}
	ffcm.SRV.Db.(*dtm.MemH).Errs["Add"] = nil
	os.RemoveAll("tmp")

	//
	_, err = FOI_F(rt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = FOI_F(&F{
		Path: "XXXX",
		SHA:  "abcx",
		MD5:  "xyzx",
	})
	if err != nil {
		t.Error(err.Error())
		return
	}
	// if true {
	// 	return
	// }
	for {
		rt, err = FindF(rt.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println("xxb->waiting result...")
		if len(rt.Info) > 0 {
			break
		}
		time.Sleep(time.Second)
	}
	fmt.Println("result->", util.S2Json(rt.Info))
	fmt.Println(rt.Id)
	//
	//
	//
	//
	fmt.Println("\n\n\n<--------------------------------------------->\n\n\n")
	mgo.C(CN_F).RemoveAll(nil)
	mgo.C(CN_FILE).RemoveAll(nil)
	_, err = FindF(rt.Id)
	if err == nil {
		t.Error("error")
		return
	}
	fmt.Printf("err->%v\n", err)
	//
	os.Chmod("tmp", 0)
	rt.Info = nil
	_, err = FOI_F(rt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	rt, err = FindF(rt.Id)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println("waiting result...")
	if len(rt.Info) > 0 {
		fmt.Println(util.S2Json(rt.Info))
		t.Error("error")
		return
	}
	time.Sleep(100 * time.Millisecond)
	for {
		task, err := ffcm.SRV.Db.Find(rt.Id)
		if err != nil || task == nil {
			t.Error(err)
			return
		}
		fmt.Println("xxa_0->waiting error...")
		if task.Status == "COV_ERR" {
			break
		}
		time.Sleep(time.Second)
	}
	os.RemoveAll("tmp")
	fmt.Println("xxx->a")
	ffcm.SRV.Db.Del(&dtm.Task{Id: rt.Id})
	_, err = SyncTask([]string{".mp4"}, []string{"abc"}, 100)
	if err != nil {
		t.Error(err.Error())
		return
	}

	for {
		rt, err = FindF(rt.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println("xx_001->waiting result...", rt.Id)
		if len(rt.Info) > 0 {
			break
		}
		time.Sleep(time.Second)
	}
	fmt.Println("result->", util.S2Json(rt.Info))
	//
	//
	//test verify
	fmt.Println("test verify->\n\n\n")
	total, fail, err := VerifyVideo(".", "sdata_o", []string{".mp4"}, nil, []string{"abc"})
	if err != nil || total < 1 || fail > 0 {
		fmt.Println(total, fail, err)
		t.Error("error")
		return
	}
	os.RemoveAll("sdata_o")
	total, fail, err = VerifyVideo(".", "sdata_o", []string{".mp4"}, nil, []string{"abc"})
	if err != nil || total < 1 || fail < 1 {
		fmt.Println(total, fail, err)
		t.Error("error")
		return
	}
	for {
		rt, err = FindF(rt.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println("xx_001->waiting result...", rt.Id)
		if len(rt.Info) > 0 {
			break
		}
		time.Sleep(time.Second)
	}
	fmt.Println("\n\n\n")
	//
	//
	mgo.C(CN_F).RemoveAll(nil)
	mgo.C(CN_FILE).RemoveAll(nil)
	_, err = FindF(rt.Id)
	if err == nil {
		t.Error("error")
		return
	}
	total, err = SyncAllTask([]string{".mp4"})
	if err != nil || total > 0 {
		fmt.Println(err, total)
		t.Error("error")
		return
	}
	mgo.C("ffcm_task").Insert(bson.M{"_id": rt.Id})
	total, err = SyncAllTask([]string{".mp4"})
	if err != nil || total > 0 {
		fmt.Println(err, total)
		t.Error("error")
		return
	}
}

func TestImg(t *testing.T) {

}

func TestReg(t *testing.T) {
	fmt.Println(regexp.MustCompile("^[^X]+[^K]+.*$").MatchString("XXX"))
}

func TestFFCM_H_err(t *testing.T) {
	var ffcm = NewFFCM_H()
	//
	var err = ffcm.OnDone(nil, &dtm.Task{
		Proc: map[string]*dtm.Proc{
			"xx": &dtm.Proc{},
		},
	})
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = ffcm.OnDone(nil, &dtm.Task{
		Proc: map[string]*dtm.Proc{
			"xx": &dtm.Proc{
				Res: "sss",
			},
		},
	})
	if err == nil {
		t.Error("error")
		return
	}
	//
	err = ffcm.OnDone(nil, &dtm.Task{
		Proc: map[string]*dtm.Proc{
			"xx": &dtm.Proc{
				Res: util.Map{},
			},
		},
	})
	if err == nil {
		t.Error("error")
		return
	}
	//
	ffcm.OnStart(nil, &dtm.Task{
		Id: "xkssdf",
	})
	//
	update_exec(&F{
		Id: "sss",
	}, ES_ERROR)
}

func TestMapValV(t *testing.T) {
	var mv, ok = MapVal(map[string]interface{}{
		"xa": util.Map{
			"a1": 1,
			"b1": 2,
		},
		"xb": bson.M{
			"a2": "1",
			"b2": "2",
		},
		"xc": map[string]interface{}{
			"a2": "1",
			"b2": "2",
		},
	})
	if !ok {
		t.Error("error")
		return
	}
	fmt.Println(util.S2Json(mv))
}
