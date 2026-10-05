package gitdata

import "time"

func archiveEntry(name string, ago time.Duration) GitDataEntry {
	return &GitArchiveDesc{
		MetadataPath: name,
		ArchivePath:  name + ".tar",
		Metadata:     &ArchiveMetadata{LastAccessTimestamp: time.Now().Add(-ago).Unix()},
	}
}

func patchEntry(name string, ago time.Duration) GitDataEntry {
	return &GitPatchDesc{
		MetadataPath: name,
		PatchPath:    name + ".patch",
		Metadata:     &PatchMetadata{LastAccessTimestamp: time.Now().Add(-ago).Unix()},
	}
}

func worktreeEntry(name string, ago time.Duration, hasSubmodules bool) GitDataEntry {
	return &GitWorktreeDesc{
		Path:          name,
		LastAccessAt:  time.Now().Add(-ago),
		HasSubmodules: hasSubmodules,
	}
}

func mirrorEntry(name string, ago time.Duration) GitDataEntry {
	return &GitRepoDesc{Path: name, LastAccessAt: time.Now().Add(-ago)}
}

func entryNames(entries []GitDataEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.GetPaths()[0])
	}
	return names
}
