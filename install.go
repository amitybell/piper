package piper

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	distArcNm = "dist.tar.zst"
)

func extractTar(rootDir, dstNm string, h *tar.Header, r io.Reader) (retErr error) {
	rootDir = filepath.Clean(rootDir)
	dstFn := filepath.Join(rootDir, dstNm)
	rel, err := filepath.Rel(rootDir, dstFn)
	if err != nil {
		return fmt.Errorf("extract: rel(%s,%s): %w", rootDir, dstFn, err)
	}
	if filepath.Join(rootDir, rel) != dstFn {
		return fmt.Errorf("extract: `%s` appears to escaped root `%s`", dstFn, rootDir)
	}

	if h.Typeflag == tar.TypeDir {
		return nil
	}

	if h.Typeflag == tar.TypeSymlink && h.Linkname != "" {
		err := os.Symlink(h.Linkname, dstFn)
		if err != nil {
			return fmt.Errorf("extract: link(%s, %s): `%w`", h.Linkname, dstFn, err)
		}
		return nil
	}

	if h.Typeflag != tar.TypeReg {
		return fmt.Errorf("extract: unsupported file `%s`: type(%d) is not a dir(%d), symlink(%d) or regular file(%d)",
			dstFn, h.Typeflag, tar.TypeDir, tar.TypeSymlink, tar.TypeReg)
	}

	os.MkdirAll(filepath.Dir(dstFn), 0755)
	f, err := os.Create(dstFn)
	if err != nil {
		return fmt.Errorf("extract: create `%s`: %w", dstFn, err)
	}

	_, copyErr := io.Copy(f, r)
	closeErr := f.Close()
	if copyErr != nil {
		return fmt.Errorf("extract: copy `%s`: %w", dstFn, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("extract: close `%s`: %w", dstFn, closeErr)
	}
	return nil
}

func extractTarZst(dstDir string, srcFS fs.FS, srcPth string) error {
	arcRd, err := srcFS.Open(srcPth)
	if err != nil {
		return fmt.Errorf("extractTarZst: open fs `%s`: %w", srcPth, err)
	}
	defer arcRd.Close()

	arc, err := openTarZst(arcRd)
	if err != nil {
		return fmt.Errorf("extractTarZst: open archive `%s`: %w", srcPth, err)
	}
	defer arc.Close()

	for {
		h, err := arc.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return fmt.Errorf("extractTarZst: archive next `%s`: %w", srcPth, err)
		}
		if err := extractTar(dstDir, h.Name, h, arc); err != nil {
			return fmt.Errorf("extractTarZst: archive next `%s`: %w", srcPth, err)
		}
	}
}

func installHash(dstDir string, srcFS fs.FS) (alreadyInstalled bool, err error) {
	hashNm := "hash.txt"
	dstFn := filepath.Join(dstDir, hashNm)
	srcHash, err := fs.ReadFile(srcFS, hashNm)
	if err != nil {
		return false, fmt.Errorf("installHash: read hash file `fs://%s`: %w", hashNm, err)
	}
	dstHash, err := os.ReadFile(dstFn)
	if err == nil && bytes.Equal(dstHash, srcHash) {
		return true, nil
	}
	os.MkdirAll(dstDir, 0755)
	if err := os.WriteFile(dstFn, srcHash, 0644); err != nil {
		return false, fmt.Errorf("installHash: write hash file `%s`: %w", dstFn, err)
	}
	return false, nil
}

func readModelCard(dir string) (string, error) {
	fn := filepath.Join(dir, "MODEL_CARD")
	s, err := os.ReadFile(fn)
	if err != nil {
		return "", fmt.Errorf("readModelCard: `%s`: %w", fn, err)
	}
	return string(s), nil
}

func installVoice(dstDir string, srcFS fs.FS) (desc, onnxFn, jsonFn string, err error) {
	onnxFn = filepath.Join(dstDir, "voice.onnx")
	jsonFn = filepath.Join(dstDir, "voice.json")
	alreadyInstalled, err := installHash(dstDir, srcFS)
	if err != nil {
		return "", "", "", fmt.Errorf("installVoice: %w", err)
	}
	if alreadyInstalled {
		s, err := readModelCard(dstDir)
		if err == nil {
			desc = s
			return desc, onnxFn, jsonFn, nil
		}
	}

	if err := extractTarZst(dstDir, srcFS, distArcNm); err != nil {
		return "", "", "", fmt.Errorf("installVoice: %w", err)
	}

	desc, err = readModelCard(dstDir)
	if err != nil {
		return "", "", "", fmt.Errorf("installVoice: %w", err)
	}

	return desc, onnxFn, jsonFn, nil
}

func installPiper(dataDir string) (exeFn string, err error) {
	pkgDir := filepath.Join(dataDir, "piper-bin-"+piperAsset.Name)
	exeFn = filepath.Join(pkgDir, piperExe)
	defer os.Chmod(exeFn, 0755)
	if ok, err := installHash(pkgDir, piperAsset.FS); err != nil {
		return "", fmt.Errorf("installPiper: %w", err)
	} else if ok {
		return exeFn, nil
	}
	if err := extractTarZst(pkgDir, piperAsset.FS, distArcNm); err != nil {
		return "", fmt.Errorf("installPiper: walk embedded fs: %w", err)
	}
	return exeFn, nil
}
