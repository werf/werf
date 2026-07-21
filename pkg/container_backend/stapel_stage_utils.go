package container_backend

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func makeScript(commands []string, verbose bool) []byte {
	var scriptCommands []string
	for _, c := range commands {
		if verbose {
			scriptCommands = append(scriptCommands, fmt.Sprintf(`printf "$ %%s\n" %q`, c))
		}
		scriptCommands = append(scriptCommands, c)
	}

	if verbose {
		return []byte(fmt.Sprintf(`#!/bin/sh

set -e

if [ "x$_IS_REEXEC" = "x" ]; then
	if type bash >/dev/null 2>&1 ; then
		echo "# Using bash to execute commands"
		echo
		export _IS_REEXEC="1"
		exec bash $0
	else
		echo "# Using /bin/sh to execute commands"
		echo
	fi
fi

%s
`, strings.Join(scriptCommands, "\n")))
	} else {
		return []byte(fmt.Sprintf(`#!/bin/sh

set -e

if [ "x$_IS_REEXEC" = "x" ]; then
	if type bash >/dev/null 2>&1 ; then
		export _IS_REEXEC="1"
		exec bash $0
	fi
fi

%s
`, strings.Join(scriptCommands, "\n")))
	}
}

func parseVolume(volume string) (string, string, string, error) {
	volumeParts := strings.Split(volume, ":")

	switch len(volumeParts) {
	case 2:
		return volumeParts[0], volumeParts[1], "", nil
	case 3:
		return volumeParts[0], volumeParts[1], volumeParts[2], nil
	default:
		return "", "", "", fmt.Errorf("expected SOURCE:DESTINATION[:OPTIONS] format")
	}
}

// lchownIfSet applies ownership to a path when uid or gid is explicitly requested.
// Tar archives produced by git don't include an entry for the root destination
// directory itself (e.g. /srv/app), only for its contents. Because os.MkdirAll
// creates that directory as root:root, we must chown it separately — otherwise
// git.add with owner/group applies ownership to files but not to the destination.
func lchownIfSet(path string, uid, gid *uint32) error {
	if uid == nil && gid == nil {
		return nil
	}

	numUID, numGID := -1, -1
	if uid != nil {
		numUID = int(*uid)
	}
	if gid != nil {
		numGID = int(*gid)
	}

	if err := os.Lchown(path, numUID, numGID); err != nil {
		return fmt.Errorf("chown %q: %w", path, err)
	}

	return nil
}

func extractTarWithChown(tarFileReader io.Reader, dstDir string, uid, gid *uint32) error {
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return fmt.Errorf("create dir %q: %w", dstDir, err)
	}

	if err := lchownIfSet(dstDir, uid, gid); err != nil {
		return err
	}

	tarReader := tar.NewReader(tarFileReader)
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		entryPath := filepath.Join(dstDir, hdr.Name)
		fi := hdr.FileInfo()

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(entryPath, fi.Mode()); err != nil {
				return fmt.Errorf("create dir %q: %w", entryPath, err)
			}
		case tar.TypeBlock, tar.TypeChar, tar.TypeReg, tar.TypeFifo:
			if err := os.MkdirAll(filepath.Dir(entryPath), os.ModePerm); err != nil {
				return fmt.Errorf("create dir %q: %w", filepath.Dir(entryPath), err)
			}
			f, err := os.OpenFile(entryPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
			if err != nil {
				return fmt.Errorf("create file %q: %w", entryPath, err)
			}
			if _, err := io.Copy(f, tarReader); err != nil {
				f.Close()
				return fmt.Errorf("write file %q: %w", entryPath, err)
			}
			f.Close()
		case tar.TypeLink:
			if err := os.MkdirAll(filepath.Dir(entryPath), os.ModePerm); err != nil {
				return fmt.Errorf("create dir %q: %w", filepath.Dir(entryPath), err)
			}
			if err := os.Link(filepath.Join(dstDir, hdr.Linkname), entryPath); err != nil {
				return fmt.Errorf("create hard link %q: %w", entryPath, err)
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(entryPath), os.ModePerm); err != nil {
				return fmt.Errorf("create dir %q: %w", filepath.Dir(entryPath), err)
			}
			if err := os.Symlink(hdr.Linkname, entryPath); err != nil {
				return fmt.Errorf("create symlink %q: %w", entryPath, err)
			}
		default:
			return fmt.Errorf("tar entry %q has unexpected type %d", hdr.Name, hdr.Typeflag)
		}

		if err := lchownIfSet(entryPath, uid, gid); err != nil {
			return err
		}
	}

	return nil
}
