package objects

import (
	"cmp"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/Lekuruu/gosu/internal/math/curves"
	"github.com/Lekuruu/gosu/internal/math/mutils"
	"github.com/Lekuruu/gosu/internal/math/vector"
	"github.com/Lekuruu/gosu/pkg/beatmaps/audio"
	"github.com/Lekuruu/gosu/pkg/beatmaps/difficulty"
	"github.com/Lekuruu/gosu/pkg/beatmaps/timing"
)

const (
	maxPathLength           = 100_000_000
	maxRepeats              = 10_000
	maxSliderTicksPerRepeat = 32_768
)

type TickPoint struct {
	Time      float64
	IsReverse bool
}

type SliderEdgeSound struct {
	Sound       audio.HitSound
	SampleSet   int
	AdditionSet int
}

type Slider struct {
	*HitObject

	multiCurve *curves.MultiCurve

	Timings *timing.Timings
	TPoint  timing.ControlPoint

	PixelLength float64
	RepeatCount int64

	scorePoints []TickPoint
	edgeSounds  []SliderEdgeSound

	diff *difficulty.Difficulty

	spanDuration float64
}

func NewSlider(data []string) *Slider {
	slider := &Slider{
		HitObject: commonParse(data, 10),
	}
	slider.PositionDelegate = slider.PositionAt

	pixelLength, err := strconv.ParseFloat(data[7], 64)
	if err != nil || pixelLength < 0 || math.IsNaN(pixelLength) || math.IsInf(pixelLength, 0) {
		return nil
	}
	slider.PixelLength = pixelLength

	repeatCount, err := strconv.ParseInt(data[6], 10, 64)
	if err != nil || repeatCount < 1 {
		return nil
	}
	slider.RepeatCount = repeatCount
	if slider.PixelLength*float64(slider.RepeatCount) > maxPathLength*10 {
		return nil
	}
	slider.PixelLength = min(slider.PixelLength, maxPathLength)
	slider.RepeatCount = min(slider.RepeatCount, maxRepeats)

	slider.multiCurve = slider.parseCurve(data[5])
	if slider.multiCurve == nil {
		return nil
	}
	if slider.PixelLength == 0 {
		slider.PixelLength = float64(slider.multiCurve.GetLength())
	}

	slider.EndTime = slider.StartTime
	slider.EndPosRaw = slider.multiCurve.PointAt(1.0)

	baseSample := slider.sounds[0]
	baseHitSample := slider.GetHitSample()

	slider.sounds = make([]audio.HitSound, slider.RepeatCount+1)
	slider.edgeSounds = make([]SliderEdgeSound, slider.RepeatCount+1)

	for i := range slider.sounds {
		slider.sounds[i] = baseSample
		slider.edgeSounds[i] = SliderEdgeSound{
			Sound:       baseSample,
			SampleSet:   baseHitSample.SampleSet,
			AdditionSet: baseHitSample.AdditionSet,
		}
	}

	if len(data) > 8 {
		subData := strings.Split(data[8], "|")
		for i, v := range subData[:min(len(subData), len(slider.sounds))] {
			sound, _ := strconv.Atoi(v)
			slider.sounds[i] = audio.HitSound(sound)
			slider.edgeSounds[i].Sound = audio.HitSound(sound)
		}
	}

	if len(data) > 9 {
		subData := strings.Split(data[9], "|")
		for i, v := range subData[:min(len(subData), len(slider.edgeSounds))] {
			sampleSet, additionSet, _ := strings.Cut(v, ":")
			slider.edgeSounds[i].SampleSet, _ = strconv.Atoi(sampleSet)
			slider.edgeSounds[i].AdditionSet, _ = strconv.Atoi(additionSet)
		}
	}

	return slider
}

func (slider *Slider) parseCurve(curveData string) *curves.MultiCurve {
	curveDef := curves.CurveDef{
		CurveType: -1,
		Points:    []vector.Vector2f{slider.StartPosRaw},
	}
	curveDefs := make([]curves.CurveDef, 0, 1)
	nextType := curves.CType(-1)
	pointCount := 0

	for part := range strings.SplitSeq(curveData, "|") {
		xValue, yValue, isPoint := strings.Cut(part, ":")
		if !isPoint {
			if curveType := parseCurveType(part); curveType >= 0 {
				if curveDef.CurveType < 0 {
					curveDef.CurveType = curveType
				} else {
					nextType = curveType
				}
			}
			continue
		}

		x, _ := strconv.ParseFloat(xValue, 32)
		y, _ := strconv.ParseFloat(yValue, 32)
		point := vector.NewVec2f(float32(x), float32(y))
		if pointCount > 0 || point != slider.StartPosRaw {
			curveDef.Points = append(curveDef.Points, point)
		}
		pointCount++

		if nextType >= 0 {
			curveDefs = append(curveDefs, curveDef)
			curveDef = curves.CurveDef{
				CurveType: nextType,
				Points:    []vector.Vector2f{point},
			}
			nextType = -1
		}
	}

	if len(curveDef.Points) > 1 || len(curveDefs) == 0 {
		if curveDef.CurveType < 0 {
			curveDef.CurveType = curves.CCatmull
		}
		curveDefs = append(curveDefs, curveDef)
	}
	for _, def := range curveDefs {
		if def.CurveType != curves.CBezier {
			continue
		}

		controlDistance := float32(0)
		previous := def.Points[0]
		for _, point := range def.Points[1:] {
			controlDistance += point.Dst(previous)
			previous = point
		}
		if controlDistance >= 2*maxPathLength {
			return nil
		}
	}

	if slider.PixelLength == 0 {
		return curves.NewMultiCurve(curveDefs)
	}
	return curves.NewMultiCurveT(curveDefs, slider.PixelLength)
}

