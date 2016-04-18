package gfs

import (
	"github.com/Centny/gfs/gfsapi"
	"github.com/Centny/gwf/util"
	"testing"
	"time"
)

func TestSrv(t *testing.T) {
	go func() {
		fcfg := util.NewFcfg3()
		fcfg.InitWithFilePath2("gfs_s.properties", true)
		fcfg.Print()
		RunGFS_S(fcfg)
		panic("done...")
	}()
	time.Sleep(time.Second)
	go func() {
		fcfg := util.NewFcfg3()
		fcfg.InitWithFilePath2("gfs_c.properties", true)
		fcfg.Print()
		RunGFS_C(fcfg)
		panic("done...")
	}()
	time.Sleep(time.Second)
	fcfg := util.NewFcfg3()
	err := RunGFS_S(fcfg)
	if err == nil {
		t.Error("error")
		return
	}
	fcfg.SetVal("db_con", "cny:123@loc.m:27017/cny")
	fcfg.SetVal("db_name", "cny")
	err = RunGFS_S(fcfg)
	if err == nil {
		t.Error("error")
		return
	}
	fcfg.SetVal("db_con", "cny:123@loc.m:27017/cny")
	fcfg.SetVal("db_name", "cny")
	fcfg.SetVal("sender_l", "sdfs")
	err = RunGFS_S(fcfg)
	if err == nil {
		t.Error("error")
		return
	}
	gfsapi.SrvAddr = func() string {
		return "http://127.0.0.1:2325"
	}
	_, err = gfsapi.DoUpF("gfs_test.go", "", "maxxrk", "", "", "", 1, 1)
	if err != nil {
		t.Error("error")
		return
	}
}
