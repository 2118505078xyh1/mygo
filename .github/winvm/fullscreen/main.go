// Fullscreen checks that a page's own full screen takes its window full
// screen (#1872). The harness clicks the page and writes commands for the
// app to $FSCMD; everything is logged to stdout.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/egoist/mygo"
)

type Log struct{}

func (Log) Say(s string) { fmt.Println(ts(), "page:", s) }

func ts() string { return time.Now().Format("15:04:05.000") }

const page = `<!doctype html><html><head><title>fs check</title>
<style>html,body{margin:0;height:100%}button{width:100%;height:100%;font-size:40px;background:#3a7;color:#fff;border:0}</style></head>
<body><button id=b>full screen</button><script>
b.onclick = () => document.documentElement.requestFullscreen().then(() => mygo.call("Log.Say", "request ok"), (e) => mygo.call("Log.Say", "request err " + e));
document.addEventListener("fullscreenchange", () => mygo.call("Log.Say", "fullscreenElement=" + (document.fullscreenElement ? document.fullscreenElement.tagName : "none")));
</script></body></html>`

func main() {
	mygo.Bind(Log{})
	cmd := os.Getenv("FSCMD")
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{Title: "fs check", Width: 640, Height: 480})
		report := func(ev string) {
			wd, ht := w.Size()
			fmt.Println(ts(), ev, "full=", w.IsFullScreen(), "size=", wd, ht)
		}
		w.OnEnterFullScreen(func() { report("OnEnterFullScreen") })
		w.OnLeaveFullScreen(func() { report("OnLeaveFullScreen") })
		w.Page().LoadHTML(page, "")
		go func() {
			for range time.Tick(200 * time.Millisecond) {
				b, err := os.ReadFile(cmd)
				if err != nil || len(b) == 0 {
					continue
				}
				os.Remove(cmd)
				c := strings.TrimSpace(string(b))
				mygo.RunOnMain(func() {
					switch c {
					case "app-full":
						w.SetFullScreen(true)
					case "app-leave":
						w.SetFullScreen(false)
					}
					report("cmd " + c)
				})
			}
		}()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
