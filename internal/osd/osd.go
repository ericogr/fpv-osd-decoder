// Package osd decodes .osd files recorded by Walksnail (Avatar) goggles/VTX
// running INAV firmware.
//
// Layout (little-endian), obtained by analysing the files:
//
//	header (40 bytes): "INAV" ... u16 columns @0x24, u16 rows @0x26
//	frame: u32 timestamp (ms) + columns*rows u16 values (glyph codes)
package osd

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

const (
	headerSize = 40
	magic      = "INAV"
	colsOffset = 0x24
	rowsOffset = 0x26
)

// Frame is one OSD screen at a point in time.
type Frame struct {
	TimeMs uint32
	Cells  []uint16 // len = Cols*Rows, row by row
}

// File is a whole .osd file loaded in memory.
type File struct {
	Cols, Rows int
	Frames     []Frame
}

// Row returns the cells of row y. The slice aliases the frame's cells.
func (f *File) Row(fr *Frame, y int) []uint16 {
	return fr.Cells[y*f.Cols : (y+1)*f.Cols]
}

// Duration returns the timestamp of the last frame, in milliseconds. A parsed
// File always has at least one frame.
func (f *File) Duration() uint32 {
	return f.Frames[len(f.Frames)-1].TimeMs
}

// Open reads and validates an .osd file.
func Open(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Parse decodes the contents of an .osd file.
func Parse(data []byte) (*File, error) {
	if len(data) < headerSize {
		return nil, errors.New("osd: file is shorter than the header")
	}
	if string(data[:len(magic)]) != magic {
		return nil, fmt.Errorf("osd: unexpected magic %q (want %q)", data[:len(magic)], magic)
	}
	cols := int(binary.LittleEndian.Uint16(data[colsOffset:]))
	rows := int(binary.LittleEndian.Uint16(data[rowsOffset:]))
	if cols == 0 || rows == 0 {
		return nil, fmt.Errorf("osd: invalid grid %dx%d", cols, rows)
	}

	frameSize := 4 + cols*rows*2
	body := data[headerSize:]
	if len(body) == 0 {
		return nil, errors.New("osd: no frames after the header")
	}
	if len(body)%frameSize != 0 {
		return nil, fmt.Errorf("osd: %d bytes after the header are not a multiple of the frame size (%d bytes)", len(body), frameSize)
	}

	f := &File{Cols: cols, Rows: rows, Frames: make([]Frame, len(body)/frameSize)}
	for i := range f.Frames {
		rec := body[i*frameSize : (i+1)*frameSize]
		cells := make([]uint16, cols*rows)
		for j := range cells {
			cells[j] = binary.LittleEndian.Uint16(rec[4+j*2:])
		}
		f.Frames[i] = Frame{TimeMs: binary.LittleEndian.Uint32(rec), Cells: cells}
	}
	return f, nil
}
