package gfsapi

import "github.com/Centny/gwf/routing"

func (f *FSH) Hand(pre string, mux *routing.SessionMux) {
	mux.HFunc("^"+pre+"/pub/api/info(\\?.*)?", f.Info)
	mux.HFunc("^"+pre+"/usr/api/redoTask(\\?.*)?", RedoTask)
	mux.HFunc("^"+pre+"/pub/api/listInfo(\\?.*)?", f.ListInfo)
	mux.HFunc("^"+pre+"/usr/api/uload(\\?.*)?", f.Up)
	mux.HFunc("^"+pre+"/usr/api/dload(\\?.*)?", f.Down)
	mux.HFunc("^"+pre+"/usr/api/listFile(\\?.*)?", ListFile)
	mux.HFunc("^"+pre+"/usr/api/updateFile(\\?.*)?$", UpdateFile)
	mux.HFunc("^"+pre+"/usr/api/updateFileParent(\\?.*)?$", UpdateFileParent)
	mux.HFunc("^"+pre+"/usr/api/removeFile(\\?.*)?", RemoveFile)
	mux.HFunc("^"+pre+"/usr/api/addFolder(\\?.*)?", AddFolder)
	mux.HFunc("^"+pre+"/usr/test.html(\\?.*)?", TestHtml)
	mux.HFunc("^"+pre+"/.*$", f.Pub)
}

func AdmHand(pre string, mux *routing.SessionMux) {
	mux.HFunc("^"+pre+"/adm/verify(\\?.*)?", AdmVerify)
}
