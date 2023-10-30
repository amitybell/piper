package piper

import (
	"bytes"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/amitybell/piper-asset"
	"os/exec"
	"path/filepath"
	"strings"
)

type TTS struct {
	ModelCard string
	VoiceName string

	onnxFn   string
	jsonFn   string
	piperExe string
	piperDir string
}

func (t *TTS) Synthesize(text string) (wav []byte, err error) {
	stdin := strings.NewReader(text)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := exec.Command(t.piperExe,
		"--model", t.onnxFn,
		"--config", t.jsonFn,
		"--output_file", "-")
	cmd.Dir = t.piperDir
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.SysProcAttr = sysProcAttr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("TextToSpeech: %s: %s: %s", cmd, err, stderr.Bytes())
	}
	return stdout.Bytes(), nil
}

func New(dataDir string, voice asset.Asset) (*TTS, error) {
	if dataDir == "" {
		dir, err := xdg.DataFile("ab-piper")
		if err != nil {
			return nil, fmt.Errorf("piper.Install: cannot create data dir: %w", err)
		}
		dataDir = dir
	}

	desc, onnxFn, jsonFn, err := installVoice(filepath.Join(dataDir, "piper-voice-"+voice.Name), voice.FS)
	if err != nil {
		return nil, fmt.Errorf("piper.Install: cannot install piper voice: %w", err)
	}
	exeFn, err := installPiper(dataDir)
	if err != nil {
		return nil, fmt.Errorf("piper.Install: cannot install piper binary: %w", err)
	}
	t := &TTS{
		ModelCard: desc,
		VoiceName: voice.Name,
		onnxFn:    onnxFn,
		jsonFn:    jsonFn,
		piperDir:  filepath.Dir(exeFn),
		piperExe:  exeFn,
	}
	return t, nil
}
