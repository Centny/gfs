package gfsapi

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"

	"html/template"

	"os/exec"

	"net/url"

	"github.com/Centny/gfs/gfsdb"
	"github.com/Centny/gwf/log"
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/util"
	"github.com/Centny/iconv/auto"
	"github.com/anacrolix/sync"
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
	log.D("create default sender by dir(%v),pre(%v)", dir, pre)
	return NewDefaultSender(http.FileServer(http.Dir(dir)), pre)
}
func (d *DefaultSender) Send(hs *routing.HTTPSession, rf *gfsdb.F, etype string, dl bool, idx int) routing.HResult {
	hs.R.URL.Path = d.Pre + rf.Path
	return d.DoH(hs, rf, etype, dl, idx)
}
func (d *DefaultSender) DoH(hs *routing.HTTPSession, rf *gfsdb.F, etype string, dl bool, idx int) routing.HResult {
	if dl {
		var filename = hs.CheckValA("filename")
		if len(filename) < 1 {
			filename = rf.Name
		}
		var header = hs.W.Header()
		header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s; filename*=UTF-8''%s", url.QueryEscape(filename), url.QueryEscape(filename)))
	}
	slog("DefaultSender do normal http file server(%v) to %v", d.FH, hs.R.URL.Path)
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
	if rf.Info == nil || len(rf.Info) < 1 {
		hs.W.WriteHeader(404)
		var msg = fmt.Sprintf("file(%v,%v) /info attribute is not exist, the type/index operator is not supported", rf.Id, rf.Pub)
		log.E("%v", msg)
		fmt.Fprintf(hs.W, "%v", msg)
		return routing.HRES_RETURN
	}
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
	if rf.Info == nil || len(rf.Info) < 1 {
		hs.W.WriteHeader(404)
		var msg = fmt.Sprintf("file(%v,%v) /info attribute is not exist, the type/index operator is not supported", rf.Id, rf.Pub)
		log.E("%v", msg)
		fmt.Fprintf(hs.W, "%v", msg)
		return routing.HRES_RETURN
	}
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

type MarkdownSender struct {
	Base        string
	Supported   map[string]int
	MarkdownCmd string
	Errf        *template.Template
	Delay       int64
	Timeout     int64
	Running     bool
	rcmds       map[*exec.Cmd]int64
	rlck        sync.RWMutex
}

func NewMarkdownSender(base, supported, mardkwon string) *MarkdownSender {
	sm := map[string]int{}
	for _, s := range strings.Split(supported, ",") {
		sm[s] = 1
	}
	return &MarkdownSender{
		Base:        base,
		Supported:   sm,
		MarkdownCmd: mardkwon,
		rcmds:       map[*exec.Cmd]int64{},
		Delay:       1000,
		Timeout:     5000,
	}
}

func (m *MarkdownSender) errwrite(hs *routing.HTTPSession, msg interface{}) routing.HResult {
	if m.Errf != nil {
		m.Errf.Execute(hs.W, util.Map{
			"err": msg,
		})
	} else {
		fmt.Fprintf(hs.W, "%v", msg)
	}
	return routing.HRES_RETURN
}

func (m *MarkdownSender) TimeoutLoop() {
	m.Running = true
	for m.Running {
		m.dotimeout()
		time.Sleep(time.Duration(m.Delay) * time.Millisecond)
	}
}

func (m *MarkdownSender) dotimeout() {
	defer func() {
		err := recover()
		if err != nil {
			log.E("MarkdownSender do timeout panic(%v)->\n%v", util.CallStatck())
		}
	}()
	tc := 0
	m.rlck.RLock()
	now := util.Now()
	for cmd, start := range m.rcmds {
		if now-start < m.Timeout {
			continue
		}
		tc++
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
	m.rlck.RUnlock()
	if tc > 0 {
		log.D("MarkdownSender found %v timeout command", tc)
	}
}

func (m *MarkdownSender) Send(hs *routing.HTTPSession, rf *gfsdb.F, etype string, dl bool, idx int) routing.HResult {
	hs.W.Header().Set("Content-Type", "text/html;charset=utf8")
	if m.Supported[rf.EXT] < 1 {
		hs.W.WriteHeader(404)
		var msg = fmt.Sprintf("markdown is not supported by ext(%s) on file(%s)", rf.EXT, rf.Id)
		log.E("%v", msg)
		return m.errwrite(hs, msg)
	}
	tf := fmt.Sprintf("%s/%s", m.Base, rf.Path)
	dataBuf, err := auto.ReadFileAsUtf8(tf)
	if err != nil {
		log.E("MarkdownSender read source file(%s) fail with err(%v)", tf, err)
		return m.errwrite(hs, err.Error())
	}
	var markdown = m.MarkdownCmd
	var errBuf = bytes.NewBuffer(nil)
	var cmd = util.NewCmd(markdown)
	cmd.Stdout = hs.W
	cmd.Stderr = errBuf
	writer, err := cmd.StdinPipe()
	if err != nil {
		log.E("MarkdownSender open command(%v) stdin pipe fail with err(%v)", markdown, err)
		return m.errwrite(hs, err.Error())
	}
	err = cmd.Start()
	if err != nil {
		writer.Close()
		log.E("MarkdownSender start command(%v) fail with err(%v)", markdown, err)
		return m.errwrite(hs, err.Error())
	}
	_, err = fmt.Fprintf(writer, `
%v%v
%v
%v
	`, "```", rf.EXT, string(dataBuf), "```")
	if err != nil {
		writer.Close()
		log.E("MarkdownSender send data to command(%v) fail with err(%v)", markdown, err)
		return m.errwrite(hs, err.Error())
	}
	writer.Close()
	m.rlck.Lock()
	m.rcmds[cmd] = util.Now()
	m.rlck.Unlock()
	err = cmd.Wait()
	m.rlck.Lock()
	delete(m.rcmds, cmd)
	m.rlck.Unlock()
	if err != nil {
		log.E("MarkdownSender wait command fail with err(%v)->\n%v", err, errBuf.String())
		return m.errwrite(hs, err.Error())
	}
	return routing.HRES_RETURN
}

func (m *MarkdownSender) String() string {
	return "MarkdownSender"
}

func (m *MarkdownSender) ParseErrf(errf string) error {
	tmpl, err := template.ParseFiles(errf)
	if err != nil {
		log.E("MarkdownSender parsing errf by path(%v) fail with %v", errf, err)
		return err
	}
	m.Errf = tmpl
	log.I("MarkdownSender parsing errf by path(%v) success", errf)
	return nil
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
		case "markdown":
			mts := NewMarkdownSender(dir,
				cfg.Val2(sender+"/s_supported", ""),
				cfg.Val2(sender+"/s_cmds", "pandoc --highlight-style tango -s"),
			)
			mts.Delay = cfg.Int64ValV(sender+"/s_delay", 1000)
			mts.Timeout = cfg.Int64ValV(sender+"/s_timeout", 5000)
			errf := cfg.Val2(sender+"/s_errf", "")
			if len(errf) > 0 {
				mts.ParseErrf(errf)
			}
			ts = mts
			go mts.TimeoutLoop()
		default:
			return nil, util.Err("not support type(%v) found on %v/s_type", sender)
		}
		for _, st := range stype_s {
			ss[st] = ts
		}
	}
	return ss, nil
}
