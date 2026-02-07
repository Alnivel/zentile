package daemon

import (
	"fmt"
	"slices"

	"github.com/Alnivel/zentile/internal/daemon/state"
)

type Workspace struct {
	IsTiling        bool
	activeLayoutNum uint
	layoutOrder     []string
	layouts         map[string]Layout
}

func CreateWorkspaces() map[uint]*Workspace {
	workspaces := make(map[uint]*Workspace)
	defaultLayoutOrder := []string{"vertical", "horizontal", "fullscreen"}

	for i := uint(0); i < state.DeskCount; i++ {
		ws := Workspace{
			IsTiling:    false,
			layoutOrder: defaultLayoutOrder,
			layouts:     createLayouts(defaultLayoutOrder, i),
		}

		workspaces[i] = &ws
	}

	return workspaces
}

func createLayouts(layoutList []string, workspaceNum uint) map[string]Layout {
	layouts := make(map[string]Layout, len(layoutList))

	for _, name := range layoutList {
		switch name {
		case "vertical":
			layouts[name] = &VerticalLayout{&VertHorz{
				Store:        buildStore(),
				Proportion:   0.5,
				WorkspaceNum: workspaceNum,
			}}
		case "horizontal":
			layouts[name] = &HorizontalLayout{&VertHorz{
				Store:        buildStore(),
				Proportion:   0.5,
				WorkspaceNum: workspaceNum,
			}}
		case "fullscreen":
			layouts[name] = &FullScreen{
				Store:        buildStore(),
				WorkspaceNum: workspaceNum,
			}
		}
	}

	return  layouts
}

func (ws *Workspace) SetLayoutByName(layoutName string) error {
	layoutNum := slices.Index(ws.layoutOrder, layoutName)
	if layoutNum == -1 {
		return fmt.Errorf("Failed to set non-existent layout %s", layoutName)
	}

	ws.activeLayoutNum = uint(layoutNum)
	ws.IsTiling = true
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

// Tiles the active layout in a workspace
func (ws *Workspace) Tile() {
	if ws.IsTiling {
		ws.ActiveLayout().Do()
	}
}

// Untiles the active layout in a workspace.
func (ws *Workspace) Untile() {
	ws.IsTiling = false
	ws.ActiveLayout().Undo()
}

func (ws *Workspace) printStore() {
	l := ws.ActiveLayout()
	st := l.sto()
	fmt.Println("Number of masters is ", len(st.masters))
	fmt.Println("Number of slaves is", len(st.slaves))

	for i, c := range st.masters {
		fmt.Println("master ", " ", i, " - ", c.name())
	}

	for i, c := range st.slaves {
		fmt.Println("slave ", " ", i, " - ", c.name())
	}
}
