// fpv-osd decodes the OSD recordings of FPV video systems. For now it reads
// the .osd files made by Walksnail (Avatar) goggles running INAV.
//
//	fpv-osd csv  [options] file.osd   convert to CSV, joining the .srt of the same flight
//	fpv-osd dump [options] file.osd   inspect the raw OSD screens
package main

import (
	"io"
	"os"
)

func main() {
	os.Exit(runMain(os.Args[1:], os.Stdout, os.Stderr))
}

// runMain dispatches to a subcommand and returns the process exit code:
// 0 on success, 1 on a runtime error and 2 on a usage error.
func runMain(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "csv":
		return csvCommand(args[1:], stdout, stderr)
	case "dump":
		return dumpCommand(args[1:], stdout, stderr)
	case "help", "-h", "-help", "--help":
		usage(stdout)
		return 0
	}
	outf(stderr, "error: unknown command %q\n\n", args[0])
	usage(stderr)
	return 2
}

func usage(w io.Writer) {
	out(w, `fpv-osd decodes FPV OSD flight recordings (currently Walksnail/INAV .osd files).

usage:
  fpv-osd csv  [options] file.osd    convert to CSV, joining the .srt of the same flight
  fpv-osd dump [options] file.osd    inspect the raw OSD screens (frames, glyphs, regions)

Run "fpv-osd <command> -h" for the options of a command.
`)
}
