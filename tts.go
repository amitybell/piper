package piper

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	asset "github.com/amitybell/piper-asset"
)

type TTS struct {
	ModelCard string
	VoiceName string

	onnxFn   string
	jsonFn   string
	piperExe string
	piperDir string
}

type VoiceOptions struct {
	// default is: 1.0
	speed float32
	// default is: 0.667
	noise float32
}

func (t *TTS) Synthesize(text string, options *VoiceOptions) (wav []byte, err error) {
	stdoutFn := "-"
	var stdout io.Writer
	if runtime.GOOS != "windows" {
		stdout = bytes.NewBuffer(nil)
	} else {
		tmpDir, err := os.MkdirTemp("", "ab-piper.")
		if err != nil {
			return nil, fmt.Errorf("TTS.Synthesize: Cannot create temp file: %w", err)
		}
		defer os.RemoveAll(tmpDir)
		stdoutFn = filepath.Join(tmpDir, "tts.wav")
	}

	args := []string{
		"--model", t.onnxFn,
		"--config", t.jsonFn,
		"--output_file", stdoutFn,
	}

	if options != nil {
		if options.speed != 0 {
			args = append(args, "--length_scale", fmt.Sprintf("%f", options.speed))
		}
		if options.noise != 0 {
			args = append(args, "--noise_scale", fmt.Sprintf("%f", options.noise))
		}
	}

	stdin := strings.NewReader(text)
	stderr := bytes.NewBuffer(nil)
	cmd := exec.Command(t.piperExe, args...)

	cmd.Dir = t.piperDir
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.SysProcAttr = sysProcAttr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("TTS.Synthesize: %s: %s: %s", cmd, err, stderr.Bytes())
	}

	if stdout != nil {
		return stdout.(*bytes.Buffer).Bytes(), nil
	}

	wav, err = os.ReadFile(stdoutFn)
	if err != nil {
		return nil, fmt.Errorf("TTS.Synthesize: %s", err)
	}
	return wav, nil
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
