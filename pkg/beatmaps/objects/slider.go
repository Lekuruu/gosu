package objects

import (
	"cmp"
	"math"
	"slices"
	"sort"
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
	maxPathLength           = 100_000_000 // Sanity limits, XNOR reaches 10M pixel length so 100M should be enough
	maxRepeats              = 10_000      // Same limit as osu!
	maxSliderTicksPerRepeat = 32_768
)

type TickPoint struct {
	Time      float64
	IsReverse bool
}

type pathLine struct {
	time1 int64
	time2 int64
	line  curves.Linear
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
	scorePath   []pathLine
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
	slider.RepeatCount = min(slider.RepeatCount, maxRepeats) // The same limit as in lazer

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

func (slider *Slider) GetLength() float32 {
	return slider.multiCurve.GetLength()
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

		if pointCount > 0 || point != slider.StartPosRaw { // skip the first point if it's the same as start position.
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
		// Lazer's multi-type slider has 1 point line
		if curveDef.CurveType < 0 {
			// osu! uses catmull if there's no curve type
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
			// Skip sliders which are too computationally expensive
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
	default: // It's a point
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

	index := sort.Search(len(slider.scorePath), func(i int) bool {
		return float64(slider.scorePath[i].time2) >= time
	})

	pLine := slider.scorePath[max(0, min(index, len(slider.scorePath)-1))]

	clamped := mutils.ClampF64(time, float64(pLine.time1), float64(pLine.time2))

	if pLine.time2 == pLine.time1 {
		return pLine.line.Point2
	}

	return pLine.line.PointAt(float32(clamped-float64(pLine.time1)) / float32(pLine.time2-pLine.time1))
}

func (slider *Slider) SetTiming(timings *timing.Timings, beatmapVersion int) {
	slider.Timings = timings
	slider.TPoint = timings.GetPointAt(slider.StartTime)

	nanTimingPoint := math.IsNaN(slider.TPoint.GetRawBeatLength())

	lines := slider.multiCurve.GetLines()

	startTime := slider.StartTime

	velocity := slider.Timings.GetVelocity(slider.TPoint)

	cLength := float64(slider.multiCurve.GetLength())

	minDistanceFromEnd := velocity * 0.01
	tickDistance := slider.Timings.GetTickDistance(slider.TPoint)
	if beatmapVersion < 8 {
		tickDistance = slider.Timings.GetScoringDistance()
	}

	if slider.multiCurve.GetLength() > 0 && tickDistance > slider.PixelLength {
		tickDistance = slider.PixelLength
	}
	// Sanity limit to 32768 ticks per repeat
	if cLength/tickDistance > maxSliderTicksPerRepeat {
		tickDistance = cLength / maxSliderTicksPerRepeat
	}

	scoringLengthTotal := 0.0
	scoringDistance := 0.0

	// Stable-like score point processing, ugly AF.
	for span := range slider.RepeatCount {
		distanceToEnd := float64(slider.multiCurve.GetLength())
		skipTick := nanTimingPoint // NaN SV acts like 1.0x SV, but doesn't spawn slider ticks

		reverse := span%2 == 1

		start := 0
		end := len(lines)
		direction := 1

		if reverse {
			start = len(lines) - 1
			end = -1
			direction = -1
		}

		for j := start; j != end; j += direction {
			line := lines[j]

			p1, p2 := line.Point1, line.Point2

			if reverse {
				p1, p2 = p2, p1
			}

			distance := float32(line.GetCustomLength())

			progress := 1000.0 * float64(distance) / velocity

			slider.scorePath = append(slider.scorePath, pathLine{time1: int64(startTime), time2: int64(startTime + progress), line: curves.NewLinear(p1, p2)})

			startTime += progress
			slider.EndTime = math.Floor(startTime)

			scoringDistance += float64(distance)

			for scoringDistance >= tickDistance && !skipTick && tickDistance != 0 {
				scoringLengthTotal += tickDistance
				scoringDistance -= tickDistance
				distanceToEnd -= tickDistance

				skipTick = distanceToEnd <= minDistanceFromEnd
				if skipTick {
					break
				}

				scoreTime := slider.StartTime + math.Floor(float64(float32(scoringLengthTotal))/velocity*1000)

				slider.scorePoints = append(slider.scorePoints, TickPoint{Time: scoreTime})
			}
		}

		scoringLengthTotal += scoringDistance

		scoreTime := slider.StartTime + math.Floor(float64(float32(scoringLengthTotal))/velocity*1000)

		if span < slider.RepeatCount-1 {
			slider.scorePoints = append(slider.scorePoints, TickPoint{
				Time:      scoreTime,
				IsReverse: true,
			})
		} else {
			slider.scorePoints = append(slider.scorePoints, TickPoint{
				Time: max(slider.StartTime+(slider.EndTime-slider.StartTime)/2, slider.EndTime-36),
			})
		}

		if skipTick {
			scoringDistance = 0
		} else {
			scoringLengthTotal -= tickDistance - scoringDistance
			scoringDistance = tickDistance - scoringDistance
		}
	}

	slider.spanDuration = (slider.EndTime - slider.StartTime) / float64(slider.RepeatCount)

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
