package vsmod

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/tidwall/jsonc"
)

func ReadZipModInfo(fsys fs.FS, path string) (*Info, error) {
	zipFile, err := fsys.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipFile.Close()

	zipStat, err := zipFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat zip file: %w", err)
	}

	zipFileRA, ok := zipFile.(io.ReaderAt)
	if !ok {
		// TODO: fallback
		return nil, fmt.Errorf("zip file does not support io.ReaderAt")
	}

	zipReader, err := zip.NewReader(zipFileRA, zipStat.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to read zip file: %w", err)
	}

	infoFile, err := zipReader.Open("modinfo.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open modinfo.json in zip: %w", err)
	}
	defer infoFile.Close()

	data, err := io.ReadAll(infoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read modinfo.json: %w", err)
	}

	var info Info
	err = json.Unmarshal(jsonc.ToJSON(data), &info)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modinfo.json: %w", err)
	}

	return &info, nil
}

func ReadModInfos(fsys fs.FS, path string) (map[string]*Info, error) {
	infos := make(map[string]*Info)
	err := fs.WalkDir(fsys, path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(d.Name(), ".zip") {
			info, err := ReadZipModInfo(fsys, path)
			if err != nil {
				return fmt.Errorf("failed to read mod info from %s: %w", d.Name(), err)
			}

			infos[d.Name()] = info
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", path, err)
	}

	return infos, nil
}
