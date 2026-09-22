package beatmaps_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Lekuruu/gosu/pkg/beatmaps"
)

func TestParseFromReaderParsesUTF8Beatmap(t *testing.T) {
	file := openFixture(t, "lost-umbrella.osu")

	beatMap, err := beatmaps.ParseFromReader(file)
	if err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}

	if beatMap.Artist != "inabakumori" {
		t.Errorf("Artist = %q, want %q", beatMap.Artist, "inabakumori")
	}
	if beatMap.ArtistUnicode != "稲葉曇" {
		t.Errorf("ArtistUnicode = %q, want %q", beatMap.ArtistUnicode, "稲葉曇")
	}
	if beatMap.Title != "Lost Umbrella" {
		t.Errorf("Title = %q, want %q", beatMap.Title, "Lost Umbrella")
	}
	if beatMap.TitleUnicode != "ロストアンブレラ" {
		t.Errorf("TitleUnicode = %q, want %q", beatMap.TitleUnicode, "ロストアンブレラ")
	}
	if len(beatMap.HitObjects) != 887 {
		t.Errorf("len(HitObjects) = %d, want 887", len(beatMap.HitObjects))
	}
	if !beatMap.Timings.HasPoints() {
		t.Error("Timings.HasPoints() = false, want true")
	}
}

func TestParseFromReaderParsesUTF16Beatmap(t *testing.T) {
	file := openFixture(t, "lost-umbrella-utf16.osu")

	beatMap, err := beatmaps.ParseFromReader(file)
	if err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}

	if beatMap.ArtistUnicode != "稲葉曇" {
		t.Errorf("ArtistUnicode = %q, want %q", beatMap.ArtistUnicode, "稲葉曇")
	}
	if beatMap.TitleUnicode != "ロストアンブレラ" {
		t.Errorf("TitleUnicode = %q, want %q", beatMap.TitleUnicode, "ロストアンブレラ")
	}
}

func TestParseFromByteParsesLargeBeatmap(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "save-me.osu"))
	if err != nil {
		t.Fatal(err)
	}

	beatMap, err := beatmaps.ParseFromByte(data)
	if err != nil {
		t.Fatalf("ParseFromByte() error = %v", err)
	}

	if beatMap.Artist != "Avenged Sevenfold" {
		t.Errorf("Artist = %q, want %q", beatMap.Artist, "Avenged Sevenfold")
	}
	if len(beatMap.HitObjects) != 3201 {
		t.Errorf("len(HitObjects) = %d, want 3201", len(beatMap.HitObjects))
	}
	if beatMap.MapID != 1256809 {
		t.Errorf("MapID = %d, want 1256809", beatMap.MapID)
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
