package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	G            float64 = 0.125
	max_mass     float64 = 100.0
	max_density  float64 = 10.0
	max_velocity float64 = 10.0
	max_energy   float64 = 100
)

var (
	bodyImage  *ebiten.Image
	emptyImage = ebiten.NewImage(3, 3)

	// emptySubImage is an internal sub image of emptyImage.
	// Use emptySubImage at DrawTriangles instead of emptyImage in order to avoid bleeding edges.
	emptySubImage = emptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func init() {
	emptyImage.Fill(color.White)
	bodyImage = ebiten.NewImage(32, 32)
	var arcPath vector.Path
	arcPath.Arc(16, 16, 16, 0, 2*math.Pi, vector.Clockwise)
	op := &ebiten.DrawTrianglesOptions{
		FillRule: ebiten.EvenOdd,
	}
	vs, is := arcPath.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
	}
	bodyImage.DrawTriangles(vs, is, emptySubImage, op)
}

type Vec struct {
	x float64
	y float64
}

func NewVec(x, y float64) *Vec {
	return &Vec{
		x: x,
		y: y,
	}
}

func (v *Vec) GetLength() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y)
}

func (v *Vec) Normalize() {
	l := v.GetLength()
	v.x /= l
	v.y /= l
}
func (v *Vec) Scale(factor float64) {
	v.x *= factor
	v.y *= factor
}
func Distance(v1 *Vec, v2 *Vec) float64 {
	d2 := DistanceSquared(v1, v2)
	if d2 > 0 {
		return math.Sqrt(d2)
	}
	return 0
}

func DistanceSquared(v1 *Vec, v2 *Vec) float64 {
	dx := v2.x - v1.x
	dy := v2.y - v1.y
	return dx*dx + dy*dy
}

func Diff(v1 *Vec, v2 *Vec) *Vec {
	v := NewVec(v2.x-v1.x, v2.y-v1.y)
	return v
}

func SumForces(forces []*Force) *Vec {
	x := 0.0
	y := 0.0
	for _, force := range forces {
		x += force.vector.x * force.strength
		y += force.vector.y * force.strength
	}
	return NewVec(x, y)
}

type Force struct {
	targetBody *Body
	strength   float64
	vector     *Vec
}

type Body struct {
	name             string
	pos              *Vec
	velocity         *Vec
	mass             float64
	density          float64
	energy           float64
	active           bool
	spotForces       []*Force
	maxForceStrength float64
}

func NewBody(name string, x, y, mass, density, energy float64) *Body {
	return &Body{
		name:       name,
		pos:        NewVec(x, y),
		velocity:   NewVec(0, 0),
		mass:       mass,
		density:    density,
		energy:     energy,
		active:     true,
		spotForces: make([]*Force, 0),
	}
}

func (b *Body) MergeBodies(bodies []*Body) {
	for _, body := range bodies {
		if body.active {

			fmt.Println("Merging ", body.name, " + ", b.name)
			b.name = fmt.Sprint(body.name, ", ", b.name)
			b.density = (b.density*b.mass + body.density*body.mass) / (b.mass + body.mass)
			b.velocity.x = (b.velocity.x*b.mass + body.velocity.x*body.mass) / (b.mass + body.mass)
			b.velocity.y = (b.velocity.y*b.mass + body.velocity.y*body.mass) / (b.mass + body.mass)
			b.pos = body.pos
			b.mass += body.mass
			b.energy += body.energy
			if b.active {
				body.active = false
			}
		}
	}
}

func (b *Body) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	sizeFactor := (b.mass / b.density)
	op.GeoM.Scale(sizeFactor, sizeFactor)
	op.GeoM.Translate(b.pos.x, b.pos.y)
	op.GeoM.Translate(0, 0)

	op.ColorM.Scale(b.energy/max_energy, b.velocity.GetLength()/max_velocity, b.velocity.GetLength()/max_velocity, b.density/max_density)
	screen.DrawImage(bodyImage, op)

	ebitenutil.DebugPrintAt(screen, b.name, int(b.pos.x), int(b.pos.y))

	// if b.maxForceStrength > 0 {
	// 	for _, force := range b.spotForces {
	// 		x1 := b.pos.x
	// 		y1 := b.pos.y
	// 		x2 := x1 + force.vector.x
	// 		y2 := y1 + force.vector.y
	// 		alpha := math.Max((force.strength/b.maxForceStrength)*255, 128)
	// 		ebitenutil.DrawLine(screen, x1, y1, x2, y2, color.RGBA{255, 0, 0, uint8(alpha)})
	// 		ebitenutil.DrawLine(screen, x1, y1, force.targetBody.pos.x, force.targetBody.pos.y, color.RGBA{0, 255, 0, 64})
	// 	}
	// }

	// s := fmt.Sprint("forces: ", b.spotForces)
	// ebitenutil.DebugPrintAt(screen, s, int(b.pos.x), int(b.pos.y)+35)
	// s := fmt.Sprintf("Velocity: %5.2f", b.velocity.GetLength())
	// ebitenutil.DebugPrintAt(screen, s, int(b.pos.x), int(b.pos.y)+40)

	// s = fmt.Sprintf("Mass: %5.3f", b.mass)
	// ebitenutil.DebugPrintAt(screen, s, int(b.pos.x), int(b.pos.y)+60)
	// s = fmt.Sprintf("Density: %5.3f", b.density)
	// ebitenutil.DebugPrintAt(screen, s, int(b.pos.x), int(b.pos.y)+80)
}

func (b *Body) Update(delta time.Duration, accel *Vec) {
	b.velocity.x += accel.x * delta.Seconds()
	b.velocity.y += accel.y * delta.Seconds()
	b.pos.x += b.velocity.x * delta.Seconds()
	b.pos.y += b.velocity.y * delta.Seconds()
}
