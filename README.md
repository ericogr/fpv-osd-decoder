# fpv-osd-decoder

Turns the OSD recordings of FPV video systems into CSV telemetry.

An OSD file does not store numbers. It stores the **screen** the pilot saw,
frame by frame, as a grid of font glyphs. This tool reads those glyphs,
recognizes each field by its icon and unit, and writes one CSV column per
telemetry type it finds. It also joins the `.srt` of the same flight (video
link data) by time.

**Supported today:** `.osd` files recorded by Walksnail (Avatar) goggles while
the flight controller runs **INAV**. Other video systems are not supported yet.

## Build

Needs Go 1.26 or newer (see `go.mod`). `make` is optional.

```sh
make build        # writes bin/fpv-osd
make install      # installs fpv-osd into GOBIN
```

or without `make`: `go build -o fpv-osd ./cmd/fpv-osd`.

## Quick start

```sh
fpv-osd csv AvatarG0031.osd      # writes AvatarG0031.csv, joining AvatarG0031.srt if it exists
fpv-osd dump -stats AvatarG0031.osd
```

One binary, two subcommands:

| command | what it does |
|---|---|
| `fpv-osd csv [options] file.osd` | converts the recording to CSV, joining the `.srt` of the same flight |
| `fpv-osd dump [options] file.osd` | shows the raw OSD screens, for diagnostics; it does not write a CSV |

`fpv-osd <command> -h` lists the options of a command. Exit codes: `0` success,
`1` runtime error (unreadable or invalid file), `2` usage error.

## `csv`: convert a flight to CSV

One row per OSD frame. The header only has the columns that have a value in at
least one frame, and the unit is the suffix of the column name
(`osd_altitude_m`). A value that is not readable (blinking, no GPS, splash or
stats screen) is left **empty**, never `0`. A trimmed excerpt:

```
t_ms,osd_rssi_dbm,osd_lq_pct,osd_batt_cell_v,osd_current_a,osd_altitude_m,osd_flight_mode,srt_latency_ms
449678,-72,100,3.69,6.19,213,CRUZ,33
```

When it finishes, the program prints to stderr every column written, in how
many frames it had a value, and what it means.

