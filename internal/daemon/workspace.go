package daemon

import (
	"fmt"
	"slices"
)

type Workspace struct {
	isTiling        bool
	activeLayoutNum uint
	layoutOrder     []string
	layouts         map[string]Layout
}

type WorkspaceFactory struct {
}

var defaultLayoutOrder = []string{"vertical", "horizontal", "fullscreen"}

func (wsf WorkspaceFactory) NewWorkspace(tracker Tracker, num uint) *Workspace {
	//TODO: Add settings of layout order from config

	return &Workspace{
		isTiling:    false,
		layoutOrder: defaultLayoutOrder,
		layouts:     wsf.createLayouts(tracker, defaultLayoutOrder, num),
	}
}

func (wsf WorkspaceFactory) createLayouts(tracker Tracker, layoutList []string, workspaceNum uint) map[string]Layout {
	layouts := make(map[string]Layout, len(layoutList))

	for _, name := range layoutList {
		switch name {
		case "vertical":
			layouts[name] = &VerticalLayout{&VertHorz{
				Tracker:      tracker,
				Store:        buildStore(),
				Proportion:   0.5,
				WorkspaceNum: workspaceNum,
			}}
		case "horizontal":
			layouts[name] = &HorizontalLayout{&VertHorz{
				Tracker:      tracker,
				Store:        buildStore(),
				Proportion:   0.5,
				WorkspaceNum: workspaceNum,
			}}
		case "fullscreen":
			layouts[name] = &FullScreen{
				Tracker:      tracker,
				Store:        buildStore(),
				WorkspaceNum: workspaceNum,
			}
		}
	}

	return layouts
}

func (ws *Workspace) SetLayoutByName(layoutName string) error {
	layoutNum := slices.Index(ws.layoutOrder, layoutName)
	if layoutNum == -1 {
		return fmt.Errorf("Failed to set non-existent layout %s", layoutName)
	}

	ws.activeLayoutNum = uint(layoutNum)
	ws.isTiling = true
	ws.ActiveLayout().Do()

	return nil
}

func (ws *Workspace) GetLayoutByName(layoutName string) Layout {
	return ws.layouts[layoutName]
}

func (ws *Workspace) ActiveLayoutName() string {
	return ws.layoutOrder[ws.activeLayoutNum]
}

func (ws *Workspace) ActiveLayout() Layout {
	return ws.layouts[ws.layoutOrder[ws.activeLayoutNum]]
}

// Cycle through the available layouts
func (ws *Workspace) SwitchLayout() {
	ws.activeLayoutNum = (ws.activeLayoutNum + 1) % uint(len(ws.layouts))
	ws.ActiveLayout().Do()
}

// Adds client to all the layouts in a workspace
func (ws *Workspace) AddClient(c Client) {
	for _, l := range ws.layouts {
		l.Add(c)
	}
}

// Removes client from all the layouts in a workspace
func (ws *Workspace) RemoveClient(c Client) {
	for _, l := range ws.layouts {
		l.Remove(c)
	}
}

// Is the workspace tiling
func (ws *Workspace) IsTiling() bool {
	return ws.isTiling
}

// Tiles the active layout in a workspace
func (ws *Workspace) Tile() {
	if ws.isTiling {
		ws.ActiveLayout().Do()
	}
}

// Untiles the active layout in a workspace.
func (ws *Workspace) Untile() {
	ws.isTiling = false
	ws.ActiveLayout().Undo()
}
