The File Privew Usage
===
### For Public URL
for preview public url is adding query path after the public url.the full privew url is liked `http://<public url>/<privew command>/<page number>`

example:

* preview `http://fs.dev.gdy.io/OhJoA1==.docx` second page is `http://fs.dev.gdy.io/OhJoA1==.docx/D_docx/1.jpg`

* preview `http://fs.dev.gdy.io/SIQqMx==.go` source is `http://fs.dev.gdy.io/SIQqMx==.go/mdview.html`

note: all supported command is in `Supported` section


### For Private URL
for preview private url is adding query argument on the private url.the full privew url is liked `http://~/usr/api/dload?fid=xxx&type=<privew command>&idx=<page number>`


example:

* preview `http://fs.dev.gdy.io/usr/api/dload?fid=587d9317d624d31884c97621` second page is `http://fs.dev.gdy.io/usr/api/dload?fid=587d9317d624d31884c97621&type=D_docx&idx=1`


* preview `http://fs.dev.gdy.io/usr/api/dload?fid=587d9698d624d31884c97627` second page is `http://fs.dev.gdy.io/usr/api/dload?fid=587d9698d624d31884c97627&type=mdview`


note: all supported command is in `Supported` section


### Suppored
all supporeted command is mapping by file extendsion

* `.doc,.docx,.doc,.docx,.xps,.rtf` -> `D_docx`   --Convert Needed
* `.ppt,.pptx` -> `D_pptx`  --Convert Needed
* `.pdf` -> `D_pdfx` --Convert Needed
* `.go,.h,.hpp,.c,.cpp,.java,.js,.cs,.m,.sh,.swift,.xml,.properties,.ini,.html,.css,.json,.sql,.txt` -> `mdview`
* `.wmv,.rm,.rmvb,.mpg,.mpeg,.mpe,.3gp,.mov,.mp4,.m4v,.avi,.mkv,.flv,.vob` ->`V_pc`(PC) ->`V_phone`(phone) --Convert Needed
* `.jpg,.jpeg,.png,.bmp` ->`small`
* `.amr` -> `smp3`
* `.flac,.wav` -> `mp3` --Convert Needed