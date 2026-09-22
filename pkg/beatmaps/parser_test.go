package beatmaps_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lekuruu/gosu/pkg/beatmaps"
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
}

func TestParseFromByteParsesLargeBeatmap(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "save-me.osu"))
	if err != nil {
		t.Fatal(err)
	}

	beatmap, err := beatmaps.ParseFromByte(data)
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
