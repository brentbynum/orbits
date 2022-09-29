package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	bodies []*Body
}

const (
	screen_width  = 1280
	screen_height = 720
)

var (
	lastUpdate time.Time
)

func (g *Game) GetCollisions(body *Body) []*Body {
	result := make([]*Body, 0)
	for _, b := range g.bodies {
		if body != b {
			d := Distance(body.pos, b.pos)
			if d <= ((body.mass / body.density) + (b.mass/b.density)/3) {
				result = append(result, b)
			}
		}
	}
	return result
}

func (g *Game) Update() error {

	delta := time.Since(lastUpdate)
	lastUpdate = time.Now()
	for _, b := range g.bodies {
		if b.active {
			accel := b.CalcTotalAccelleration(g.bodies)
			b.Update(delta, accel)
			collisions := g.GetCollisions(b)
			b.MergeBodies(collisions)
		}
	}
	filterInactive := make([]*Body, 0)
	for _, b := range g.bodies {
		if b.active {
			filterInactive = append(filterInactive, b)
			if b.pos.x > screen_width {
				b.pos.x = 0
			}
			if b.pos.x < 0 {
				b.pos.x = screen_width
			}
			if b.pos.y > screen_height {
				b.pos.y = 0
			}
			if b.pos.y < 0 {
				b.pos.y = screen_height
			}
		}
	}
	g.bodies = filterInactive
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, b := range g.bodies {
		b.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func NewGame() *Game {
	cx := float64(screen_width) / 2
	cy := float64(screen_height) / 2
	g := &Game{}
	g.bodies = []*Body{
		NewBody("core", cx, cy, 1000, 500, 100), // crank up the density in the core or it gets hard to see
		NewBody("satellite1", cx, cy-150, 7, 6, 1),
		NewBody("satellite2", cx+200, cy, 11, 5, 1),
		NewBody("satellite3", cx-100, cy+150, 13, 9, 1),
		NewBody("satellite4", cx+150, cy-150, 17, 4, 1),
		NewBody("satellite5", cx, cy-75, 7, 5, 3),
		NewBody("satellite6", cx+125, cy+100, 11, 8, 1),
		NewBody("satellite7", cx-125, cy+50, 13, 7, 1),
		NewBody("satellite8", cx+150, cy+150, 17, 12, 1),
	}
	for _, b := range g.bodies {
		b.velocity.x = rand.Float64() * 6
		b.velocity.y = rand.Float64() * 6
	}
	return g
}

func main() {
	lastUpdate = time.Now()
	ebiten.SetWindowSize(screen_width, screen_height)
	ebiten.SetWindowTitle("Orbits!")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
