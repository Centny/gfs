package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Centny/gfs"
	"github.com/Centny/gfs/gfsapi"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/smartio"
	"github.com/Centny/gwf/util"
)

func usage() {
	fmt.Println(`Usage:
	gfs -c <configure file>					run client mode by configure file
	gfs -s <configure file>					run server on database store mode by configure file.
	gfs -u <server address> <upload file> <mark> <other argumnets>
	gfs -verify <configure file>					run verify all file.
		`)
}

var ef = os.Exit

func main() {
	if len(os.Args) < 2 {
		usage()
		ef(1)
		return
	}
	switch os.Args[1] {
	case "-c":
		var cfg = "conf/gfs_c.properties"
		if len(os.Args) > 2 {
			cfg = os.Args[2]
		}
		var fcfg_s = util.NewFcfg3()
		fcfg_s.InitWithFilePath2(cfg, true)
		fcfg_s.Print()
		redirect_l(fcfg_s)
		fmt.Println(gfs.RunGFS_C(fcfg_s))
		smartio.ResetStd()
		time.Sleep(time.Second)
	case "-s":
		var cfg = "conf/gfs_s.properties"
		if len(os.Args) > 2 {
			cfg = os.Args[2]
		}
		var fcfg_s = util.NewFcfg3()
		fcfg_s.InitWithFilePath2(cfg, true)
		fcfg_s.Print()
		gfsapi.ShowLog = true
		redirect_l(fcfg_s)
		fmt.Println(gfs.RunGFS_S(fcfg_s))
		smartio.ResetStd()
		time.Sleep(time.Second)
	case "-u":
		if len(os.Args) < 4 {
			usage()
			ef(1)
			return
		}
		gfsapi.SrvAddr = func() string {
			return os.Args[2]
		}
		gfsapi.SrvArgs = func() string {
			if len(os.Args) > 5 {
				return os.Args[5]
			} else {
				return ""
			}
		}
		var mark = ""
		if len(os.Args) > 4 {
			mark = os.Args[4]
		}
		res, err := gfsapi.DoUpF(os.Args[3], "", mark, "", "", "", 1, 1)
		if err == nil {
			fmt.Println(util.S2Json(res))
		} else {
			fmt.Println(err)
		}
	case "-verify":
		var cfg = "conf/gfs_s.properties"
		if len(os.Args) > 2 {
			cfg = os.Args[2]
		}
		var fcfg_s = util.NewFcfg3()
		fcfg_s.InitWithFilePath2(cfg, true)
	default:
		usage()
		ef(1)
	}
}

func redirect_l(fcfg *util.Fcfg) {
	var out_l = fcfg.Val2("out_l", "")
	var err_l = fcfg.Val2("err_l", "")
	log.RedirectV(out_l, err_l, false)
}
