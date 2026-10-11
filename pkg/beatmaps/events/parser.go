package events

import (
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type commandFrame struct {
	indent  int
	command *Command
}

// Parser parses raw lines from an [Events] section.
type Parser struct {
	events        Collection
	currentObject *StoryboardObject
	commandStack  []commandFrame
}

func (parser *Parser) Events() Collection {
	return parser.events
}

func (parser *Parser) ParseLine(line string) error {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "//") {
		return nil
	}

	commandLine := strings.TrimLeft(line, " \t")
	commandName, _, _ := strings.Cut(commandLine, ",")

	if commandType, ok := parseCommandType(strings.TrimSpace(commandName)); ok {
		if parser.currentObject == nil {
			return fmt.Errorf("storyboard command %q has no preceding sprite or animation", commandType)
		}
		return parser.parseCommand(line, commandLine, commandType)
	}

	parser.currentObject = nil
	parser.commandStack = nil

	fields, err := parseFields(trimmed)
	if err != nil {
		return fmt.Errorf("parse event fields: %w", err)
	}

	event, err := parseEvent(fields, line)
	if err != nil {
		return err
	}
	parser.events = append(parser.events, event)

	switch event := event.(type) {
	case *Sprite:
		parser.currentObject = &event.StoryboardObject
	case *Animation:
		parser.currentObject = &event.StoryboardObject
	}

	return nil
}

func (parser *Parser) parseCommand(line, commandLine string, commandType CommandType) error {
	fields, err := parseFields(commandLine)
	if err != nil {
		return fmt.Errorf("parse storyboard command fields: %w", err)
	}

	command := &Command{
		Type:       commandType,
		Parameters: append([]string(nil), fields[1:]...),
	}
	indent := len(line) - len(commandLine)

	for len(parser.commandStack) > 0 && indent <= parser.commandStack[len(parser.commandStack)-1].indent {
		parser.commandStack = parser.commandStack[:len(parser.commandStack)-1]
	}

	if len(parser.commandStack) == 0 {
		parser.currentObject.Commands = append(parser.currentObject.Commands, command)
	} else {
		parent := parser.commandStack[len(parser.commandStack)-1].command
		parent.Children = append(parent.Children, command)
	}

	if commandType.IsBlock() {
		parser.commandStack = append(parser.commandStack, commandFrame{
			indent:  indent,
			command: command,
		})
	}

	return nil
}

func parseEvent(fields []string, raw string) (Event, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("event is empty")
	}

	eventType, known := parseEventType(fields[0])
	if !known {
		return &RawEvent{Raw: raw}, nil
	}

	switch eventType {
	case TypeBackground:
		return parseBackground(fields)
	case TypeVideo:
		return parseVideo(fields)
	case TypeBreak:
		return parseBreak(fields)
	case TypeColourTransformation:
		return parseColourTransformation(fields)
	case TypeSprite:
		return parseSprite(fields)
	case TypeSample:
		return parseSample(fields)
	case TypeAnimation:
		return parseAnimation(fields)
	default:
		return &RawEvent{Raw: raw}, nil
	}
}

func parseBackground(fields []string) (Event, error) {
	if err := requireFieldCount("background", fields, 3, 5); err != nil {
		return nil, err
	}

	startTime, err := parseInt64(fields[1], "background start time")
	if err != nil {
		return nil, err
	}
	x, y, err := parseOptionalPosition(fields[3:])
	if err != nil {
		return nil, err
	}

	return &Background{
		StartTime: startTime,
		Filename:  fields[2],
		X:         x,
		Y:         y,
	}, nil
}

func parseVideo(fields []string) (Event, error) {
	if err := requireFieldCount("video", fields, 3, 5); err != nil {
		return nil, err
	}

	startTime, err := parseInt64(fields[1], "video start time")
	if err != nil {
		return nil, err
	}
	x, y, err := parseOptionalPosition(fields[3:])
	if err != nil {
		return nil, err
	}

	return &Video{
		StartTime: startTime,
		Filename:  fields[2],
		X:         x,
		Y:         y,
	}, nil
}

func parseBreak(fields []string) (Event, error) {
	if err := requireFieldCount("break", fields, 3, 3); err != nil {
		return nil, err
	}

	startTime, err := parseInt64(fields[1], "break start time")
	if err != nil {
		return nil, err
	}
	endTime, err := parseInt64(fields[2], "break end time")
	if err != nil {
		return nil, err
	}

	return &Break{StartTime: startTime, EndTime: endTime}, nil
}

func parseColourTransformation(fields []string) (Event, error) {
	if err := requireFieldCount("colour transformation", fields, 5, 5); err != nil {
		return nil, err
	}

	startTime, err := parseInt64(fields[1], "colour transformation start time")
	if err != nil {
		return nil, err
	}
	red, err := parseInt(fields[2], "red channel")
	if err != nil {
		return nil, err
	}
	green, err := parseInt(fields[3], "green channel")
	if err != nil {
		return nil, err
	}
	blue, err := parseInt(fields[4], "blue channel")
	if err != nil {
		return nil, err
	}

	return &ColourTransformation{
		StartTime: startTime,
		Red:       red,
		Green:     green,
		Blue:      blue,
	}, nil
}

