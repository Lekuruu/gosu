package beatmaps

import (
	"bytes"
	"cmp"
	"errors"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/Lekuruu/gosu/internal/files"
	"github.com/Lekuruu/gosu/internal/math/mutils"
	"github.com/Lekuruu/gosu/pkg/beatmaps/objects"
)

const bufferSize = 10 * 1024 * 1024

func parseGeneral(line []string, beatmap *Beatmap) bool {
	switch line[0] {
	case "Mode":
		beatmap.Mode, _ = strconv.Atoi(line[1])
	case "StackLeniency":
		beatmap.StackLeniency, _ = strconv.ParseFloat(line[1], 64)
		if math.IsNaN(beatmap.StackLeniency) {
			beatmap.StackLeniency = 0.0
		}
	case "AudioFilename":
		beatmap.Audio += line[1]
	case "PreviewTime":
		beatmap.PreviewTime, _ = strconv.ParseInt(line[1], 10, 64)
		//case "SampleSet":
		//	switch line[1] {
		//	case "Normal", "All":
		//		beatmap.Timings.BaseSet = 1
		//	case "Soft", "None":
		//		beatmap.Timings.BaseSet = 2
		//	case "Drum":
		//		beatmap.Timings.BaseSet = 3
		//	}
		//	beatmap.Timings.LastSet = beatmap.Timings.BaseSet
	}

	return false
}

func parseMetadata(line []string, beatmap *Beatmap) {
	switch line[0] {
	case "Title":
		beatmap.Title = line[1]
	case "TitleUnicode":
		beatmap.TitleUnicode = line[1]
	case "Artist":
		beatmap.Artist = line[1]
	case "ArtistUnicode":
		beatmap.ArtistUnicode = line[1]
	case "Creator":
		beatmap.Creator = line[1]
	case "FileVersion":
		beatmap.Version = line[1]
	case "Source":
		beatmap.Source = line[1]
	case "Tags":
		beatmap.Tags = line[1]
	case "BeatmapID":
		beatmap.MapID, _ = strconv.ParseInt(line[1], 10, 64)
	case "BeatmapSetID":
		beatmap.SetID, _ = strconv.ParseInt(line[1], 10, 64)
	}
}

func parseDifficulty(line []string, beatmap *Beatmap) {
	switch line[0] {
	case "SliderMultiplier":
		beatmap.SliderMultiplier, _ = strconv.ParseFloat(line[1], 64)
		beatmap.Timings.SliderMult = beatmap.SliderMultiplier
	case "ApproachRate":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatmap.Difficulty.SetAR(mutils.ClampF64(parsed, 0, 10))
		beatmap.arSpecified = true
	case "CircleSize":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatmap.Difficulty.SetCS(mutils.ClampF64(parsed, 0, 10))
	case "SliderTickRate":
		beatmap.Timings.TickRate, _ = strconv.ParseFloat(line[1], 64)
	case "HPDrainRate":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatmap.Difficulty.SetHP(mutils.ClampF64(parsed, 0, 10))
	case "OverallDifficulty":
		parsed, _ := strconv.ParseFloat(line[1], 64)
		beatmap.Difficulty.SetOD(mutils.ClampF64(parsed, 0, 10))

		if !beatmap.arSpecified {
			beatmap.Difficulty.SetAR(beatmap.Difficulty.GetOD())
		}
	}
}

func parseEvents(line []string, beatmap *Beatmap) {
	switch line[0] {
	case "Background", "0":
		beatmap.Bg = strings.Replace(line[2], "\"", "", -1)
	case "Break", "2":
		beatmap.Pauses = append(beatmap.Pauses, NewPause(line))
	}
}

func parseHitObjects(line []string, beatmap *Beatmap) {
	obj := objects.CreateObject(line)

	if obj != nil {
		beatmap.HitObjects = append(beatmap.HitObjects, obj)
	}
}

func tokenize(line, delimiter string) []string {
	return tokenizeN(line, delimiter, -1)
}

