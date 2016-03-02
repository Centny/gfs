package gfs

import (
	"github.com/Centny/dbm/mgo"
	"github.com/Centny/ffcm"
	"github.com/Centny/ffcm/mdb"
	"github.com/Centny/gfs/gfsapi"
	"github.com/Centny/gfs/gfsdb"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/routing/filter"
	"github.com/Centny/gwf/util"
	"github.com/Centny/gwf/log"
)

func RunGFS_C(fcfg *util.Fcfg) error {
	return ffcm.RunFFCM_Cv(fcfg)
}

func RunGFS_S(fcfg *util.Fcfg) error {
	err := mgo.AddDefault(fcfg.Val2("db_con", ""), fcfg.Val2("db_name", ""))
	if err != nil {
		return err
	}
	gfsdb.C = mgo.C
	fsh, err := gfsapi.NewFSH2(fcfg)
	if err != nil {
		return err
	}
	routing.Shared.HFilterFunc("^/usr/api/uload(\\?.*)?$", filter.ParseQuery)
	routing.Shared.HFilterFunc("^/usr/.*$", func(hs *routing.HTTPSession) routing.HResult {
		hs.SetVal("uid", fcfg.Val2("uid", "sys"))
		return routing.HRES_CONTINUE
	})
	err = ffcm.InitDtcmS(fcfg, mdb.DefaultDbc, gfsdb.NewFFCM_H())
	if err != nil {
		return err
	}
	ffcm.SRV.Hand("", routing.Shared)
	fsh.Hand("", routing.Shared)
	routing.Shared.Print()
	var listen = fcfg.Val("listen")
	log.D("listen web server on %v", listen)
	return routing.ListenAndServe(listen)
}
