package gfsdb

import (
	"github.com/Centny/ffcm"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/netw/dtm"
	"github.com/Centny/gwf/util"
	tmgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"path/filepath"
	"reflect"
)

func FOI_F(rf *F) (int, error) {
	if len(rf.Path) < 1 {
		return 0, util.Err("FOI_F the F.path is empty ")
	}
	if len(rf.SHA) < 1 && len(rf.MD5) < 1 {
		return 0, util.Err("FOI_F the F.sha/F.md5 is empty ")
	}
	rf.Id = bson.NewObjectId().Hex()
	rf.Exec = ES_NONE
	var res, err = C(CN_F).Find(bson.M{
		"$or": []bson.M{
			bson.M{"sha": rf.SHA},
			bson.M{"md5": rf.MD5},
		},
	}).Apply(tmgo.Change{
		Update: bson.M{
			"$setOnInsert": rf,
		},
		Upsert:    true,
		ReturnNew: true,
	}, rf)
	if err != nil {
		return 0, err
	}
	if res.UpsertedId == nil || len(res.UpsertedId.(string)) < 1 {
		log.D("FOI_F adding really file(%v) on path(%v) success with file already exist", rf.Id, rf.Path)
		return 0, nil
	}
	if ffcm.SRV == nil {
		log.D("FOI_F adding really file(%v) on path(%v) success with ffcm server is not running", rf.Id, rf.Path)
		return 1, nil
	}
	go func() {
		var out = CreateOutPath(rf)
		err = ffcm.SRV.AddTaskV(rf.Id, rf.Id, rf.Path, out, filepath.Ext(rf.Path))
		if err == nil {
			log.D("FOI_F adding really file(%v) on path(%v) success with ffcm task out path(%v)", rf.Id, rf.Path, out)
		} else if dtm.IsNotMatchedErr(err) {
			log.D("FOI_F adding really file(%v) on path(%v) success with not ffcm task matched", rf.Id, rf.Path)
		} else {
			log.E("FOI_F adding really file(%v) on path(%v) success, but add ffcm task to out path(%v) error->%v, will mark it to exec error", rf.Id, rf.Path, out, err)
			update_exec(rf)
		}
	}()
	return res.Updated, nil
}

func update_exec(rf *F) error {
	var err = UpdateExecF(rf.Id, ES_ERROR)
	if err == nil {
		log.D("FOI_F mark really file(%v) to exec error success")
	} else {
		log.E("FOI_F mark really file(%v) to exec error fail with error->%v", rf.Id, err)
	}
	return err
}

func FOI_Mark(mark, fid string) (*Mark, error) {
	var mk = &Mark{}
	var _, err = C(CN_MARK).Find(bson.M{
		"_id": mark,
	}).Apply(tmgo.Change{
		Update: bson.M{
			"$setOnInsert": bson.M{"fid": fid},
		},
		Upsert:    true,
		ReturnNew: true,
	}, mk)
	return mk, err
}

func CountF() (int, error) {
	return C(CN_F).Count()
}

func UpdateF(id string, set bson.M) error {
	return C(CN_F).UpdateId(id, bson.M{"$set": set})
}

func FindFv(query bson.M) (*F, error) {
	var f = &F{}
	return f, C(CN_F).Find(query).One(f)
}

func FindF(id string) (*F, error) {
	return FindFv(bson.M{"_id": id})
}

func FindHashF(sha, md5 string) (*F, error) {
	var query = []bson.M{}
	if len(sha) > 0 {
		query = append(query, bson.M{"sha": sha})
	}
	if len(md5) > 0 {
		query = append(query, bson.M{"md5": md5})
	}
	if len(query) < 1 {
		return nil, util.Err("md5 or sha must not be empty")
	}
	return FindFv(bson.M{
		"$or": query,
	})
}

func FindMarkF(mark string) (*F, error) {
	var mk = &Mark{}
	var err = C(CN_MARK).FindId(mark).One(&mk)
	if err == nil {
		return FindF(mk.Fid)
	} else {
		return nil, err
	}
}

func FindPubF(pub string) (*F, error) {
	return FindFv(bson.M{
		"pub": pub,
	})
}

func UpdateExecF(id, es string) error {
	return C(CN_F).Update(
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"exec": es,
			},
		})
}

type FFCM_H struct {
}

func NewFFCM_H() *FFCM_H {
	return &FFCM_H{}
}
func (f *FFCM_H) OnStart(dtcm *dtm.DTCM_S, task *dtm.Task) {
	var err = UpdateF(task.Id, bson.M{"exec": ES_RUNNING})
	if err == nil {
		log.D("FFCM_H update task(%v) exec status to %v", task.Id, ES_RUNNING)
	} else {
		log.E("FFCM_H update task(%v) exec status to %v fail with error->%v", task.Id, ES_RUNNING, err)
	}
}
func (f *FFCM_H) ParseRes(task *dtm.Task, res util.Map) error {
	var err error
	for key, proc := range task.Proc {
		if proc.Res == nil {
			err = util.Err("parsing result fail on task(%v),proc(%v) with result is nil", task.Id, key)
			log.E("%v", err)
			return err
		}
		mv, ok := MapVal(proc.Res)
		if !ok {
			err = util.Err("parsing result(%v) as type(%v) to util.Map value fail on task(%v),proc(%v)",
				util.S2Json(proc.Res), reflect.ValueOf(proc.Res).Type().String(), task.Id, key)
			log.E("%v", err)
			return err
		}
		var data = mv.MapVal("data")
		if data == nil {
			err = util.Err("parsing data from result(%v) on task(%v),proc(%v) is nil", util.S2Json(mv), task.Id, key)
			log.E("%v", err)
			return err
		}
		delete(data, "src")
		res.SetVal(key, data)
	}
	return nil
}
func (f *FFCM_H) OnDone(dtcm *dtm.DTCM_S, task *dtm.Task) error {
	var info = util.Map{}
	var err = f.ParseRes(task, info)
	if err == nil {
		info["code"] = 0
	} else {
		info["code"] = 1
		info["info"] = task.Info
		info["error"] = err.Error()
	}
	return UpdateF(task.Id, bson.M{
		"info": info,
		"exec": ES_DONE,
	})
}

func MapVal(v interface{}) (util.Map, bool) {
	var res, ok = MapValV(v).(util.Map)
	return res, ok
}

func MapValV(v interface{}) interface{} {
	if bv, ok := v.(bson.M); ok {
		var res = util.Map{}
		for key, val := range bv {
			res[key] = MapValV(val)
		}
		return res
	} else if mv, ok := v.(map[string]interface{}); ok {
		var res = util.Map{}
		for key, val := range mv {
			res[key] = MapValV(val)
		}
		return res
	} else {
		return v
	}
}
