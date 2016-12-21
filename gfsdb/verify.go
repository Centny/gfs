package gfsdb

import "gopkg.in/mgo.v2/bson"
import "github.com/Centny/gwf/log"
import "github.com/Centny/ffcm"
import "path/filepath"
import "fmt"
import "github.com/Centny/gwf/util"

func VerifyVideo(diri, diro string, exts, ignore []string) (total, fail int, err error) {
	var query = bson.M{
		"exec": bson.M{
			"$in": []string{ES_DONE},
		},
	}
	if len(exts) > 0 {
		query["ext"] = bson.M{
			"$in": exts,
		}
	}
	if len(ignore) > 0 {
		query["_id"] = bson.M{
			"$nin": ignore,
		}
	}
	fmt.Println(util.S2Json(query))
	total, err = C(CN_F).Find(query).Count()
	if err != nil {
		return
	}
	var done = 0
	var fs []*F
	for {
		err = C(CN_F).Find(query).Sort("_id").Skip(done).Limit(1000).All(&fs)
		if err != nil {
			return
		}
		if len(fs) < 1 {
			break
		}
		log.D("VerifyVideo start verify video process %v/%v", done+len(fs), total)
		for _, rf := range fs {
			err = VerifyVideoF(diri, diro, rf)
			if err == nil {
				continue
			}
			fail++
			log.W("VerifyVideo %v, will mark file(%v) to exec running and redo", err, rf.Id)
			err = UpdateExecF(rf.Id, ES_RUNNING)
			if err != nil {
				return
			}
			err = DoAddTask(rf)
			if err != nil {
				return
			}
		}
		done += len(fs)
	}
	log.D("VerifyVideo verify video done with total(%v),fail(%v)", total, fail)
	return
}

func VerifyVideoF(diri, diro string, rf *F) error {
	var pc = rf.Info.StrValP("/V_pc/text")
	if len(pc) > 0 {
		err := ffcm.VerifyVideo(filepath.Join(diri, rf.Path), filepath.Join(diro, pc))
		if err != nil {
			return err
		}
	}
	var phone = rf.Info.StrValP("/V_phone/text")
	if len(phone) > 0 {
		err := ffcm.VerifyVideo(filepath.Join(diri, rf.Path), filepath.Join(diro, phone))
		if err != nil {
			return err
		}
	}
	return nil
}