func tokenizeN(line, delimiter string, n int) []string {
	if strings.HasPrefix(line, "//") || !strings.Contains(line, delimiter) {
		return nil
	}

	divided := strings.SplitN(line, delimiter, n)

	for i, a := range divided {
		divided[i] = strings.TrimSpace(a)
	}

	return divided
}

func getSection(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "[") {
		return strings.TrimRight(strings.TrimLeft(line, "["), "]")
	}

	return ""
}

func ParseFromByte(data []byte) (*Beatmap, error) {
	return ParseFromReader(bytes.NewReader(data))
}

func ParseFromReader(reader io.Reader) (*Beatmap, error) {
	beatmap := NewBeatmap()

	scanner := files.NewScannerBuf(reader, bufferSize)

	var currentSection string

	counter := 0

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "osu file format v") {
			trim := strings.TrimPrefix(line, "osu file format v")
			beatmap.FileVersion, _ = strconv.Atoi(trim)
		}

		section := getSection(line)
		if section != "" {
			currentSection = section
			continue
		}

		switch currentSection {
		case "General":
			if arr := tokenizeN(line, ":", 2); len(arr) > 1 {
				parseGeneral(arr, beatmap)
			}
		case "Metadata":
			if arr := tokenizeN(line, ":", 2); len(arr) > 1 {
				parseMetadata(arr, beatmap)
			}
		case "Difficulty":
			if arr := tokenizeN(line, ":", 2); len(arr) > 1 {
				parseDifficulty(arr, beatmap)
			}
		case "Events":
			if arr := tokenize(line, ","); len(arr) > 1 {
				parseEvents(arr, beatmap)
			}
		case "TimingPoints":
			if arr := tokenize(line, ","); len(arr) > 1 {
				beatmap.ParsePoint(line)
				counter++
			}
		case "HitObjects":
			if arr := tokenize(line, ","); arr != nil {
				var time string

				objTypeI, _ := strconv.Atoi(arr[3])
				objType := objects.Type(objTypeI)
				if (objType & objects.CIRCLE) > 0 {
					beatmap.Circles++
					time = arr[2]
				} else if (objType & objects.SPINNER) > 0 {
					beatmap.Spinners++
					time = arr[5]
				} else if (objType & objects.SLIDER) > 0 {
					beatmap.Sliders++
					time = arr[2]
				} else if (objType & objects.LONGNOTE) > 0 {
					beatmap.Sliders++
					time = strings.Split(arr[5], ":")[0]
				}
				timeI, _ := strconv.Atoi(time)

				beatmap.Length = max(beatmap.Length, timeI)

				parseHitObjects(arr, beatmap)
			}
		}
	}

	beatmap.FinalizePoints()

	if beatmap.Title+beatmap.Artist+beatmap.Creator == "" || counter == 0 {
		return nil, errors.New("corrupted file")
	}

	slices.SortStableFunc(beatmap.HitObjects, func(a, b objects.IHitObject) int {
		return cmp.Compare(a.GetStartTime(), b.GetStartTime())
	})

	num := 0
	comboNumber := 1
	comboSet := 0
	comboSetHax := 0
	forceNewCombo := false

	for _, iO := range beatmap.HitObjects {
		if iO.GetType() == objects.SPINNER {
			forceNewCombo = true
		} else if iO.IsNewCombo() || forceNewCombo {
			iO.SetNewCombo(true)
			comboNumber = 1
			comboSet++
			comboSetHax += int(iO.GetColorOffset()) + 1

			forceNewCombo = false
		}

		iO.SetID(num)
		iO.SetComboNumber(comboNumber)
		iO.SetComboSet(comboSet)
		iO.SetComboSetHax(comboSetHax)

		comboNumber++
		num++
	}

	for _, obj := range beatmap.HitObjects {
		obj.SetTiming(beatmap.Timings)
	}

	calculateStackLeniency(beatmap)

	return beatmap, nil
}
