package curves

import (
	"math"

	"github.com/Lekuruu/gosu/internal/math/vector"
)

func ApproximateCircularArc(pt1, pt2, pt3 vector.Vector2f, detail float32) []Linear {
	arc := NewCirArc(pt1, pt2, pt3)

	if arc.Unstable {
		return []Linear{NewLinear(pt1, pt2), NewLinear(pt2, pt3)}
	}

	segments := int(math.Abs((arc.tFinalS-arc.tInitialS)*float64(arc.rS)) * float64(detail))
	lines := make([]Linear, segments)

	if segments == 0 {
		return lines
	}

	previous := pt1
	for i := 1; i < segments; i++ {
		point := arc.PointAtS(float64(i) / float64(segments))
		lines[i-1] = NewLinear(previous, point)
		previous = point
	}
	lines[segments-1] = NewLinear(previous, pt3)

	return lines
}

func ApproximateCatmullRom(points []vector.Vector2f, detail int) []Linear {
	catmull := NewCatmull(points)
	lines := make([]Linear, detail)

	for i := 0; i < detail; i++ {
		lines[i] = NewLinear(catmull.PointAt(float32(i)/float32(detail)), catmull.PointAt(float32(i+1)/float32(detail)))
	}

	return lines
}

func ApproximateBezier(points []vector.Vector2f) []Linear {
	extracted := NewBezierApproximator(points).CreateBezier()
	lines := make([]Linear, len(extracted)-1)

	for i := 0; i < len(lines); i++ {
		lines[i] = NewLinear(extracted[i], extracted[i+1])
	}

	return lines
}
