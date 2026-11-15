// Package diff provides diff algorithms and diff models.
package diff

type Diff struct {
	Entries []Entry
}

type Entry struct {
	File *File
	Text string
}

type File struct {
	RawHeader []string
	OldPath   string
	NewPath   string
	Binary    bool
	Hunks     []Hunk
}

type Hunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Lines    []Line
}

type Line struct {
	Kind           LineKind
	Content        string
	OldNumber      int
	NewNumber      int
	NoNewlineAtEOF bool
}

type LineKind int

const (
	LineContext LineKind = iota
	LineAdded
	LineDeleted
)
