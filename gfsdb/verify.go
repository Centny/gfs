package gfsdb

import "gopkg.in/mgo.v2/bson"
import "github.com/Centny/gwf/log"
import "github.com/Centny/ffcm"
import "path/filepath"

func VerifyVideo(diri, diro string, exts, ids, ignore []string) (total, fail int, err error) {
	var query = bson.M{
		"exec": bson.M{
			"$in": []string{ES_DONE},
		},
		"$or": []bson.M{
			{
				"verify": bson.M{
					"$exists": 0,
				},
			},
			{
				"verify": bson.M{
					"$in": []string{VS_REDO, VS_ERROR},
				},
			},
		},
	}
	if len(exts) > 0 {
		query["ext"] = bson.M{
			"$in": exts,
		}
	}
	if len(ids) > 0 {
		query["_id"] = bson.M{
			"$in": ids,
		}
	}
	if len(ignore) > 0 {
		query["_id"] = bson.M{
			"$nin": ignore,
		}
	}
	total, err = C(CN_F).Find(query).Count()
	if err != nil {
		return
	}
	var code = 0
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
			code, err = VerifyVideoF(diri, diro, rf)
			if err == nil {
				err = UpdateVerifyF(rf.Id, VS_VERIFIED)
				if err != nil {
					return
				}
				continue
			}
			switch code {
			case 1:
				fail++
				log.W("VerifyVideo %v, will mark file(%v) to zero", err, rf.Id)
				err = UpdateVerifyF(rf.Id, VS_ZERO)
				if err != nil {
					return
				}
			default:
				fail++
				log.W("VerifyVideo %v, will mark file(%v) to redo", err, rf.Id)
				err = UpdateVerifyF(rf.Id, VS_REDO)
				if err != nil {
					return
				}
				_, err = DoAddTask(rf)
				if err != nil {
					return
				}
			}

		}
		done += len(fs)
	}
	log.D("VerifyVideo verify video done with total(%v),fail(%v)", total, fail)
	return
}

func VerifyVideoF(diri, diro string, rf *F) (int, error) {
	var pc = rf.Info.StrValP("/V_pc/text")
	if len(pc) > 0 {
		code, err := ffcm.VerifyVideo(filepath.Join(diri, rf.Path), filepath.Join(diro, pc))
		if err != nil {
			return code, err
		}
	}
	var phone = rf.Info.StrValP("/V_phone/text")
	if len(phone) > 0 {
		code, err := ffcm.VerifyVideo(filepath.Join(diri, rf.Path), filepath.Join(diro, phone))
		if err != nil {
			return code, err
		}
	}
	return 0, nil
}
