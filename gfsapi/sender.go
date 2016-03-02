package gfsapi

import (
	"fmt"
	"github.com/Centny/gfs/gfsdb"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/util"
	"net/http"
	"net/url"
	"strings"
)

type FSedner interface {
	String() string
	Send(hs *routing.HTTPSession, rf *gfsdb.F, etype string, dl bool, idx int) routing.HResult
}

// var SenderL = map[string]FSedner{}

// func AddSender(s FSedner) {
// 	SenderL[s.Type()] = s
// }

type DefaultSender struct {
	FH  http.Handler
	Pre string
}

func NewDefaultSender(fh http.Handler, pre string) *DefaultSender {
	return &DefaultSender{FH: fh, Pre: pre}
}
func NewDefaultSender2(dir, pre string) *DefaultSender {
	return NewDefaultSender(http.FileServer(http.Dir(dir)), pre)
}
func (d *DefaultSender) Send(hs *routing.HTTPSession, rf *gfsdb.F, etype string, dl bool, idx int) routing.HResult {
	hs.R.URL.Path = d.Pre + rf.Path
	return d.DoH(hs, rf, etype, dl, idx)
}
func (d *DefaultSender) DoH(hs *routing.HTTPSession, rf *gfsdb.F, etype string, dl bool, idx int) routing.HResult {
	if dl {
		var header = hs.W.Header()
		header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", url.QueryEscape(rf.Name)))
	}
	slog("DefaultSender do normal http file server to %v", hs.R.URL.Path)
	d.FH.ServeHTTP(hs.W, hs.R)
	return routing.HRES_RETURN
}
func (d *DefaultSender) String() string {
	return "DefaultSender"
}

type TextSender struct {
	*DefaultSender
}

func NewTextSender(sender *DefaultSender) *TextSender {
	return &TextSender{DefaultSender: sender}
}
func (t *TextSender) Send(hs *routing.HTTPSession, rf *gfsdb.F, etype string, dl bool, idx int) routing.HResult {
	var eval = rf.Info.MapVal(etype)
	if eval == nil {
		hs.W.WriteHeader(404)
		var msg = fmt.Sprintf("file(%v,%v) extern type(/info/%v) attribute is not exist", rf.Id, rf.Pub, etype)
		log.E("%v", msg)
		fmt.Fprintf(hs.W, "%v", msg)
		return routing.HRES_RETURN
	}
	var lines = strings.Split(eval.StrVal("text"), "\n")
	if idx >= len(lines) || idx < 0 {
		hs.W.WriteHeader(404)
		var msg = fmt.Sprintf("file(%v,%v) page file not found by index(%v) on extern type(/info/%v), %v page files found",
			rf.Id, rf.Pub, idx, etype, len(lines))
		log.E("%v", msg)
		fmt.Fprintf(hs.W, "%v", msg)
		return routing.HRES_RETURN
	}
	hs.R.URL.Path = t.Pre + strings.Trim(lines[idx], " \t")
	slog("TextSender sending extern file on file(%v,%v) by redirect to %v", rf.Id, rf.Pub, hs.R.URL.Path)
	return t.DefaultSender.DoH(hs, rf, etype, dl, idx)
}
func (t *TextSender) String() string {
	return "TextSender"
}

type JsonSender struct {
	*DefaultSender
}

func NewJsonSender(sender *DefaultSender) *JsonSender {
	return &JsonSender{DefaultSender: sender}
}
func (t *JsonSender) Send(hs *routing.HTTPSession, rf *gfsdb.F, etype string, dl bool, idx int) routing.HResult {
	var eval = rf.Info.MapVal(etype)
	if eval == nil {
		hs.W.WriteHeader(404)
		var msg = fmt.Sprintf("file(%v,%v) extern type(/info/%v) attribute is not exist", rf.Id, rf.Pub, etype)
		log.E("%v", msg)
		fmt.Fprintf(hs.W, "%v", msg)
		return routing.HRES_RETURN
	}
	var files = eval.AryVal("files")
	if idx >= len(files) || idx < 0 {
		hs.W.WriteHeader(404)
		var msg = fmt.Sprintf("file(%v,%v) page file not found by index(%v) on extern type(/info/%v), %v page files found",
			rf.Id, rf.Pub, idx, etype, len(files))
		log.E("%v", msg)
		fmt.Fprintf(hs.W, "%v", msg)
		return routing.HRES_RETURN
	}
	hs.R.URL.Path = t.Pre + strings.Trim(fmt.Sprintf("%v", files[idx]), " \t")
	return t.DefaultSender.DoH(hs, rf, etype, dl, idx)
}
func (t *JsonSender) String() string {
	return "JsonSender"
}

func ParseSenderL(cfg *util.Fcfg, sender_l []string) (map[string]FSedner, error) {
	var ts FSedner
	var ss = map[string]FSedner{}
	for _, sender := range sender_l {
		var sname = cfg.Val2(sender+"/sender", "")
		if len(sname) < 1 {
			return nil, util.Err("the %v/sender is empty", sender)
		}
		var dir = cfg.Val2(sender+"/s_wdir", ".")
		var pref = cfg.Val2(sender+"/s_pref", "")
		var stype_s = strings.Split(cfg.Val2(sender+"/s_type", sender), ",")
		switch sname {
		case "json":
			ts = NewJsonSender(NewDefaultSender2(dir, pref))
		case "text":
			ts = NewTextSender(NewDefaultSender2(dir, pref))
		case "default":
			ts = NewDefaultSender2(dir, pref)
		default:
			return nil, util.Err("not support type(%v) found on %v/s_type", sender)
		}
		for _, st := range stype_s {
			ss[st] = ts
		}
	}
	return ss, nil
}
