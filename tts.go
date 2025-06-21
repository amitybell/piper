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

type VoiceConfig struct {
	VoiceName string
	ModelCard string
	ModelFn   string
	ConfigFn  string
}

type TTS struct {
	VoiceConfig

	piperExe  string
	piperDir  string
	espeakDir string
}

func (t *TTS) String() string {
	if t == nil {
		return "<TTS>"
	}
	return t.VoiceName
}

func (t *TTS) Synthesize(text string) (wav []byte, err error) {
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

	stdin := strings.NewReader(text)
	stderr := bytes.NewBuffer(nil)
	cmd := exec.Command(t.piperExe,
		"--model", relPath(t.piperDir, t.ModelFn),
		"--config", relPath(t.piperDir, t.ConfigFn),
		"--espeak_data", relPath(t.piperDir, t.espeakDir),
		"--output_file", stdoutFn,
	)
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

func newTTS(dataDir string, configureVoice func(dataDir string) (VoiceConfig, error)) (*TTS, error) {
	if dataDir == "" {
		dir, err := xdg.DataFile("ab-piper")
		if err != nil {
			return nil, fmt.Errorf("piper.Install: cannot create data dir: %w", err)
		}
		dataDir = dir
	}

	vc, err := configureVoice(dataDir)
	if err != nil {
		return nil, fmt.Errorf("piper.Install: cannot install piper voice: %w", err)
	}

	exeFn, err := installPiper(dataDir)
	if err != nil {
		return nil, fmt.Errorf("piper.Install: cannot install piper binary: %w", err)
	}
	piperDir := filepath.Dir(exeFn)

	t := &TTS{
		VoiceConfig: vc,
		piperDir:    piperDir,
		piperExe:    exeFn,
		espeakDir:   filepath.Join(piperDir, "espeak-ng-data"),
	}
	return t, nil
}

func NewEmbedded(piperDir string, va asset.Asset) (*TTS, error) {
	return newTTS(piperDir, func(piperDir string) (VoiceConfig, error) {
		return installEmbeddedVoice(piperDir, va)
	})
}

// DEPRECATED: use NewEmbedded instead
func New(piperDir string, va asset.Asset) (*TTS, error) {
	return NewEmbedded(piperDir, va)
}

func NewExtracted(piperDir string, voiceDir string) (*TTS, error) {
	return newTTS(piperDir, func(piperDir string) (VoiceConfig, error) {
		return configureExtractedVoice(voiceDir)
	})
}

func relPath(rootDir, fn string) string {
	rel, err := filepath.Rel(rootDir, fn)
	if err != nil {
		return fn
	}
	return rel
}
