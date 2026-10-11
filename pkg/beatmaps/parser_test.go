package beatmaps_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lekuruu/gosu/pkg/beatmaps"
	"github.com/Lekuruu/gosu/pkg/beatmaps/events"
)

func TestParseFromReaderParsesUTF8Beatmap(t *testing.T) {
	file := openFixture(t, "lost-umbrella.osu")

	beatmap, err := beatmaps.ParseFromReader(file)
	if err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}

	if beatmap.Artist != "inabakumori" {
		t.Errorf("Artist = %q, want %q", beatmap.Artist, "inabakumori")
	}
	if beatmap.ArtistUnicode != "稲葉曇" {
		t.Errorf("ArtistUnicode = %q, want %q", beatmap.ArtistUnicode, "稲葉曇")
	}
	if beatmap.Title != "Lost Umbrella" {
		t.Errorf("Title = %q, want %q", beatmap.Title, "Lost Umbrella")
	}
	if beatmap.TitleUnicode != "ロストアンブレラ" {
		t.Errorf("TitleUnicode = %q, want %q", beatmap.TitleUnicode, "ロストアンブレラ")
	}
	if len(beatmap.HitObjects) != 887 {
		t.Errorf("len(HitObjects) = %d, want 887", len(beatmap.HitObjects))
	}
	if !beatmap.Timings.HasPoints() {
		t.Error("Timings.HasPoints() = false, want true")
	}
	if len(beatmap.Events.Backgrounds()) != 1 {
		t.Errorf("len(Events.Backgrounds()) = %d, want 1", len(beatmap.Events.Backgrounds()))
	}
	if len(beatmap.Events.Breaks()) != 4 {
		t.Errorf("len(Events.Breaks()) = %d, want 4", len(beatmap.Events.Breaks()))
	}
}

func TestParseFromReaderParsesUTF16Beatmap(t *testing.T) {
	file := openFixture(t, "lost-umbrella-utf16.osu")

	beatmap, err := beatmaps.ParseFromReader(file)
	if err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}

	if beatmap.ArtistUnicode != "稲葉曇" {
		t.Errorf("ArtistUnicode = %q, want %q", beatmap.ArtistUnicode, "稲葉曇")
	}
	if beatmap.TitleUnicode != "ロストアンブレラ" {
		t.Errorf("TitleUnicode = %q, want %q", beatmap.TitleUnicode, "ロストアンブレラ")
	}
	if len(beatmap.Events.Backgrounds()) != 1 {
		t.Errorf("len(Events.Backgrounds()) = %d, want 1", len(beatmap.Events.Backgrounds()))
	}
}

func TestParseFromByteParsesLargeBeatmap(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "save-me.osu"))
	if err != nil {
		t.Fatal(err)
	}

	beatmap, err := beatmaps.ParseFromBytes(data)
	if err != nil {
		t.Fatalf("ParseFromByte() error = %v", err)
	}

	if beatmap.Artist != "Avenged Sevenfold" {
		t.Errorf("Artist = %q, want %q", beatmap.Artist, "Avenged Sevenfold")
	}
	if len(beatmap.HitObjects) != 3201 {
		t.Errorf("len(HitObjects) = %d, want 3201", len(beatmap.HitObjects))
	}
	if beatmap.MapID != 1256809 {
		t.Errorf("MapID = %d, want 1256809", beatmap.MapID)
	}
}

func TestParseFromReaderParsesOsuTutorial(t *testing.T) {
	file := openFixture(t, "osu-tutorial.osu")

	beatmap, err := beatmaps.ParseFromReader(file)
	if err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}
	if len(beatmap.Events) != 97 {
		t.Fatalf("len(Events) = %d, want 97", len(beatmap.Events))
	}

	breaks := beatmap.Events.Breaks()
	if len(breaks) != 2 {
		t.Fatalf("len(Events.Breaks()) = %d, want 2", len(breaks))
	}
	if breaks[0].StartTime != 34862 || breaks[0].EndTime != 82246 {
		t.Errorf("first break = (%d, %d), want (34862, 82246)", breaks[0].StartTime, breaks[0].EndTime)
	}

	sprites := beatmap.Events.Filter[*events.Sprite]()
	if len(sprites) != 57 {
		t.Fatalf("len(sprites) = %d, want 57", len(sprites))
	}

	firstSprite := sprites[0]
	if firstSprite.Filename != "bg.jpg" || firstSprite.X != 320 || firstSprite.Y != 264 {
		t.Errorf("first sprite = %#v, want bg.jpg at (320, 264)", firstSprite.StoryboardObject)
	}
	if len(firstSprite.Commands) != 3 {
		t.Fatalf("len(first sprite commands) = %d, want 3", len(firstSprite.Commands))
	}

	animations := beatmap.Events.Filter[*events.Animation]()
	if len(animations) != 2 {
		t.Fatalf("len(animations) = %d, want 2", len(animations))
	}

	samples := beatmap.Events.Filter[*events.Sample]()
	if len(samples) != 35 {
		t.Fatalf("len(samples) = %d, want 35", len(samples))
	}

	transformations := beatmap.Events.Filter[*events.ColourTransformation]()
	if len(transformations) != 1 {
		t.Fatalf("len(colour transformations) = %d, want 1", len(transformations))
	}

	if raw := beatmap.Events.Filter[*events.RawEvent](); len(raw) != 0 {
		t.Errorf("len(raw events) = %d, want 0", len(raw))
	}
}

func openFixture(t *testing.T, name string) *os.File {
	t.Helper()

	file, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close fixture: %v", err)
		}
	})

	return file
}
