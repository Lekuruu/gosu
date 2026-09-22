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

	slider.PixelLength, _ = strconv.ParseFloat(data[7], 64)
	slider.RepeatCount, _ = strconv.ParseInt(data[6], 10, 64)

	list := strings.Split(data[5], "|")
	points := []vector.Vector2f{slider.StartPosRaw}

	for i := 1; i < len(list); i++ {
		list2 := strings.Split(list[i], ":")
		x, _ := strconv.ParseFloat(list2[0], 32)
		y, _ := strconv.ParseFloat(list2[1], 32)
		points = append(points, vector.NewVec2f(float32(x), float32(y)))
	}

	slider.multiCurve = curves.NewMultiCurveT(list[0], points, slider.PixelLength)

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
