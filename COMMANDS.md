# Commands
## Command sequences
Both the CLI and configuration file support command sequencing using spaces or commas. While commands will greedily consume as many arguments as they can accept, a comma serves as an explicit delimiter to end a command's argument list and begin a new one.

### Examples
This command sequence is valid because `swap` accepts up to two arguments (Window IDs). Since two IDs are provided, the parser recognizes `make_active_window_master` as a new command:
```
swap 123 234 make_active_window_master
```

In the example below, the sequence fails because the parser treats `make_active_window_master` as the second argument for `swap`:
```
swap 234 make_active_window_master
```

To resolve this ambiguity, use a comma as an explicit delimiter:
```
swap 234, make_active_window_master
```
The parser now correctly identifies two distinct actions: swapping the active window with window `234`, and then promoting the active window to master.


## List of commands

### Actions

#### tile
Enable tiling for the current workspace

#### untile
Disable tiling for the current workspace

#### make_active_window_master
Makes active window master

#### switch_layout
Cycle through layouts for the current workspace

#### increase_master
Increase number of windows in master column/row

#### decrease_master
Decrease number of windows in master column/row

#### increment_master
Grow width/height of master column/row

#### decrement_master
Shrink width/height of master column/row

#### next_window
Focus of the next window

#### previous_window
Focus of the previous window

#### swap \[WID_A\] WID_B
Swap windows locations in the current layout. If only one WID provided, swap with active window. Do nothing if any of the windows is not on the current workspace.

### Queries 
Queries is prefixed by `query` keyword and mainly useful for scripting. 
Example (will return the layout for current workspace):
```
$ zentile query layout
```

#### query layout
Print layout of the current workspace

### Setters
Setters is prefixed by `set` keyword.

#### set layout LAYOUT_NAME
Sets the layout for current workspace. Setting to `none` will untile the workspace, setting to valid layout will tile it if needed.