func parseSprite(fields []string) (Event, error) {
	if err := requireFieldCount("sprite", fields, 6, 6); err != nil {
		return nil, err
	}

	object, err := parseStoryboardObject(fields[1:6])
	if err != nil {
		return nil, err
	}

	return &Sprite{StoryboardObject: object}, nil
}

func parseAnimation(fields []string) (Event, error) {
	if err := requireFieldCount("animation", fields, 8, 9); err != nil {
		return nil, err
	}

	object, err := parseStoryboardObject(fields[1:6])
	if err != nil {
		return nil, err
	}
	frameCount, err := parseInt(fields[6], "animation frame count")
	if err != nil {
		return nil, err
	}
	frameDelay, err := parseInt64(fields[7], "animation frame delay")
	if err != nil {
		return nil, err
	}
	loop := AnimationLoopForever
	if len(fields) == 9 {
		loop = AnimationLoop(fields[8])
	}

	return &Animation{
		StoryboardObject: object,
		FrameCount:       frameCount,
		FrameDelay:       frameDelay,
		Loop:             loop,
	}, nil
}

func parseSample(fields []string) (Event, error) {
	if err := requireFieldCount("sample", fields, 4, 5); err != nil {
		return nil, err
	}

	startTime, err := parseInt64(fields[1], "sample start time")
	if err != nil {
		return nil, err
	}
	layer, err := parseLayer(fields[2])
	if err != nil {
		return nil, err
	}
	volume := 100
	if len(fields) == 5 {
		volume, err = parseInt(fields[4], "sample volume")
		if err != nil {
			return nil, err
		}
	}

	return &Sample{
		StartTime: startTime,
		Layer:     layer,
		Filename:  fields[3],
		Volume:    volume,
	}, nil
}

func parseStoryboardObject(fields []string) (StoryboardObject, error) {
	layer, err := parseLayer(fields[0])
	if err != nil {
		return StoryboardObject{}, err
	}
	x, err := parseFloat(fields[3], "storyboard x position")
	if err != nil {
		return StoryboardObject{}, err
	}
	y, err := parseFloat(fields[4], "storyboard y position")
	if err != nil {
		return StoryboardObject{}, err
	}

	return StoryboardObject{
		Layer:    layer,
		Origin:   Origin(fields[1]),
		Filename: fields[2],
		X:        x,
		Y:        y,
	}, nil
}

func parseOptionalPosition(fields []string) (float64, float64, error) {
	var x, y float64
	var err error

	if len(fields) > 0 {
		x, err = parseFloat(fields[0], "x offset")
		if err != nil {
			return 0, 0, err
		}
	}
	if len(fields) > 1 {
		y, err = parseFloat(fields[1], "y offset")
		if err != nil {
			return 0, 0, err
		}
	}

	return x, y, nil
}

func parseEventType(value string) (Type, bool) {
	if numeric, err := strconv.Atoi(value); err == nil {
		if numeric >= int(TypeBackground) && numeric <= int(TypeAnimation) {
			return Type(numeric), true
		}

		return TypeUnknown, false
	}

	switch value {
	case "Background":
		return TypeBackground, true
	case "Video":
		return TypeVideo, true
	case "Break":
		return TypeBreak, true
	case "Colour", "Color", "ColourTransformation", "ColorTransformation":
		return TypeColourTransformation, true
	case "Sprite":
		return TypeSprite, true
	case "Sample":
		return TypeSample, true
	case "Animation":
		return TypeAnimation, true
	default:
		return TypeUnknown, false
	}
}

func parseLayer(value string) (Layer, error) {
	if numeric, err := strconv.Atoi(value); err == nil {
		if numeric >= int(LayerBackground) && numeric <= int(LayerOverlay) {
			return Layer(numeric), nil
		}

		return 0, fmt.Errorf("storyboard layer is invalid: %q", value)
	}

	switch value {
	case "Background":
		return LayerBackground, nil
	case "Fail", "Failing":
		return LayerFail, nil
	case "Pass", "Passing":
		return LayerPass, nil
	case "Foreground":
		return LayerForeground, nil
	case "Overlay":
		return LayerOverlay, nil
	default:
		return 0, fmt.Errorf("storyboard layer is invalid: %q", value)
	}
}

func parseFields(line string) ([]string, error) {
	reader := csv.NewReader(strings.NewReader(line))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	fields, err := reader.Read()
	if err != nil {
		return nil, err
	}
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}

	return fields, nil
}

func requireFieldCount(name string, fields []string, minimum, maximum int) error {
	if len(fields) < minimum || len(fields) > maximum {
		if minimum == maximum {
			return fmt.Errorf("%s event has %d fields; expected %d", name, len(fields), minimum)
		}
		return fmt.Errorf("%s event has %d fields; expected %d to %d", name, len(fields), minimum, maximum)
	}
	return nil
}

func parseInt(value, name string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %q", name, value)
	}
	return parsed, nil
}

func parseInt64(value, name string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %q", name, value)
	}
	return parsed, nil
}

func parseFloat(value, name string) (float64, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %q", name, value)
	}
	if math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, fmt.Errorf("%s is invalid: %q", name, value)
	}
	return parsed, nil
}