func parseCurveType(value string) curves.CType {
	switch value {
	case "P":
		return curves.CCirArc
	case "L":
		return curves.CLine
	case "B":
		return curves.CBezier
	case "C":
		return curves.CCatmull
	default:
		return -1
	}
}

func (slider *Slider) GetEdgeSounds() []SliderEdgeSound {
	return slices.Clone(slider.edgeSounds)
}

func (slider *Slider) GetScorePoints() []TickPoint {
	return slices.Clone(slider.scorePoints)
}

func (slider *Slider) AppendScorePoints(tick TickPoint) {
	slider.scorePoints = append(slider.scorePoints, tick)
	slices.SortFunc(slider.scorePoints, func(a, b TickPoint) int {
		return cmp.Compare(a.Time, b.Time)
	})
}

func (slider *Slider) PositionAt(time float64) vector.Vector2f {
	if slider.IsRetarded() {
		return slider.StartPosRaw
	}
	t1 := mutils.ClampF64(time, slider.StartTime, slider.EndTime)

	progress := (t1 - slider.StartTime) / slider.spanDuration
	progress = math.Mod(progress, 2)
	if progress >= 1 {
		progress = 2 - progress
	}

	return slider.multiCurve.PointAt(float32(progress))
}

func (slider *Slider) SetTiming(timings *timing.Timings) {
	slider.Timings = timings
	slider.TPoint = timings.GetPointAt(slider.StartTime)

	nanTimingPoint := math.IsNaN(slider.TPoint.GetRawBeatLength())

	velocity := slider.Timings.GetVelocity(slider.TPoint)

	cLength := float64(slider.multiCurve.GetLength())

	slider.spanDuration = cLength * 1000 / velocity

	slider.EndTime = slider.StartTime + cLength*1000*float64(slider.RepeatCount)/velocity

	minDistanceFromEnd := velocity * 0.01
	tickDistance := slider.Timings.GetTickDistance(slider.TPoint)

	if slider.multiCurve.GetLength() > 0 && tickDistance > slider.PixelLength {
		tickDistance = slider.PixelLength
	}
	if cLength/tickDistance > maxSliderTicksPerRepeat {
		tickDistance = cLength / maxSliderTicksPerRepeat
	}

	for span := 0; span < int(slider.RepeatCount); span++ {
		spanStartTime := slider.StartTime + float64(span)*slider.spanDuration
		reversed := span%2 == 1

		// skip ticks if timingPoint has NaN beatLength
		for d := tickDistance; d <= cLength && !nanTimingPoint; d += tickDistance {
			if d >= cLength-minDistanceFromEnd {
				break
			}

			// Always generate ticks from the start of the path rather than the span to ensure
			// that ticks in repeat spans are positioned identically to those in non-repeat spans
			timeProgress := d / cLength
			if reversed {
				timeProgress = 1 - timeProgress
			}

			slider.scorePoints = append(slider.scorePoints, TickPoint{
				Time: spanStartTime + timeProgress*slider.spanDuration,
			})
		}

		if span < int(slider.RepeatCount)-1 {
			slider.scorePoints = append(slider.scorePoints, TickPoint{
				Time:      spanStartTime + slider.spanDuration,
				IsReverse: true,
			})
		} else {
			slider.scorePoints = append(slider.scorePoints, TickPoint{
				Time: max(slider.StartTime+(slider.EndTime-slider.StartTime)/2, slider.EndTime-36),
			})
		}
	}

	slices.SortFunc(slider.scorePoints, func(a, b TickPoint) int {
		return cmp.Compare(a.Time, b.Time)
	})

	slider.EndPosRaw = slider.PositionAt(slider.EndTime)
}

func (slider *Slider) SetDifficulty(diff *difficulty.Difficulty) {
	slider.diff = diff
}

func (slider *Slider) IsRetarded() bool {
	return slider.StartTime == slider.EndTime
}

func (slider *Slider) GetType() Type {
	return SLIDER
}
