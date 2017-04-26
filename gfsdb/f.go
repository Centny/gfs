package gfsdb

import (
	"path/filepath"
	"reflect"

	"github.com/Centny/ffcm"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/netw/dtm"
	"github.com/Centny/gwf/util"
	tmgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var MockStartTaskErr int = 0
var ExternalExecSupported = []string{}

func IsExternalSupported(ext string) bool {
	for _, supported := range ExternalExecSupported {
		if ext == supported {
			return true
		}
	}
	return false
}

func FOI_F(rf *F) (int, error) {
	if len(rf.Path) < 1 {
		return 0, util.Err("FOI_F the F.path is empty ")
	}
	if len(rf.SHA) < 1 && len(rf.MD5) < 1 {
		return 0, util.Err("FOI_F the F.sha/F.md5 is empty ")
	}
	rf.Id = bson.NewObjectId().Hex()
	if IsExternalSupported(rf.EXT) ||
		(ffcm.SRV != nil && ffcm.SRV.MatchArgsV(rf.Id, rf.Id, rf.Path, "", rf.EXT)) {
		rf.Exec = ES_RUNNING
	} else {
		rf.Exec = ES_NONE
	}
	var mv = rf.ToBsonM()
	delete(mv, "pub")
	var update = bson.M{
		"$setOnInsert": mv,
	}
	if len(rf.Pub) > 0 {
		update["$set"] = bson.M{
			"pub": rf.Pub,
		}
	}
	var res, err = C(CN_F).Find(bson.M{
		"$or": []bson.M{
			bson.M{"sha": rf.SHA},
			bson.M{"md5": rf.MD5},
		},
	}).Apply(tmgo.Change{
		Update:    update,
		Upsert:    true,
		ReturnNew: true,
	}, rf)
	if err != nil {
		return 0, err
	}
	log.D("FOI_F adding file success->%v", util.S2Json(rf))
	if res.UpsertedId == nil || len(res.UpsertedId.(string)) < 1 {
		log.D("FOI_F adding really file(%v) on path(%v) success with file already exist", rf.Id, rf.Path)
		return 0, nil
	}
	if ffcm.SRV == nil {
		log.D("FOI_F adding really file(%v) on path(%v) success with ffcm server is not running", rf.Id, rf.Path)
		return 1, nil
	}
	return DoAddTask(rf)
}

func DoAddTask(rf *F) (int, error) {
	if ffcm.SRV.MatchLocArgsV(rf.Id, rf.Id, rf.Path, "", filepath.Ext(rf.Path)) {
		log.D("FOI_F adding really file(%v) on path(%v) by doing local task", rf.Id, rf.Path)
		var out = CreateOutPath(rf)
		res, err := ffcm.SRV.RunLocTaskV(rf.Id, rf.Id, rf.Path, out, filepath.Ext(rf.Path))
		if err != nil {
			err = util.Err("FOI_F running local task fail with error->%v", err)
			log.E("%v", err)
			do_remove(rf.Id)
			return 0, err
		}
		if len(res) > 0 {
			if rf.Info == nil {
				rf.Info = util.Map{}
			}
			var setv = bson.M{}
			for key, _ := range res {
				data := res.MapValP("/" + key + "/data")
				if data == nil {
					err = util.Err("FOI_F running local task fail with error(result data is nil), the all data is %v", util.S2Json(res))
					log.E("%v", err)
					do_remove(rf.Id)
					return 0, err
				}
				rf.Info[key] = data
				setv["info."+key] = data
			}
			err = UpdateF(rf.Id, setv)
			if err != nil {
				err = util.Err("FOI_F running local task fail with update file error(%v)", err)
				log.E("%v", err)
				do_remove(rf.Id)
				return 0, err
			}
		}
	}
	go DoAddTaskRemote(rf)
	return 1, nil
}

func do_remove(id string) error {
	return C(CN_F).RemoveId(id)
}

func DoAddTaskRemote(rf *F) error {
	if ffcm.SRV == nil {
		log.W("start ffcm task fail with the server is not running")
		return nil
	}
	var out = CreateOutPath(rf)
	var err = util.Err("Mock Error")
	if MockStartTaskErr < 1 {
		err = ffcm.SRV.AddTaskV(rf.Id, rf.Id, rf.Path, out, filepath.Ext(rf.Path))
	}
	if err == nil {
		log.D("FOI_F adding really file(%v) on path(%v) success with ffcm task out path(%v)", rf.Id, rf.Path, out)
	} else if dtm.IsNotMatchedErr(err) {
		log.D("FOI_F adding really file(%v) on path(%v) success with not ffcm task matched", rf.Id, rf.Path)
		update_exec(rf, ES_NONE)
	} else {
		log.E("FOI_F adding really file(%v) on path(%v) success, but add ffcm task to out path(%v) error->%v, will mark it to exec ignore", rf.Id, rf.Path, out, err)
		update_exec(rf, ES_IGNORE)
	}
	return err
}

func update_exec(rf *F, exec string) error {
	var err = UpdateExecF(rf.Id, exec)
	if err == nil {
		log.D("FOI_F mark really file(%v) to exec error success", rf.Id)
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

func ListFv(query bson.M) ([]*F, error) {
	var fs = []*F{}
	var err = C(CN_F).Find(query).All(&fs)
	return fs, err
}

func FindF(id string) (*F, error) {
	return FindFv(bson.M{"_id": id})
}

func ListF(ids []string) ([]*F, error) {
	return ListFv(bson.M{"_id": bson.M{"$in": ids}})
}

func ListF_m(ids []string) (map[string]*F, error) {
	var fs, err = ListF(ids)
	if err != nil {
		return nil, err
	}
	var fs_m = map[string]*F{}
	for _, f := range fs {
		fs_m[f.Id] = f
	}
	return fs_m, nil
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

func ListHashF(sha, md5 []string) ([]*F, error) {
	var query = []bson.M{}
	if len(sha) > 0 {
		query = append(query, bson.M{"sha": bson.M{"$in": sha}})
	}
	if len(md5) > 0 {
		query = append(query, bson.M{"md5": bson.M{"$in": md5}})
	}
	if len(query) < 1 {
		return nil, util.Err("md5 or sha must not be empty")
	}
	return ListFv(bson.M{
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

func ListMarkF(mark []string) ([]*F, error) {
	var fs, _, err = ListMarkFv(mark)
	return fs, err
}

func ListMarkFv(mark []string) ([]*F, map[string]string, error) {
	var mk = []*Mark{}
	var err = C(CN_MARK).Find(bson.M{"_id": bson.M{"$in": mark}}).All(&mk)
	if err != nil {
		return nil, nil, err
	}
	var fids = []string{}
	var mfids = map[string]string{}
	for _, m := range mk {
		fids = append(fids, m.Fid)
		mfids[m.Id] = m.Fid
		// mfids[m.Fid] = m.Id
	}
	fs, err := ListF(fids)
	return fs, mfids, err
}

func FindPubF(pub string) (*F, error) {
	return FindFv(bson.M{
		"pub": pub,
	})
}

func ListPubF(pub []string) ([]*F, error) {
	return ListFv(bson.M{"pub": bson.M{"$in": pub}})
}
func ListShaF(sha []string) ([]*F, error) {
	return ListFv(bson.M{"sha": bson.M{"$in": sha}})
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

func UpdateVerifyF(id, es string) error {
	return C(CN_F).Update(
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"verify": es,
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
		if data.Exist("code") {
			data["code"] = int(data.IntVal("code"))
		}
		if data.Exist("count") {
			data["count"] = int(data.IntVal("count"))
		}
		res.SetVal(key, data)
	}
	return nil
}
func (f *FFCM_H) OnDone(dtcm *dtm.DTCM_S, task *dtm.Task) error {
	log.D("FFCM_H receive done for task(%v)", task.Id)
	var info = util.Map{}
	var setv = bson.M{}
	if task.Code == 0 {
		var err = f.ParseRes(task, info)
		if err == nil {
			info["code"] = 0
		} else {
			info["code"] = 1
			info["info"] = task.Info
			info["error"] = err.Error()
		}
		setv["exec"] = ES_DONE
		for key, val := range info {
			setv["info."+key] = val
		}
	} else {
		setv["exec"] = ES_ERROR
	}
	return UpdateF(task.Id, setv)
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

func ListTaskIds() ([]string, error) {
	var pipe = []bson.M{
		bson.M{
			"$group": bson.M{
				"_id": 0,
				"ids": bson.M{
					"$push": "$_id",
				},
			},
		},
	}
	var res []util.Map
	var err = C("ffcm_task").Pipe(pipe).All(&res)
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return res[0].AryStrVal("ids"), nil
	} else {
		return nil, nil
	}
}

func SyncTask(exts, ignore []string, limit int) (int, error) {
	var query = bson.M{
		"exec": bson.M{
			"$in": []string{ES_ERROR, ES_RUNNING},
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
	var pipe = []bson.M{
		bson.M{
			"$match": query,
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "ffcm_task",
				"localField":   "_id",
				"foreignField": "_id",
				"as":           "task",
			},
		},
		bson.M{
			"$match": bson.M{
				"task": bson.M{
					"$size": 0,
				},
			},
		},
		bson.M{
			"$limit": limit,
		},
	}
	// log.D("abc->%v", util.S2Json(pipe))
	var fs = []*F{}
	var err = C(CN_F).Pipe(pipe).All(&fs)
	if err != nil {
		log.E("SyncTask list file by exts(%v),ignore(%v) fail with error(%v), the query is \n%v\n", exts, ignore, err, util.S2Json(query))
		return 0, err
	}
	log.D("SyncTask list file by exts(%v),ignore(%v) success with %v found", exts, ignore, len(fs))
	for _, rf := range fs {
		_, err = DoAddTask(rf)
		if err != nil {
			ignore = append(ignore, rf.Id)
		}
	}
	return len(fs), nil
}

func SyncAllTask(exts []string) (total int, err error) {
	return SyncAllTaskV(exts, nil)
}

func SyncAllTaskV(exts, ignore []string) (total int, err error) {
	log.D("Do sync all task by exts(%v),ignore(%v)", exts, ignore)
	var matched = 0
	for {
		matched, err = SyncTask(exts, ignore, 100)
		if err != nil {
			return
		}
		if matched < 1 {
			break
		}
		total += matched
	}
	return
}
