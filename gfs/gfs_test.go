package main

import (
	"github.com/Centny/gwf/routing"
	"github.com/Centny/gwf/util"
	"os"
	"testing"
	"time"
)

func init() {
	ef = func(int) {}
}
func TestSrv(t *testing.T) {
	os.RemoveAll("www")
	util.Exec("rm -f *.log")
	routing.Shared.ShowLog = true
	go func() {
		os.Args = []string{"gfs", "-s", "../gfs_s.properties"}
		main()
		panic("done...")
	}()
	time.Sleep(time.Second)
	go func() {
		os.Args = []string{"gfs", "-c", "../gfs_c.properties"}
		main()
		panic("done...")
	}()
	time.Sleep(time.Second)
	//
	os.Args = []string{"gfs", "-u", "http://127.0.0.1:2325", "pkg.sh", "xx", "a=1"}
	main()
	//
	os.Args = []string{"gfs", "-u", "http://127.0.0.1:2325", "gfs.go", "xx"}
	main()
	//
	os.Args = []string{"gfs", "-u", "http://127.0.0.1:23x5", "pkg.sh", "xx", "a=1"}
	main()
	//
	os.Args = []string{"gfs", "-xx", "http://127.0.0.1:23x5", "pkg.sh", "xx", "a=1"}
	main()
	//
	os.Args = []string{"gfs", "-u"}
	main()
	//
	os.Args = []string{"gfs"}
	main()
	//
	os.RemoveAll("www")
	util.Exec("rm -f *.log")
}
