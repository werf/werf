package status

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/types"
	"github.com/werf/werf/v3/pkg/true_git"
)

// Status returns Result with path lists of untracked files and modified files for index and worktree.
// The function counts each file status as Modified if it is not Unmodified or Untracked ([ADU] == M).
// The function does not work with ignored, renamed and copied files.
func Status(ctx context.Context, workTreeDir string) (r Result, err error) {
	logboek.Context(ctx).Debug().
		LogProcess("Status %s", workTreeDir).
		Options(func(options types.LogProcessOptionsInterface) {
			if !debug() {
				options.Mute()
			}
		}).
		Do(func() {
			r, err = status(ctx, workTreeDir)

			if debug() {
				logboek.Context(ctx).Debug().LogF("result: %+v\n", r)
				logboek.Context(ctx).Debug().LogLn("err:", err)
			}
		})

	return
}

func status(ctx context.Context, workTreeDir string) (Result, error) {
	result := Result{}

	statusCmd := true_git.NewGitCmd(ctx, &true_git.GitCmdOptions{RepoDir: workTreeDir}, "status", "--porcelain=v2", "-z", "--untracked-files=all", "--no-renames")
	if err := statusCmd.Run(ctx); err != nil {
		return result, fmt.Errorf("git status command failed: %w", err)
	}

	detailedErrFunc := func(err error) error {
		return fmt.Errorf("%w\n\ncommand: %q\noutput:\n%s", err, statusCmd, strings.ReplaceAll(statusCmd.OutErrBuf.String(), "\x00", "\n"))
	}

	// With -z entries are NUL-terminated and paths are printed verbatim, without C-style quoting.
	for _, entryLine := range strings.SplitAfter(statusCmd.OutBuf.String(), "\x00") {
		entryLine = strings.TrimSuffix(entryLine, "\x00")
		if len(entryLine) == 0 {
			continue
		}

		formatTypeCode := entryLine[0]
		switch formatTypeCode {
		case '1':
			if err := parseOrdinaryEntry(&result, entryLine); err != nil {
				return result, detailedErrFunc(err)
			}
		case 'u':
			if err := parseUnmergedEntry(&result, entryLine); err != nil {
				return result, detailedErrFunc(err)
			}
		case '?':
			if err := parseUntrackedEntry(&result, entryLine); err != nil {
				return result, detailedErrFunc(err)
			}
		case '2', '!':
			panic(detailedErrFunc(fmt.Errorf("unexpected git status line format: %q", entryLine)))
		default:
			return result, detailedErrFunc(fmt.Errorf("invalid git status line format: %q", entryLine))
		}
	}

	return result, nil
}

type ordinaryEntry struct {
	xy         string
	sub        string
	mh, mi, mw string
	hH, hI     string
	path       string

	raw string
}

// 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
func parseOrdinaryEntry(r *Result, entryLine string) error {
	fields := strings.SplitN(entryLine, " ", 9)
	entry := ordinaryEntry{
		xy:   fields[1],
		sub:  fields[2],
		mh:   fields[3],
		mi:   fields[4],
		mw:   fields[5],
		hH:   fields[6],
		hI:   fields[7],
		path: fields[8],
		raw:  entryLine,
	}

	switch {
	case entry.sub == "N...":
		return parseOrdinaryRegularFileEntry(r, entry)
	case entry.sub[0] == 'S':
		return parseOrdinarySubmoduleEntry(r, entry)
	default:
		return fmt.Errorf("invalid git status ordinary <sub> field: %q (%q)", entry.raw, entry.sub)
	}
}

// 1 <XY> N... <mH> <mI> <mW> <hH> <hI> <path>
func parseOrdinaryRegularFileEntry(result *Result, entry ordinaryEntry) error {
	if len(entry.xy) != 2 {
		return fmt.Errorf("invalid git status ordinary <XY> field: %q (%q)", entry.raw, entry.xy)
	}

	stageCode := entry.xy[0]
	worktreeCode := entry.xy[1]

	if stageCode != '.' {
		result.Index.pathList = append(result.Index.pathList, entry.path)
		result.Index.addToChecksum(entry.path, entry.mi, entry.hI)
	}

	if worktreeCode != '.' {
		result.Worktree.pathList = append(result.Worktree.pathList, entry.path)
	}

	return nil
}

// 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
// 1 <XY> S<c><m><u> <mH> <mI> <mW> <hH> <hI> <path>
func parseOrdinarySubmoduleEntry(result *Result, entry ordinaryEntry) error {
	if len(entry.sub) != 4 {
		return fmt.Errorf("invalid git status ordinary <sub> field: %q (%q)", entry.raw, entry.sub)
	}

	stageCode := entry.xy[0]
	worktreeCode := entry.xy[1]

	commitChangedCode := entry.sub[1]
	trackedChangesCode := entry.sub[2]
	untrackedChangesCode := entry.sub[3]
	newSubmoduleFunc := func(scopeCode uint8) submodule {
		return submodule{
			Path:                entry.path,
			IsAdded:             scopeCode == 'A',
			IsDeleted:           scopeCode == 'D',
			IsModified:          scopeCode == 'M',
			IsCommitChanged:     commitChangedCode != '.',
			HasTrackedChanges:   trackedChangesCode != '.',
			HasUntrackedChanges: untrackedChangesCode != '.',
		}
	}

	if stageCode != '.' {
		result.Index.submodules = append(result.Index.submodules, newSubmoduleFunc(stageCode))
		result.Index.addToChecksum(entry.path, entry.mi, entry.hI)
	}

	if worktreeCode != '.' {
		result.Worktree.submodules = append(result.Worktree.submodules, newSubmoduleFunc(worktreeCode))
	}

	return nil
}

type unmergedEntry struct {
	xy             string
	sub            string
	m1, m2, m3, mW string
	h1, h2, h3     string
	path           string

	raw string
}

// u <xy> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
func parseUnmergedEntry(result *Result, entryLine string) error {
	fields := strings.SplitN(entryLine, " ", 11)
	entry := unmergedEntry{
		xy:   fields[1],
		sub:  fields[2],
		m1:   fields[3],
		m2:   fields[4],
		m3:   fields[5],
		mW:   fields[6],
		h1:   fields[7],
		h2:   fields[8],
		h3:   fields[9],
		path: fields[10],
		raw:  entryLine,
	}

	if len(entry.xy) != 2 {
		return fmt.Errorf("invalid git status ordinary <XY> field: %q (%q)", entry.raw, entry.xy)
	}

	stageCode := entry.xy[0]
	worktreeCode := entry.xy[1]

	if stageCode != '.' {
		result.Index.pathList = append(result.Index.pathList, entry.path)
		result.Index.addToChecksum(entry.path)
	}

	if worktreeCode != '.' {
		result.Worktree.pathList = append(result.Worktree.pathList, entry.path)
	}

	return nil
}

type untrackedEntry struct {
	path string

	raw string
}

// ? <path>
func parseUntrackedEntry(result *Result, entryLine string) error {
	entry := untrackedEntry{
		path: strings.TrimPrefix(entryLine, "? "),
		raw:  entryLine,
	}

	result.UntrackedPathList = append(result.UntrackedPathList, entry.path)

	return nil
}

func debug() bool {
	return os.Getenv("WERF_DEBUG_GIT_STATUS") == "1"
}
