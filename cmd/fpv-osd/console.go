package main

import (
	"fmt"
	"io"
)

// Console output goes to stdout or stderr. When writing a message fails (a
// closed pipe, say) there is nothing useful left to do, so these helpers drop
// the error in this one place instead of at every call site.

func out(w io.Writer, s string) { _, _ = fmt.Fprint(w, s) }

func outf(w io.Writer, format string, args ...any) { _, _ = fmt.Fprintf(w, format, args...) }

func outln(w io.Writer, args ...any) { _, _ = fmt.Fprintln(w, args...) }
