// App shows one box that logs Element.Touches and its clicks (#1884 check).
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

var phases = []string{"began", "moved", "stationary", "ended", "canceled"}

func main() {
	clicks := 0
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{Title: "touch check", Width: 800, Height: 600, Content: ui.View(func(c *ui.Context) {
			box := ui.Box(c).Fill().Background(ui.RGB(230, 240, 255))
			down := 0
			for _, t := range box.Touches() {
				if t.Phase != ui.TouchStationary {
					fmt.Printf("%s TOUCH id=%d %s x=%.0f y=%.0f start=%.0f,%.0f p=%.2f\n", time.Now().Format("15:04:05.000"), t.ID, phases[t.Phase], t.X, t.Y, t.StartX, t.StartY, t.Pressure)
				}
				if t.Phase < ui.TouchEnded {
					down++
				}
			}
			if box.Clicked() {
				clicks++
				fmt.Printf("%s CLICK %d\n", time.Now().Format("15:04:05.000"), clicks)
			}
			ui.Text(c, fmt.Sprintf("fingers down: %d, clicks: %d", down, clicks))
		})})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
