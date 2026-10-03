# lazygdb

Highly opinionated and experimental terminal UI for GDB inspired by the _lazy_ ecosystem of TUI tools.

Heavily inspired by other existing GDB dashboards like [pwndbg](https://github.com/pwndbg/pwndbg) or [gdb-dashboard](https://github.com/cyrus-and/gdb-dashboard).

<p align="center">
  <img src="https://github.com/PawelReich/lazygdb/blob/master/meta/screenshot.png" alt="Poor man's tool screenshot" width="600"/>
</p>

## Requirements

- Go
- `gdb` installed and in `PATH`

## Build

```sh
go build -o lazygdb .
```

## Usage

```sh
./lazygdb --ex "break main" --ex "target remote :3333" --expr "p/x *0xE000ED24" --ignoreregister "f.*" --ignoreregister "v.*"
```

### Supported flags
* `-e`/`--ex` (repeatable) allows to run commands on startup
* `-x`/`--expr` (repeatable) allows to add expressions on startup
* `--debuggdb` enables GDB notification printing
* `--ignoreregister` (repeatable) allows to ignore registers with regex expressions

## Features
 * Source view _with syntax highlighting_
 * _Scrollable_ Command prompt
 * _Also syntax highlighted_ Disassembly view
 * Stacktrace viewer _allowing to show past callsites in the source viewer_
 * Register viewer _with symbol resolving functionality_
 * Unobtrusive _Working directory-based_ command history

## Keys

| Key       | Action                          |
| --------- | ------------------------------- |
| `Ctrl-Q`  | Quit                            |
| `Ctrl-C`  | Interrupt the running program   |
| `Tab`     | Move focus to the next pane     |
| `Shift-Tab`| Move focus to the previous pane|
| `Up/Down` | Scroll command history          |
| `n` | (Expressions) Add expression       |
| `d` | (Expressions) Remove expression       |
| `r` | (Expressions) Change expression       |


