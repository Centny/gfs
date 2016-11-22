package gfsapi

import (
	"github.com/Centny/gwf/routing"
)

func TestHtml(hs *routing.HTTPSession) routing.HResult {
	hs.SendT(`
<html>
<body>
<form action="/usr/api/uload?pub=1&recorded=1" method="POST" enctype="multipart/form-data">
<input type="file" name="file">
<input type="submit" value="submit">
<input type="reset" value="reset">
</form>
</body>
</html>
		`, "text/html;charset=utf8")
	return routing.HRES_RETURN
}
