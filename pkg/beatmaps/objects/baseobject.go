package objects

import (
	"strconv"
	"strings"

	"github.com/Lekuruu/gosu/internal/math/vector"
	"github.com/Lekuruu/gosu/pkg/beatmaps/audio"
)

func commonParse(data []string, hitSampleIndex int) *HitObject {
	x, _ := strconv.ParseFloat(data[0], 32)
	y, _ := strconv.ParseFloat(data[1], 32)
	time, _ := strconv.ParseFloat(data[2], 64)
	objType, _ := strconv.Atoi(data[3])

	startPos := vector.NewVec2f(float32(x), float32(y))

	sound, _ := strconv.Atoi(data[4])

	hitObject := &HitObject{
		StartPosRaw: startPos,
		EndPosRaw:   startPos,
		StartTime:   time,
		EndTime:     time,
		HitObjectID: -1,
		NewCombo:    (Type(objType) & NEWCOMBO) == NEWCOMBO,
		ColorOffset: (objType >> 4) & 7,
		sounds:      []audio.HitSound{audio.HitSound(sound)},
	}
	hitObject.BasicHitSound = parseHitSample(data, hitSampleIndex)

	return hitObject
}

func parseHitSample(data []string, index int) (info audio.HitSoundInfo) {
	if index >= len(data) || data[index] == "" {
		return info
	}

	fields := strings.Split(data[index], ":")
	if len(fields) > 0 {
		info.SampleSet, _ = strconv.Atoi(fields[0])
	}
	if len(fields) > 1 {
		info.AdditionSet, _ = strconv.Atoi(fields[1])
	}
	if len(fields) > 2 {
		info.CustomIndex, _ = strconv.Atoi(fields[2])
	}
	if len(fields) > 3 {
		volume, _ := strconv.Atoi(fields[3])
		info.CustomVolume = float64(volume) / 100
	}

	return info
}
