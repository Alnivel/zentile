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

### Targeting
Most of the commands operate on `target` workspace and window.
Initially for each command sequence it is the current workspace and the currectly active window, other target can be set using `for` commands - see [context commands](#context_commands).

## List of commands

### Actions

#### tile
Enable tiling for the target workspace

#### untile
Disable tiling for the target workspace

#### make_active_window_master
Makes active window master

#### switch_layout
Cycle through layouts for the target workspace

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
Swap windows locations in the target layout. If only one WID provided, swap with target window. Do nothing if any of the windows is not on the target workspace.

### Queries 
Queries are prefixed by `query` keyword and mainly useful for scripting. 
Example (will return the layout for target workspace):
```
$ zentile query layout
```

#### query layout
Print layout of the target workspace

### Setters
Setters are prefixed by `set` keyword.

#### set layout LAYOUT_NAME
Sets the layout for target workspace. Setting to `none` will untile the workspace, setting to valid layout will tile it if needed.

### Context commands
TODO: Write about what context commands are

#### for workspace WORKSPACE_NUM
Sets target workspace

#### for window WID
Sets target window

