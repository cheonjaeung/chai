package diff

import (
	"reflect"
	"testing"
)

func TestMyersDiff(t *testing.T) {
	o := "a\nb\na\nb\nc\n"
	n := "b\na\nb\na\nc\n"
	diff := MyersDiff(o, n)
	expect := File{Hunks: []Hunk{{
		OldStart: 1,
		OldLines: 5,
		NewStart: 1,
		NewLines: 5,
		Lines: []Line{
			deletedLine("a", 1),
			contextLine("b", 2, 1),
			contextLine("a", 3, 2),
			contextLine("b", 4, 3),
			addedLine("a", 4),
			contextLine("c", 5, 5),
		},
	}}}

	if !reflect.DeepEqual(expect, diff) {
		t.Errorf("expected=%+v, actual=%+v", expect, diff)
	}
}

func TestMyersDiffEmptyOld(t *testing.T) {
	o := ""
	n := "a\nb\nc\n"
	diff := MyersDiff(o, n)
	expect := File{Hunks: []Hunk{{
		OldStart: 0,
		OldLines: 0,
		NewStart: 1,
		NewLines: 3,
		Lines: []Line{
			addedLine("a", 1),
			addedLine("b", 2),
			addedLine("c", 3),
		},
	}}}

	if !reflect.DeepEqual(expect, diff) {
		t.Errorf("expected=%+v, actual=%+v", expect, diff)
	}
}

func TestMyersDiffEmptyNew(t *testing.T) {
	o := "a\nb\nc\n"
	n := ""
	diff := MyersDiff(o, n)
	expect := File{Hunks: []Hunk{{
		OldStart: 1,
		OldLines: 3,
		NewStart: 0,
		NewLines: 0,
		Lines: []Line{
			deletedLine("a", 1),
			deletedLine("b", 2),
			deletedLine("c", 3),
		},
	}}}

	if !reflect.DeepEqual(expect, diff) {
		t.Errorf("expected=%+v, actual=%+v", expect, diff)
	}
}

func TestMyersDiffBothEmpty(t *testing.T) {
	o := ""
	n := ""
	diff := MyersDiff(o, n)
	expect := File{}

	if !reflect.DeepEqual(expect, diff) {
		t.Errorf("expected=%+v, actual=%+v", expect, diff)
	}
}

func TestMyersDiffTotallyDifferent(t *testing.T) {
	o := "a\nb\nc\nd\n"
	n := "1\n2\n3\n4\n5\n"
	diff := MyersDiff(o, n)
	expect := File{Hunks: []Hunk{{
		OldStart: 1,
		OldLines: 4,
		NewStart: 1,
		NewLines: 5,
		Lines: []Line{
			deletedLine("a", 1),
			deletedLine("b", 2),
			deletedLine("c", 3),
			deletedLine("d", 4),
			addedLine("1", 1),
			addedLine("2", 2),
			addedLine("3", 3),
			addedLine("4", 4),
			addedLine("5", 5),
		},
	}}}

	if !reflect.DeepEqual(expect, diff) {
		t.Errorf("expected=%+v, actual=%+v", expect, diff)
	}
}

func contextLine(content string, oldNumber int, newNumber int) Line {
	return Line{
		Kind:      LineContext,
		Content:   content,
		OldNumber: oldNumber,
		NewNumber: newNumber,
	}
}

func addedLine(content string, newNumber int) Line {
	return Line{
		Kind:      LineAdded,
		Content:   content,
		NewNumber: newNumber,
	}
}

func deletedLine(content string, oldNumber int) Line {
	return Line{
		Kind:      LineDeleted,
		Content:   content,
		OldNumber: oldNumber,
	}
}
