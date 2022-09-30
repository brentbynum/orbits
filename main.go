package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	bodies []*Body
}

const (
	screen_width  = 2560
	screen_height = 1280
)

var (
	lastUpdate time.Time
)

func (g *Game) GetCollisions(body *Body) []*Body {
	result := make([]*Body, 0)
	for _, b := range g.bodies {
		if body != b {
			d := Distance(body.pos, b.pos)
			if d < ((body.mass/body.density)+(b.mass/b.density))/2 {
				result = append(result, b)
			}
		}
	}
	return result
}

func (g *Game) CalcTotalAccelleration(b *Body) *Vec {
	forces := make([]*Force, 0)
	b.maxForceStrength = 0.0
	for _, body := range g.bodies {
		if b != body {
			d2 := DistanceSquared(b.pos, body.pos)
			f := G * ((b.mass * body.mass) / d2)
			str := (f * body.mass) / b.mass
			dir := Diff(b.pos, body.pos)
			dir.Normalize()
			dir.Scale(20)
			forces = append(forces, &Force{
				strength:   str,
				vector:     dir,
				targetBody: body,
			})
			if str > b.maxForceStrength {
				b.maxForceStrength = str
			}
		}
	}
	b.spotForces = forces
	return SumForces(forces)
}

func (g *Game) ProcessBody(delta time.Duration, b *Body) {
	accel := g.CalcTotalAccelleration(b)
	b.Update(delta, accel)
	collisions := g.GetCollisions(b)
	b.MergeBodies(collisions)
}

func (g *Game) Update() error {

	filterInactive := make([]*Body, 0)
	for _, b := range g.bodies {
		if b.active {
			filterInactive = append(filterInactive, b)
		}
	}
	g.bodies = filterInactive
	delta := time.Since(lastUpdate)
	lastUpdate = time.Now()
	for _, b := range g.bodies {
		if b.active {
			g.ProcessBody(delta, b)
		}
	}

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

	}
	for i := 1; i < 20; i++ {
		x := cx + (rand.Float64() * 600) - 300
		y := cy + (rand.Float64() * 400) - 200
		m := (rand.Float64() * 30)
		d := m - (rand.Float64() * m)
		g.bodies = append(g.bodies, NewBody(fmt.Sprint("satellite ", i), x, y, m, d, 1))
	}
	for _, b := range g.bodies {
		if b.name != "core" {
			b.velocity.x = (rand.Float64() * 16) - 8.0
			b.velocity.y = (rand.Float64() * 16) - 8.0
		}
	}
	return g
}

func main() {
	lastUpdate = time.Now()
	ebiten.SetWindowSize(screen_width, screen_height)
	ebiten.SetWindowTitle("Orbits!")
	rand.Seed(time.Now().UTC().UnixNano())
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
