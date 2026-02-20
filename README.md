<p align="center">
  <img src="/docs/zentile-logo.png" alt="zentile logo"/>
</p>

On-demand tiling for Openbox, Xfce and other [EWMH Compliant Window Managers](https://en.m.wikipedia.org/wiki/Extended_Window_Manager_Hints).

### Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Config](#config)
- [Credits](#credits)

### Features
- Workspace based tiling. You can enable tiling in one workspace and leave others untouched.
- Ships with two simple tiling layouts (Vertical & Horizontal)
- Customizable gap between tiling windows.
- Autodetection of panels and docks.

### Installation

Use go install

```
$ go install github.com/Alnivel/zentile/cmd/zentile@latest
```

Or compile from source

```
$ git clone https://github.com/Alnivel/zentile
$ cd zentile
$ go build -o zentile ./cmd/zentile/
```

and then manually move the zentile binary to somewhere in your PATH.

### Usage

Run the command without arguments to start it as daemon
```
$ zentile
```

Send commands to the running instance
```
$ zentile set layout vertical
```
See the full list of commands and more in [`COMMANDS.md`](COMMANDS.md)

### Config

Default Keybinding                                  | Description
----------------------------------------------------|---------------------------------------
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>t</kbd>       | Tile current workspace 
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>u</kbd>       | Untile current workspace
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>s</kbd>       | Cycle through layouts
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>n</kbd>       | Goto next window
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>p</kbd>       | Goto previous window
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>m</kbd>       | Make the active window as master
<kbd>Ctrl</kbd>+<kbd>]</kbd>                        | Increase size of master windows
<kbd>Ctrl</kbd>+<kbd>[</kbd>                        | Decrease size of master windows
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>i</kbd>       | Increase number of master windows
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>d</kbd>       | Decrease number of master windows

The config file is located at `~/.config/zentile/config.toml`

### Credits

Inspired by BurntSushi's [pytyle](https://github.com/BurntSushi/pytyle3).  
This project would not have been possible without [xgbutil](https://github.com/BurntSushi/xgbutil).  
Logo was made with [Logomakr](https://logomakr.com/)

### License

zentile is licensed under the MIT License. See the full license text in [`LICENSE`](LICENSE).
