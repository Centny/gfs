package main

import (
	"fmt"
	"os"

	"runtime"

	"github.com/Centny/dbm/mgo"
	"github.com/Centny/gfs/gfsdb"
	"github.com/Centny/gwf/util"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:./update_mark <databases connection>")
		return
	}
	runtime.GOMAXPROCS(util.CPU())
	mgo.AddDefault2(os.Args[1])
	gfsdb.C = mgo.C
	total, err := gfsdb.CountF()
	if err != nil {
		fmt.Printf("fail with %v\n", err)
		os.Exit(1)
	}
	count := total/1000 + 1
	for i := 0; i < count; i++ {
		//list current file
		var files = []util.Map{}
		err = gfsdb.C(gfsdb.CN_F).Find(nil).Select(bson.M{"_id": 1}).Skip(i * 1000).Limit(1000).All(&files)
		if err != nil {
			fmt.Printf("fail with %v\n", err)
			os.Exit(1)
		}
		if len(files) < 1 {
			continue
		}
		var idsMap = util.Map{}
		var idsAry = []string{}
		for _, file := range files {
			var fid = file.StrVal("_id")
			idsMap[fid] = 1
			idsAry = append(idsAry, fid)
		}
		//filter marked
		var markAry = []util.Map{}
		err = gfsdb.C(gfsdb.CN_MARK).Find(
			bson.M{
				"_id": bson.M{
					"$in": idsAry,
				}},
		).Select(bson.M{"_id": 1}).All(&markAry)
		if err != nil {
			fmt.Printf("fail with %v\n", err)
			os.Exit(1)
		}
		for _, mark := range markAry {
			delete(idsMap, mark.StrVal("_id"))
		}
		if len(idsMap) < 1 {
			continue
		}
		//insert not marked
		var markList = []interface{}{}
		for fid := range idsMap {
			markList = append(markList, &gfsdb.Mark{
				Id:  fid,
				Fid: fid,
			})
		}
		err = gfsdb.C(gfsdb.CN_MARK).Insert(markList...)
		if err != nil {
			fmt.Printf("fail with %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("update_mark %v/%v success\n", i*1000+len(files), total)
	}
	fmt.Printf("update_mark all done\n")
}