| option | effect |
|---|---|
| `-profile name` | what to export: `default` (the CSV with every column and frame) or `maps` (a KML for Google My Maps; see [Profiles](#profiles)) |
| `-o file` | output file (default: the `.osd` name with `.csv`, or `.<profile>.kml` for `maps`; `-` for stdout) |
| `-srt file` | `.srt` to join (default: same name as the `.osd`, if it exists; `none` disables) |
| `-sep c` | CSV column separator (default `,`) |
| `-decimal c` | decimal separator, `.` or `,` (default `.`) |
| `-mode x,y,width` | where the flight mode is shown on the OSD (default `45,16,4`) |
| `-msg-row N` | row of the status messages (default `13`; `-1` disables) |
| `-unmapped` | lists what was left on the screen undecoded |
| `-q` | does not print the column summary |

`-mode` and `-msg-row` exist because the flight mode and the status messages
are text without an anchor icon: they can only be found by position, and that
depends on the layout configured in INAV. The defaults match `AvatarG0031`. For
another layout, run with `-unmapped` and look at what was left on the screen.

### Profiles

A profile decides what goes into the output file. `-profile default`, which is
what you get without the option, writes the CSV with every column of every
frame, as described above.

`-profile maps` does not write a CSV: it writes a **KML** that Google My Maps
imports as a layer with the flight path and the GPS points:

```sh
fpv-osd csv -profile maps AvatarG0031.osd      # writes AvatarG0031.maps.kml
```

- **The path:** the way flight log converters (PX4's `ulog2kml`, ArduPilot's
  `mavkml`) draw it: one line per stretch flown in the same flight mode, each in
  its own colour, named after the mode (`AvatarG0031 CRUZ`). A stretch starts
  where the one before it ended, so there is no gap. ACRO is orange, ANGL blue,
  CRUZ green, RTH red and MANU magenta; any other mode gets a spare colour
  (yellow, cyan, violet, white), and the rows before the first mode is read are
  grey. A mode that blinks out of the OSD keeps the last one read.
- **The altitude** (`osd_altitude_m`) goes into every position and the path is
  `relativeToGround`, because the OSD altitude is measured from the take-off
  point: Google Earth lifts the path off the ground. A reading that is missing
  keeps the last one, and the first rows count as 0. Without an altitude in the
  recording the path hugs the terrain. My Maps is flat and ignores it.
- **The points:** one per position, drawn as small dots without a label
  (Google Earth would otherwise put a pushpin on each one), and named with the
  video time (`16:29`), which shows when you click one, each with the colour of
  its flight mode. The
  description lists, as plain text, `t_ms` and these attributes, with the same
  names and units as in the default profile: `osd_flight_mode`,
  `osd_altitude_m`, `osd_gps_speed_kmh`, `osd_air_speed_kmh`, `osd_heading_deg`,
  `osd_home_dist_m`, `osd_gps_sats`, `osd_rssi_dbm`, `osd_lq_pct`,
  `srt_signal`, `osd_batt_cell_v`, `osd_batt_v` and `osd_current_a`. Only the
  ones the recording has are written. Nothing else is: no attitude, wind,
  throttle or messages.
- Only rows with a **valid position**: both coordinates present, in range, and
  not `0,0`. A position equal to the one before it is dropped, since a point
  that does not move adds nothing to a map.
- **At most 2000 points.** That is My Maps' limit for a spreadsheet; the KML
  keeps it so the points stay as easy to handle. If more are left, they are
  thinned out evenly; the first and the last always stay. The path goes through
  the points that were kept.
- The program fails if the recording has no valid position. `-sep` and
  `-decimal` do not apply to this profile and are refused.

When it finishes, the summary says how many rows were left out and why. In My
Maps, create a map, choose **Import** on a layer and pick the `.kml`. The map is
flat, so the altitude does not raise a point; it is in the description, like
the other attributes.

My Maps does not import everything a KML can hold (folders, `MultiGeometry`,
time elements and HTML balloons are dropped), so the file is a flat list of
plain placemarks. It is not limited to what My Maps reads, though: Google Earth
opens it as well. The 5 MB limit of My Maps for unzipped KML is far from the
size of a flight (about 1 MB).

### Spreadsheets with a Portuguese (Brazil) locale

Google Sheets and Excel with that locale read `3.6` as a date (June 3rd) and
keep `3.63` as text. Write the CSV with a semicolon and a decimal comma:

```sh
fpv-osd csv -sep ';' -decimal ',' AvatarG0031.osd
```

`-decimal ','` only changes numbers; texts such as `osd_message` are left
untouched. When importing, leave the separator on "Detect automatically" or
choose "Semicolon". `make csv` does this for every `.osd` in the folder. Do not
use it for Google My Maps; see the `maps` profile above.

## `dump`: inspect the raw screens

| option | effect |
|---|---|
| `-frame N`, `-n N` | prints `N` frames starting at frame index `-frame` (defaults 0 and 1) |
| `-hex` | shows non-ASCII glyphs by name (`{ALT_M}`) or code |
| `-stats` | grid size, number of frames and duration |
| `-census` | glyphs present, in how many frames and at which positions |
| `-region x,y,width` | distinct values of a strip of cells across all frames |

```sh
fpv-osd dump -frame 3000 -hex AvatarG0031.osd   # one frame, glyphs by name
fpv-osd dump -region 45,16,4 AvatarG0031.osd    # distinct values of a region
fpv-osd dump -census AvatarG0031.osd            # glyphs present and where
```

## Output columns

OSD (`osd_*`, shown by the flight controller):

| column | meaning |
|---|---|
| `osd_rssi_dbm`, `osd_rssi_antenna` | link RSSI (dBm); active antenna (1 = RSSI icon, 2 = 2RSS icon) |
| `osd_rf_mode`, `osd_lq_pct` | CRSF RF mode and link quality (`n:LQ` format) |
| `osd_flight_time_s` | armed flight time (s) |
| `osd_batt_cell_v` / `osd_batt_v` | per-cell / pack voltage (V); a reading above 4.5 V is taken as the pack |
| `osd_current_a`, `osd_mah_used`, `osd_mah_per_km` | current, capacity used, efficiency |
| `osd_throttle_pct`, `osd_throttle_auto` | throttle (%); 1 if controlled by navigation |
| `osd_flight_mode`, `osd_message` | flight mode and status message (text) |
| `osd_heading_deg` | heading (0 = north) |
| `osd_roll_deg`, `osd_pitch_deg` | roll (+ = right wing down) and pitch (+ = nose up) |
| `osd_gforce` | G-force |
| `osd_gps_sats`, `osd_gps_lat`, `osd_gps_lon` | satellites and position |
| `osd_gps_speed_kmh`, `osd_air_speed_kmh` | ground speed (GPS) and airspeed (pitot) |
| `osd_altitude_m`, `osd_home_dist_m`, `osd_trip_dist_m` | altitude, distance to home, distance travelled |
| `osd_home_dir_deg` | direction of home relative to the direction of flight (22.5° steps) |
| `osd_wind_speed_kmh`, `osd_wind_dir_deg` | estimated wind and its direction relative to the nose (45° steps) |

SRT (`srt_*`, the video link data recorded with the flight): `srt_signal`,
`srt_channel`, `srt_flight_time_s`, `srt_sky_batt_v`, `srt_ground_batt_v`,
`srt_latency_ms`, `srt_bitrate_mbps`, `srt_link_dist_m`. A field the tool does
not know becomes `srt_<name>[_<unit>]`.

Two of them are easy to mistake for each other: `srt_sky_batt_v` (`SBat`, "sky
battery") is the voltage of the **aircraft** battery, and `srt_ground_batt_v`
(`GBat`, "ground battery") is the voltage of the battery that powers the
**ground** device with the digital video receiver.

Each OSD frame is joined to the SRT entry whose time interval covers its
`t_ms`.

## How it works

The `.osd` format has no official documentation; this comes from analysing the
files (little-endian):

| part | layout |
|---|---|
| header, 40 bytes | `"INAV"`; `u16` columns at `0x24`; `u16` rows at `0x26` (53 × 20) |
| each frame, `4 + cols*rows*2` bytes | `u32` timestamp in ms, then `cols*rows` `u16` values, one per cell, row by row |

Each `u16` is a code from the INAV OSD font: ASCII for letters, digits and
`- : .`, and dedicated glyphs for icons and units (the `SYM_*` table from
INAV's `src/main/drivers/osd_symbols.h`, in `internal/inav/symbols_gen.go`). A
field is recognized by its glyphs, for example `[HOME] 1.41 [km]` is the
distance to home, so it does not matter where on the screen it is.

Details that matter when reading it:

- **Decimals:** the point is embedded in the digits. `0xA1–0xAA` are `0.`…`9.`
  and `0xB1–0xBA` are `.0`…`.9`; `[3.][.6]9` is `3.69`.
- **Units change with magnitude** (m↔km, ft↔mi, mAh↔Ah, mm:ss↔hh:mm). The
  decoder reads the unit glyph of each frame and converts everything to a
  single unit per column.
- **Arrows** (home direction, wind) are glyphs; they become degrees.
- **Blinking elements** disappear in some frames; they are left empty.
- **Splash and stats screens:** in `AvatarG0031` the first 25 frames and the
  last 20 carry no flight telemetry, so the OSD columns are empty there.

## Limitations

- Validated against one recording (`AvatarG0031`, INAV 9.1.0). Another OSD
  layout may need `-mode` and `-msg-row`, and may show fields the rules do not
  know; `-unmapped` lists them.
- The artificial horizon, the crosshair, the variometer and the splash, stats
  and post-flight screens are not decoded. `AvatarG0031` also shows a constant
  `M0` label whose meaning is unknown.
- The precision is what is on screen, at the OSD refresh rate (~6.6 frames/s).
  Raw flight controller telemetry cannot be recovered.
- The OSD and SRT clocks were checked to agree within about 1 s; the join is
  not verified at millisecond level.
- `symbols_gen.go` was generated from `osd_symbols.h` on INAV's `master`
  branch while the recordings come from firmware 9.1.0; glyphs added or moved
  between versions may require regenerating the table.

## Project layout

```
cmd/fpv-osd         the program: `csv` (OSD [+ SRT] -> CSV) and `dump` (diagnostics)
internal/osd        binary parser for the .osd (Walksnail)
internal/inav       INAV glyph table and number reading
internal/telemetry  decoding rules (anchor glyph + number + unit)
internal/srt        .srt parser and lookup by time (Walksnail)
internal/export     joins OSD telemetry and SRT into a table and writes it as CSV
```

## Development

`make help` lists every target:

| target | what it does |
|---|---|
| `make` / `make build` | builds `fpv-osd` into `bin/` |
| `make csv` | builds if needed and converts every `.osd` in the folder to CSV in the Portuguese-locale format (`;` and `,`); `FILES=a.osd` picks the files, `SEP=, DECIMAL=.` switches to the English format |
| `make run FILE=x.osd ARGS="-sep ';'"` | runs `fpv-osd csv` on `FILE` (default `AvatarG0031.osd`) with extra options; `CMD=dump` runs the other subcommand |
| `make test` / `make cover` | runs the tests / runs them with a coverage summary |
| `make check` | format check + `go vet` + tests |
| `make fmt` / `make vet` / `make tidy` | `gofmt -w`, `go vet`, `go mod tidy` |
| `make install` | installs `fpv-osd` into `GOBIN` |
| `make clean` | removes `bin/` and `coverage.out` |

Recordings (`*.osd`, `*.srt`) and generated `*.csv` files are git-ignored. The
end-to-end tests use `AvatarG0031.osd` when it is in the project folder and are
skipped otherwise.

## License

[MIT](LICENSE). Copies and derivative works must keep the copyright notice and
the license text.

The glyph codes and names in `internal/inav/symbols_gen.go` follow INAV's
`osd_symbols.h` (GPL-3.0-or-later, https://github.com/iNavFlight/inav). Only
the code-to-name table is reproduced; none of INAV's code is included.
