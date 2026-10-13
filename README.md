# FPV OSD Decoder

Turns the OSD recordings of FPV video systems into CSV telemetry.

An OSD file does not store numbers. It stores the **screen** the pilot saw,
frame by frame, as a grid of font glyphs. This tool reads those glyphs,
recognizes each field by its icon and unit, and writes one CSV column per
telemetry type it finds. It reads most of the telemetry elements of INAV 9.1.0
(see [Limitations](#limitations) for the ones it does not), and a number it
finds next to a glyph it does not know still gets a column (see
[Readings no rule knows](#readings-no-rule-knows)). It also joins the `.srt` of
the same flight (video link data) by time.

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
| `-mode x,y,width` | where the flight mode is shown on the OSD (default `auto`: found in the recording) |
| `-msg-row N` | row of the status messages (default `auto`: found in the recording; `-1` disables) |
| `-unmapped` | lists what was left on the screen undecoded (after the readings that got an `osd_raw_*` column) |
| `-q` | does not print the column summary |

The flight mode and the status messages are text without an anchor icon, so no
rule can recognize them by their glyphs, and where they are depends on the
layout the pilot set up in INAV. The program looks for them in the recording
and prints what it found in the `layout:` line of the summary:

```
layout: flight mode at 45,16,4 (found); status messages on row 13 (found)
```

- **The flight mode** is always four characters and always one of a short list
  of names (`ACRO`, `ANGL`, `CRUZ`, `RTH`, `HOLD`...), and it is on screen in
  nearly every frame. The positions where those names show up most often are
  the element. A name inside a longer word does not count, and neither does a
  position seen less than a fifth as often as the best one (the message
  `(HOLD)` is not a layout). A recording with two layouts that show the mode in two places gets
  both.
- **The status messages** are free text, so there is no list to look for. They
  are the text that changes: of what is left on the screen after everything
  else was read, the row that shows the most different texts. It is taken only
  when no other row shows as many, when the row has at least two different
  texts with a word of three letters or more, and when the text is not a plus
  code. A name that never changes (the craft name) is not a message.

When one of them is not found, its column is left out and the `layout:` line
says so. Then give the position yourself: `-mode x,y,width` and `-msg-row N`
override the search. `dump -region x,y,width` and `-unmapped` show where the
text is. The position is the same for the whole recording, so with a layout
switch the messages are only read in the layouts that have them on that row.

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
  its flight mode. The description lists, as plain text, `t_ms` and these
  attributes, with the same names and units as in the default profile: `osd_flight_mode`,
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
| `-census` | glyphs present, in how many frames and at which positions |
| `-region x,y,width` | distinct values of a strip of cells across all frames |

```sh
fpv-osd dump -frame 3000 -hex AvatarG0031.osd   # one frame, glyphs by name
fpv-osd dump -region 45,16,4 AvatarG0031.osd    # distinct values of a region
fpv-osd dump -census AvatarG0031.osd            # glyphs present and where
```

## Output columns

OSD (`osd_*`, shown by the flight controller). The first group was checked
against a recording; the others are read as the INAV 9.1.0 firmware draws them
(see [Limitations](#limitations)).

**Checked against a recording**

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

**Read from the INAV 9.1.0 source**

| column | meaning |
|---|---|
| `osd_rssi_pct` | RSSI as a percentage (the MAX glyph is 100) |
| `osd_lq_down_pct`, `osd_snr_db` | CRSF downlink link quality (%); signal-to-noise ratio (dB) |
| `osd_tx_power_mw`, `osd_rx_power_mw` | power of the uplink transmitter, and of the downlink one on the aircraft (mW) |
| `osd_rx_band`, `osd_rx_mode` | receiver band and mode (text) |
| `osd_vtx_channel`, `osd_vtx_power` | video transmitter band and channel (such as `R1`) and power level (text) |
| `osd_on_time_s`, `osd_rth_time_left_s`, `osd_glide_time_s` | time since power-on; flight time left before the return to home must start; glide time left (s) |
| `osd_clock` | real-time clock, `HH:MM:SS` (text) |
| `osd_batt_pct` | remaining battery charge (%) |
| `osd_wh_used`, `osd_wh_per_km`, `osd_power_w`, `osd_impedance_mohm` | energy used (Wh), efficiency (Wh/km), power (W), supply impedance (mOhm) |
| `osd_gps_hdop` | GPS horizontal dilution of precision |
| `osd_ground_course_deg`, `osd_azimuth_deg`, `osd_home_heading_error_deg` | GPS course over the ground; azimuth of the aircraft seen from home; angle between the direction of flight and the direction of home |
| `osd_rth_dist_left_m`, `osd_glide_range_m`, `osd_cross_track_error_m` | distance that can still be flown before the return to home must start; glide range; cross-track error (m) |
| `osd_fence_dist_h_m`, `osd_fence_dist_v_m` | horizontal and vertical distance to the nearest geofence zone (m; vertical is negative below it) |
| `osd_odometer_km`, `osd_rangefinder_m`, `osd_glideslope` | distance travelled in all flights (km); rangefinder (m); glide ratio |
| `osd_3d_speed_kmh`, `osd_vspeed_ms` | 3D speed (km/h); vertical speed (m/s, + = climbing) |
| `osd_gps_max_speed_kmh`, `osd_air_max_speed_kmh`, `osd_3d_max_speed_kmh`, `osd_min_ground_speed_kmh` | maximum ground speed, airspeed and 3D speed of the flight; minimum ground speed set for navigation (km/h) |
| `osd_wind_vspeed_kmh` | estimated vertical wind (km/h, + = up) |
| `osd_imu_temp_c`, `osd_baro_temp_c`, `osd_esc_temp_c`, `osd_temp_sensor_c` | temperatures (Celsius; Fahrenheit is converted) |
| `osd_esc_rpm` | motor speed reported by the ESC (rpm) |
| `osd_gforce_x`, `osd_gforce_y`, `osd_gforce_z` | G-force on each axis |
| `osd_climb_eff_mah_per_m` | climb efficiency, as the OSD shows it |
| `osd_profile`, `osd_blackbox_log` | active configuration profile; number of the blackbox log |
| `osd_gvar_0`, `osd_gvar_1`, `osd_gvar_2`, `osd_gvar_3` | global variables of the INAV programming framework |

`osd_mah_per_km` also reads the efficiency per mile and per nautical mile (in
mAh) and per km in Ah, which INAV shows depending on its unit setting, and
converts it to mAh/km. Ah per mile and per nautical mile are not read (see
[Limitations](#limitations)).
Distances, altitudes and speeds in miles, feet or knots are converted to
metres, kilometres per hour and so on as well.

In `AvatarG0031`, four of these appear, in the last frames before the
statistics (the layout INAV shows after disarming has them): `osd_imu_temp_c`
(40), `osd_baro_temp_c` (42), `osd_on_time_s` (20:59) and `osd_odometer_km`
(10.3, consistent with the 10.35 km of the statistics that follow). They match
the screen, which is the only check the other columns of this group have had so
far.

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

### Readings no rule knows

A number next to a glyph that no rule recognizes still becomes a column. Its
name is `osd_raw_` and the names of the glyphs around the number, as they are in
INAV's `osd_symbols.h`, in lowercase and joined by double underscores (a name has
single ones inside it, so two glyphs side by side cannot be mistaken for one
glyph): `{HDP_L}{HDP_R}1.2` would be `osd_raw_hdp_l__hdp_r`, and `35{TEMP_C}`,
`osd_raw_temp_c`. The value is
what is on screen, with no unit conversion and no rounding. These columns come
after the OSD ones and are sorted by name.

It reads only what it can trust:

- a number with a glyph right before it (an icon, perhaps with blanks between)
  or right after it (a unit). A bare number, a label, or text is not taken: there
  is no telling what it is, and `-unmapped` lists them.
- not on the INAV splash screen shown while arming (all it holds is totals and
  the version) and not on the statistics screen (see [Limitations](#limitations)).
- one column per name. If the same glyphs show a number twice on the screen, the
  first (top to bottom, left to right) keeps the column and the other is listed
  by `-unmapped`.

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
distance to home, so it does not matter where on the screen it is. The only
exceptions are the flight mode and the status message, which are text with no
glyph (see [`csv`](#csv-convert-a-flight-to-csv)).

Details that matter when reading it:

- **Decimals:** the point is embedded in the digits. `0xA1–0xAA` are `0.`…`9.`
  and `0xB1–0xBA` are `.0`…`.9`; `[3.][.6]9` is `3.69`.
- **Units change with magnitude** (m↔km, ft↔mi, mAh↔Ah, mm:ss↔hh:mm). The
  decoder reads the unit glyph of each frame and converts everything to a
  single unit per column.
- **Arrows** (home direction, wind) are glyphs; they become degrees.
- **Blinking elements** disappear in some frames; they are left empty. A wind
  estimate that is not valid yet shows a `*` where the unit goes, and is treated
  the same way.
- **Two elements that look alike:** some elements are drawn in exactly the same
  way (see [Limitations](#limitations)), so the glyphs cannot say which is
  which. The first one on the screen keeps the column. The other one is not
  lost when it is a number next to glyphs, as most are: it goes to an
  `osd_raw_*` column.
- **Other screens:** INAV draws a splash screen while arming (with its logo) and
  the flight statistics after disarming. The splash screen is read like any
  other, since it shows the position, but nothing is taken from it by the generic
  pass. The statistics screen (its title is `*** STATS`) is not read at all: its
  numbers are summaries of the flight, not what the aircraft was doing then.
  Neither of the two has a flight mode or a status message. In `AvatarG0031`,
  the first 25 frames are the splash screen, which carries the position, and the
  last 20 are the statistics.
- **Speed:** each rule would have to run on every row of every frame, so the
  decoder first checks which glyphs a row holds and skips the rules that need
  one it does not have. A test checks that this never changes what is read.

## Limitations

- **What was checked.** The first group of columns was checked against one
  recording (`AvatarG0031`, INAV 9.1.0). The second group was written from the
  format strings of `src/main/io/osd.c` in INAV 9.1.0 and is tested with frames
  built from those strings, not with recordings; only four of its columns have
  appeared in the sample flight (see [Output columns](#output-columns)). Expect
  to adjust them when a recording with those elements turns up.
- **Layouts.** The OSD of the sample has more than one layout (the one shown when
  armed and the one shown when disarmed), and a recording may have others. Every
  element but two is found wherever the pilot put it. The position of the flight
  mode and of the status messages is found in the recording (see
  [`csv`](#csv-convert-a-flight-to-csv)); the search for the messages is a
  heuristic, which prefers to find nothing over a wrong row. It can fail when
  the messages never change in the recording, when another text changes as often
  as they do (a custom element, say), or when the mode is drawn in some way the
  list of names does not cover, for example by another INAV version. The
  message is the text the rules left on its row, so another element on that row
  is read as itself and is not part of it; the other side of this is that a
  piece of text that a rule takes for a reading (`CH:`, `G1:`, a clock time) is
  not part of the message either.
- **Glyphs that look like text.** A few glyphs of the font sit on the codes of
  ASCII punctuation, so the same characters in a text (the craft name is free
  text) can look like a reading. The Ah per km glyph is a quote: a digit followed
  by a quote in a text is read as an efficiency. The Ah per mile and per nautical
  mile glyphs (`$` and `?`) are not read for that reason, and the transmitter
  power glyph (an apostrophe) is only read at the start of a word.
- **Elements that cannot be told apart.** INAV draws these in the same way, so
  only the first one on the screen is read (see above):
  - capacity used and capacity left (mAh or Ah), and energy used and energy left (Wh);
  - the main battery voltage and the sag-compensated one, for the pack and for the cell;
  - the altitude and the altitude above sea level;
  - the throttle and the scaled throttle;
  - the current draw and the power limiter's current limit, the power and the
    power limiter's power limit;
  - the heading, the course hold error and the course hold adjustment;
  - the direction of home and the direction of the nearest geofence zone;
  - an ESC temperature and the temperature sensor that uses the second icon.
- **Not read** (use `-unmapped` to see them on the screen):
  - values that are not telemetry: the PIDs, rates, expo and the other settings
    the OSD can show, and the version;
  - graphics: the artificial horizon, the crosshair, the variometer, the map, the
    radar, the heading graph and the throttle gauge;
  - text with no glyph to anchor it, which only a position can find: the craft
    and pilot names, the RC source, the plus code, the mission and multi-function
    messages, the switch indicators, and the text of the custom elements. Only
    the flight mode and the status message are read this way, by the position
    found in the recording or given with `-mode` and `-msg-row`;
  - the ADS-B warning, the formation flight element and the power limiter's burst time.
- **Splash and statistics screens.** See [How it works](#how-it-works).
  `AvatarG0031` also shows a constant `M0` label whose meaning is unknown.
- **Precision.** It is what is on screen, at the OSD refresh rate (~6.6
  frames/s). Raw flight controller telemetry cannot be recovered.
- **Clocks.** The OSD and SRT clocks were checked to agree within about 1 s; the
  join is not verified at millisecond level.
- **Glyph table.** `symbols_gen.go` was generated from `osd_symbols.h` on INAV's
  `master` branch. That header is byte for byte the same in the 9.0.0 and 9.1.0
  releases and in `master` (checked on 2026-10-12), and the table has every one
  of its glyphs. An older INAV, or a later one that adds glyphs, may need the table
  regenerated.

## Project layout

```
cmd/fpv-osd         the program: `csv` (OSD [+ SRT] -> CSV) and `dump` (diagnostics)
internal/osd        binary parser for the .osd (Walksnail)
internal/inav       INAV glyph table, number reading and the names of the flight modes
internal/telemetry  decoding rules by subject (anchor glyph + number + unit), unit tables, the generic pass, and the search for the flight mode and the messages
internal/srt        .srt parser and lookup by time (Walksnail)
internal/export     joins OSD telemetry and SRT into a table and writes it as CSV or KML
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
