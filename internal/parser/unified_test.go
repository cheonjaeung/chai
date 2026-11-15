package parser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cheonjaeung/chai/internal/diff"
)

func TestParseUnifiedModify(t *testing.T) {
	text := `diff --git a/a.txt b/a.txt
index a1b2c3d..1a2b3c4 123456
--- a/a.txt
+++ b/a.txt
@@ -1,5 +1,6 @@
 line1
-line2
+line2-changed
 line3
 line4
 line5
+line6
`
	expect := diff.Diff{Entries: []diff.Entry{
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/a.txt b/a.txt",
				"index a1b2c3d..1a2b3c4 123456",
				"--- a/a.txt",
				"+++ b/a.txt",
			},
			OldPath: "a.txt",
			NewPath: "a.txt",
			Hunks: []diff.Hunk{{
				OldStart: 1,
				OldLines: 5,
				NewStart: 1,
				NewLines: 6,
				Lines: []diff.Line{
					contextLine("line1", 1, 1),
					deletedLine("line2", 2),
					addedLine("line2-changed", 2),
					contextLine("line3", 3, 3),
					contextLine("line4", 4, 4),
					contextLine("line5", 5, 5),
					addedLine("line6", 6),
				},
			}},
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedAdd(t *testing.T) {
	text := `diff --git a/newfile.txt b/newfile.txt
new file mode 123456
index 0000000..a1b2c3d
--- /dev/null
+++ b/newfile.txt
@@ -0,0 +1,2 @@
+brand new file
+content here
`
	expect := diff.Diff{Entries: []diff.Entry{
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/newfile.txt b/newfile.txt",
				"new file mode 123456",
				"index 0000000..a1b2c3d",
				"--- /dev/null",
				"+++ b/newfile.txt",
			},
			OldPath: "/dev/null",
			NewPath: "newfile.txt",
			Hunks: []diff.Hunk{{
				OldStart: 0,
				OldLines: 0,
				NewStart: 1,
				NewLines: 2,
				Lines: []diff.Line{
					addedLine("brand new file", 1),
					addedLine("content here", 2),
				},
			}},
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedDelete(t *testing.T) {
	text := `diff --git a/a.txt b/a.txt
deleted file mode 123456
index a1b2c3d..0000000
--- a/a.txt
+++ /dev/null
@@ -1,5 +0,0 @@
-line1
-line2
-line3
-line4
-line5
`
	expect := diff.Diff{Entries: []diff.Entry{
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/a.txt b/a.txt",
				"deleted file mode 123456",
				"index a1b2c3d..0000000",
				"--- a/a.txt",
				"+++ /dev/null",
			},
			OldPath: "a.txt",
			NewPath: "/dev/null",
			Hunks: []diff.Hunk{{
				OldStart: 1,
				OldLines: 5,
				NewStart: 0,
				NewLines: 0,
				Lines: []diff.Line{
					deletedLine("line1", 1),
					deletedLine("line2", 2),
					deletedLine("line3", 3),
					deletedLine("line4", 4),
					deletedLine("line5", 5),
				},
			}},
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedBinaryAdd(t *testing.T) {
	text := `diff --git a/bin.dat b/bin.dat
new file mode 123456
index 0000000..a1b2c3d
Binary files /dev/null and b/bin.dat differ
`
	expect := diff.Diff{Entries: []diff.Entry{
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/bin.dat b/bin.dat",
				"new file mode 123456",
				"index 0000000..a1b2c3d",
				"Binary files /dev/null and b/bin.dat differ",
			},
			OldPath: "bin.dat",
			NewPath: "bin.dat",
			Binary:  true,
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedBinaryModify(t *testing.T) {
	text := `diff --git a/bin.dat b/bin.dat
index a1b2c3d..1a2b3c4 123456
Binary files a/bin.dat and b/bin.dat differ
`
	expect := diff.Diff{Entries: []diff.Entry{
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/bin.dat b/bin.dat",
				"index a1b2c3d..1a2b3c4 123456",
				"Binary files a/bin.dat and b/bin.dat differ",
			},
			OldPath: "bin.dat",
			NewPath: "bin.dat",
			Binary:  true,
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedModeModify(t *testing.T) {
	text := `diff --git a/a.txt b/a.txt
old mode 123456
new mode 654321
`
	expect := diff.Diff{Entries: []diff.Entry{
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/a.txt b/a.txt",
				"old mode 123456",
				"new mode 654321",
			},
			OldPath: "a.txt",
			NewPath: "a.txt",
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedRenamePure(t *testing.T) {
	text := `diff --git a/a.txt b/renamed.txt
similarity index 100%
rename from a.txt
rename to renamed.txt
`
	expect := diff.Diff{Entries: []diff.Entry{
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/a.txt b/renamed.txt",
				"similarity index 100%",
				"rename from a.txt",
				"rename to renamed.txt",
			},
			OldPath: "a.txt",
			NewPath: "renamed.txt",
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedRenameWithEdit(t *testing.T) {
	text := `diff --git a/a.txt b/renamed.txt
similarity index 63%
rename from a.txt
rename to renamed.txt
index a1b2c3d..1a2b3c4 123456
--- a/a.txt
+++ b/renamed.txt
@@ -1,5 +1,5 @@
 line1
-line2
+line2-changed
 line3
 line4
 line5
`
	expect := diff.Diff{Entries: []diff.Entry{
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/a.txt b/renamed.txt",
				"similarity index 63%",
				"rename from a.txt",
				"rename to renamed.txt",
				"index a1b2c3d..1a2b3c4 123456",
				"--- a/a.txt",
				"+++ b/renamed.txt",
			},
			OldPath: "a.txt",
			NewPath: "renamed.txt",
			Hunks: []diff.Hunk{{
				OldStart: 1,
				OldLines: 5,
				NewStart: 1,
				NewLines: 5,
				Lines: []diff.Line{
					contextLine("line1", 1, 1),
					deletedLine("line2", 2),
					addedLine("line2-changed", 2),
					contextLine("line3", 3, 3),
					contextLine("line4", 4, 4),
					contextLine("line5", 5, 5),
				},
			}},
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedNoNewlineAtEOF(t *testing.T) {
	text := `diff --git a/noeof.txt b/noeof.txt
index a1b2c3d..1a2b3c4 123456
--- a/noeof.txt
+++ b/noeof.txt
@@ -1 +1,2 @@
-no trailing newline
\ No newline at end of file
+no trailing newline
+second line, still no eof newline
\ No newline at end of file
`
	deleted := deletedLine("no trailing newline", 1)
	deleted.NoNewlineAtEOF = true
	added := addedLine("second line, still no eof newline", 2)
	added.NoNewlineAtEOF = true
	expect := diff.Diff{Entries: []diff.Entry{
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/noeof.txt b/noeof.txt",
				"index a1b2c3d..1a2b3c4 123456",
				"--- a/noeof.txt",
				"+++ b/noeof.txt",
			},
			OldPath: "noeof.txt",
			NewPath: "noeof.txt",
			Hunks: []diff.Hunk{{
				OldStart: 1,
				OldLines: 1,
				NewStart: 1,
				NewLines: 2,
				Lines: []diff.Line{
					deleted,
					addedLine("no trailing newline", 1),
					added,
				},
			}},
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedMultiFile(t *testing.T) {
	text := `diff --git a/a.txt b/a.txt
index a1b2c3d..1a2b3c4 123456
--- a/a.txt
+++ b/a.txt
@@ -1,5 +1,2 @@
 line1
-line2
-line3
-line4
-line5
+line2-changed
diff --git a/bin.dat b/bin.dat
index a1b2c3d..1a2b3c4 123456
Binary files a/bin.dat and b/bin.dat differ
`
	expect := diff.Diff{Entries: []diff.Entry{
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/a.txt b/a.txt",
				"index a1b2c3d..1a2b3c4 123456",
				"--- a/a.txt",
				"+++ b/a.txt",
			},
			OldPath: "a.txt",
			NewPath: "a.txt",
			Hunks: []diff.Hunk{{
				OldStart: 1,
				OldLines: 5,
				NewStart: 1,
				NewLines: 2,
				Lines: []diff.Line{
					contextLine("line1", 1, 1),
					deletedLine("line2", 2),
					deletedLine("line3", 3),
					deletedLine("line4", 4),
					deletedLine("line5", 5),
					addedLine("line2-changed", 2),
				},
			}},
		}},
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/bin.dat b/bin.dat",
				"index a1b2c3d..1a2b3c4 123456",
				"Binary files a/bin.dat and b/bin.dat differ",
			},
			OldPath: "bin.dat",
			NewPath: "bin.dat",
			Binary:  true,
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedGitLog(t *testing.T) {
	text := `commit a1b2c3d4e5a1b2c3d4e5a1b2c3d4e5a1b2c3d4e5
Author: Jane Deo <test@example.com>
Date:   Mon Aug 17 10:00:00 2026 +0900

    First commit message

diff --git a/a.txt b/a.txt
index a1b2c3d..1a2b3c4 123456
--- a/a.txt
+++ b/a.txt
@@ -1,3 +1,3 @@
 line1
-line2
+line2-changed
 line3
commit 1a2b3c4d5e1a2b3c4d5e1a2b3c4d5e1a2b3c4d5e
Author: Jane Deo <test@example.com>
Date:   Tue Aug 18 11:30:00 2026 +0900

    Second commit message

diff --git a/b.txt b/b.txt
index a1b2c3d..1a2b3c4 123456
--- a/b.txt
+++ b/b.txt
@@ -1,2 +1,2 @@
-old line
+new line
 context line
`
	expect := diff.Diff{Entries: []diff.Entry{
		{Text: "commit a1b2c3d4e5a1b2c3d4e5a1b2c3d4e5a1b2c3d4e5"},
		{Text: "Author: Jane Deo <test@example.com>"},
		{Text: "Date:   Mon Aug 17 10:00:00 2026 +0900"},
		{Text: ""},
		{Text: "    First commit message"},
		{Text: ""},
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/a.txt b/a.txt",
				"index a1b2c3d..1a2b3c4 123456",
				"--- a/a.txt",
				"+++ b/a.txt",
			},
			OldPath: "a.txt",
			NewPath: "a.txt",
			Hunks: []diff.Hunk{{
				OldStart: 1,
				OldLines: 3,
				NewStart: 1,
				NewLines: 3,
				Lines: []diff.Line{
					contextLine("line1", 1, 1),
					deletedLine("line2", 2),
					addedLine("line2-changed", 2),
					contextLine("line3", 3, 3),
				},
			}},
		}},
		{Text: "commit 1a2b3c4d5e1a2b3c4d5e1a2b3c4d5e1a2b3c4d5e"},
		{Text: "Author: Jane Deo <test@example.com>"},
		{Text: "Date:   Tue Aug 18 11:30:00 2026 +0900"},
		{Text: ""},
		{Text: "    Second commit message"},
		{Text: ""},
		{File: &diff.File{
			RawHeader: []string{
				"diff --git a/b.txt b/b.txt",
				"index a1b2c3d..1a2b3c4 123456",
				"--- a/b.txt",
				"+++ b/b.txt",
			},
			OldPath: "b.txt",
			NewPath: "b.txt",
			Hunks: []diff.Hunk{{
				OldStart: 1,
				OldLines: 2,
				NewStart: 1,
				NewLines: 2,
				Lines: []diff.Line{
					deletedLine("old line", 1),
					addedLine("new line", 1),
					contextLine("context line", 2, 2),
				},
			}},
		}},
	}}

	testParseUnified(t, expect, text)
}

func TestParseUnifiedError(t *testing.T) {
	header := "diff --git a/a.txt b/a.txt\n--- a/a.txt\n+++ b/a.txt\n"
	testSet := []struct {
		name string
		text string
	}{
		{"invalid hunk header", header + "@@ -x +1 @@\n line1\n"},
		{"unexpected end of file in hunk", header + "@@ -1,2 +1,2 @@\n line1\n"},
		{"invalid hunk line format", header + "@@ -1 +1 @@\nxline1\n"},
	}

	for _, s := range testSet {
		if _, err := ParseUnified(s.text); err == nil {
			t.Errorf("unexpected nil error: %s", s.name)
		}
	}
}

func testParseUnified(t *testing.T, expect diff.Diff, rawDiff string) {
	t.Helper()
	actual, err := ParseUnified(rawDiff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expect.Entries) != len(actual.Entries) {
		t.Fatalf("unexpected entries length: expected=%d, actual=%d", len(expect.Entries), len(actual.Entries))
	}
	for i := range expect.Entries {
		if !reflect.DeepEqual(expect.Entries[i], actual.Entries[i]) {
			t.Errorf("unexpected entry at %d: expected=%s, actual=%s", i, entryString(expect.Entries[i]), entryString(actual.Entries[i]))
		}
	}
}

func entryString(e diff.Entry) string {
	if e.File != nil {
		return fmt.Sprintf("%+v", *e.File)
	}
	return fmt.Sprintf("%q", e.Text)
}

func contextLine(content string, oldNumber int, newNumber int) diff.Line {
	return diff.Line{
		Kind:      diff.LineContext,
		Content:   content,
		OldNumber: oldNumber,
		NewNumber: newNumber,
	}
}

func addedLine(content string, newNumber int) diff.Line {
	return diff.Line{
		Kind:      diff.LineAdded,
		Content:   content,
		NewNumber: newNumber,
	}
}

func deletedLine(content string, oldNumber int) diff.Line {
	return diff.Line{
		Kind:      diff.LineDeleted,
		Content:   content,
		OldNumber: oldNumber,
	}
}
