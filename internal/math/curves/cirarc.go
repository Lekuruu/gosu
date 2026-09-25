package curves

import (
	"math"

	"github.com/Lekuruu/gosu/internal/math/math32"
	"github.com/Lekuruu/gosu/internal/math/math87"
	"github.com/Lekuruu/gosu/internal/math/vector"
)

const osuPi float32 = 3.14159274

type CirArc struct {
	pt1, pt2, pt3               vector.Vector2f
	centre                      vector.Vector2f //nolint:misspell
	startAngle, totalAngle, dir float32
	r                           float32

	centreS   vector.Vector2f //nolint:misspell
	tInitialS float64
	tFinalS   float64
	rS        float32

	Unstable bool
}

func NewCirArc(a, b, c vector.Vector2f) *CirArc {
	arc := &CirArc{pt1: a, pt2: b, pt3: c, dir: 1}

	if vector.IsStraightLine32(a, b, c) {
		arc.Unstable = true
	}

	d := 2 * (a.X*(b.Y-c.Y) + b.X*(c.Y-a.Y) + c.X*(a.Y-b.Y))
	aSq := a.LenSq()
	bSq := b.LenSq()
	cSq := c.LenSq()

	arc.centre = vector.NewVec2f(
		aSq*(b.Y-c.Y)+bSq*(c.Y-a.Y)+cSq*(a.Y-b.Y),
		aSq*(c.X-b.X)+bSq*(a.X-c.X)+cSq*(b.X-a.X)).Scl(1 / d) //nolint:misspell

	arc.r = a.Dst(arc.centre)
	arc.startAngle = a.AngleRV(arc.centre)

	endAngle := c.AngleRV(arc.centre)

	for endAngle < arc.startAngle {
		endAngle += 2 * math32.Pi
	}

	arc.totalAngle = endAngle - arc.startAngle

	aToC := c.Sub(a)
	aToC = vector.NewVec2f(aToC.Y, -aToC.X)

	if aToC.Dot(b.Sub(a)) < 0 {
		arc.dir = -arc.dir
		arc.totalAngle = 2*math.Pi - arc.totalAngle
	}

	arc.arcStable(a, b, c)

	return arc
}

func (arc *CirArc) arcStable(a, b, c vector.Vector2f) {
	aX, aY := float64(a.X), float64(a.Y)
	bX, bY := float64(b.X), float64(b.Y)
	cX, cY := float64(c.X), float64(c.Y)

	d := float32(2 * (aX*(bY-cY) + bX*(cY-aY) + cX*(aY-bY)))

	aSq := float64(a.LenSq87())
	bSq := float64(b.LenSq87())
	cSq := float64(c.LenSq87())

	arc.centreS = vector.NewVec2f(
		float32((aSq*(bY-cY)+bSq*(cY-aY)+cSq*(aY-bY))/float64(d)),
		float32((aSq*(cX-bX)+bSq*(aX-cX)+cSq*(bX-aX))/float64(d))) //nolint:misspell

	arc.rS = a.Dst87(arc.centreS)

	arc.tInitialS = ctAt(a, arc.centreS)
	tMid := ctAt(b, arc.centreS)
	arc.tFinalS = ctAt(c, arc.centreS)

	for tMid < arc.tInitialS {
		tMid += 2 * float64(osuPi)
	}

	for arc.tFinalS < arc.tInitialS {
		arc.tFinalS += 2 * float64(osuPi)
	}

	if tMid > arc.tFinalS {
		arc.tFinalS -= 2 * float64(osuPi)
	}
}

func ctAt(pt, centre vector.Vector2f) float64 {
	return math.Atan2(float64(pt.Y)-float64(centre.Y), float64(pt.X)-float64(centre.X))
}

func (arc *CirArc) PointAt(t float32) vector.Vector2f {
	return vector.NewVec2fRad(arc.startAngle+arc.dir*t*arc.totalAngle, arc.r).Add(arc.centre)
}

func (arc *CirArc) PointAtS(t float64) vector.Vector2f {
	theta := arc.tFinalS*t + arc.tInitialS*(1-t)
	return vector.NewVec2f(math87.Add87(float32(math.Cos(theta)*float64(arc.rS)), arc.centreS.X), math87.Add87(float32(math.Sin(theta)*float64(arc.rS)), arc.centreS.Y))
}

func (arc *CirArc) GetLength() float32 {
	return arc.r * arc.totalAngle
}

func (arc *CirArc) GetStartAngle() float32 {
	return arc.pt1.AngleRV(arc.PointAt(1.0 / arc.GetLength()))
}

func (arc *CirArc) GetEndAngle() float32 {
	return arc.pt3.AngleRV(arc.PointAt((arc.GetLength() - 1.0) / arc.GetLength()))
}
