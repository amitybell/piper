package piper

import (
	"github.com/amitybell/piper-asset"
	"github.com/amitybell/piper-voice-alan"
	"github.com/amitybell/piper-voice-jenny"
	"os"
	"testing"
)

func TestPiper(t *testing.T) {
	dataDir, err := os.MkdirTemp("", "ab-piper.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dataDir)

	assets := map[string]asset.Asset{
		"jenny": jenny.Asset,
		"alan":  alan.Asset,
	}

	for name, asset := range assets {
		tts, err := New(dataDir, asset)
		if err != nil {
			t.Fatal(err)
		}
		wav, err := tts.Synthesize("hello world")
		if err != nil {
			t.Fatalf("%s: %s\n", name, err)
		}
		if len(wav) < 44 {
			t.Fatalf("%s: Invalid wav file generated: len(%d)\n", name, len(wav))
		}
	}
}
