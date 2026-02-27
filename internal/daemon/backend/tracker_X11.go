package backend

import (
	"slices"
	"strings"

	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/ewmh"
	"github.com/jezek/xgbutil/icccm"
	"github.com/jezek/xgbutil/motif"
	"github.com/jezek/xgbutil/xevent"
	"github.com/jezek/xgbutil/xprop"
	"github.com/jezek/xgbutil/xwindow"
	log "github.com/sirupsen/logrus"
)

type X11Tracker[T workspace] struct {
	X               *xgbutil.XUtil
	classesToIgnore []string

	clients      map[xproto.Window]X11Client
	activeClient xproto.Window // Current Active window

	currentWorkpaceNum uint // Current Desktop
	workspaceCount     uint // Number of desktop workspaces.
	workspaces         map[uint]T

	workArea []ewmh.Workarea
}

type Workarea struct {
	X, Y          int
	Width, Height uint
}

type WorkspaceFactory[T workspace] func(tracker Tracker[T], num uint) T

func newX11Tracker[WorkspaceT workspace](backend x11Backend, classesToIgnore []string, workspaceFactory WorkspaceFactory[WorkspaceT]) (*X11Tracker[WorkspaceT], error) {
	X := backend.X

	workspaceCount, err := ewmh.NumberOfDesktopsGet(X)
	if err != nil {
		return nil, err
	}

	activeWin, err := ewmh.ActiveWindowGet(X)
	if err != nil {
		// It's possible to not have an active window
		log.Info(err)
	}

	currentWorkspace, err := ewmh.CurrentDesktopGet(X)
	if err != nil {
		return nil, err
	}

	workArea, err := ewmh.WorkareaGet(X)
	if err != nil {
		return nil, err
	}

	tracker := X11Tracker[WorkspaceT]{
		X:               X,
		classesToIgnore: classesToIgnore,

		clients:    make(map[xproto.Window]X11Client),
		workspaces: make(map[uint]WorkspaceT),

		workspaceCount:     workspaceCount,
		activeClient:       activeWin,
		currentWorkpaceNum: currentWorkspace,
		workArea:           workArea,
	}

	for i := range workspaceCount {
		tracker.workspaces[i] = workspaceFactory(&tracker, i)
	}

	return &tracker, nil
}

/* Public methods */
func (tr *X11Tracker[T]) StartTracking() {
	win := xwindow.New(tr.X, tr.X.RootWin())
	win.Listen(xproto.EventMaskPropertyChange)
	xevent.PropertyNotifyFun(tr.onPropertyNotify).Connect(tr.X, tr.X.RootWin())
	tr.updateClients()
}

func (tr *X11Tracker[T]) ParseClientId(str string) (ClientId, error) {
	return parseX11ClientId(str)
}

func (tr *X11Tracker[T]) Client(id ClientId) (client Client, exists bool) {
	xid := id.(X11ClientId) // Catch fire and explode if ClientId is not X11ClientId

	client, exists = tr.clients[xproto.Window(xid)]
	return client, exists
}

func (tr *X11Tracker[T]) ActiveClient() (client Client, exists bool) {
	client, exists = tr.clients[tr.activeClient]
	return client, exists
}

func (tr *X11Tracker[T]) CurentWorkspaceNum() uint {
	return tr.currentWorkpaceNum
}

func (tr *X11Tracker[T]) WorkspaceCount() uint {
	return tr.workspaceCount
}

func (tr *X11Tracker[T]) Workspace(index uint) T {
	return tr.workspaces[index]
}

func (tr *X11Tracker[T]) ActiveWorkspace() T {
	return tr.workspaces[tr.currentWorkpaceNum]
}

// WorkAreaDimensions returns the dimension of the requested workspace.
func (tr *X11Tracker[T]) WorkAreaDimensions(num uint) (x, y, width, height int) {
	w := tr.workArea[num]
	return w.X, w.Y, int(w.Width), int(w.Height)
}

func (tr *X11Tracker[T]) Sync() {
	tr.X.Conn().Sync()
}

/* Private methods */

func (tr *X11Tracker[T]) onPropertyNotify(X *xgbutil.XUtil, e xevent.PropertyNotifyEvent) {
	var err error
	aname, _ := xprop.AtomName(X, e.Atom)
	switch {
	case aname == "_NET_ACTIVE_WINDOW":
		tr.activeClient, err = ewmh.ActiveWindowGet(X)
	case aname == "_NET_CURRENT_DESKTOP":
		tr.currentWorkpaceNum, err = ewmh.CurrentDesktopGet(X)
	case aname == "_NET_NUMBER_OF_DESKTOPS":
		tr.workspaceCount, err = ewmh.NumberOfDesktopsGet(X)
	case aname == "_NET_WORKAREA":
		tr.workArea, err = ewmh.WorkareaGet(X)
	case aname == "_NET_CLIENT_LIST_STACKING":
		tr.workArea, err = ewmh.WorkareaGet(X)
		tr.updateClients()
		tr.workspaces[tr.currentWorkpaceNum].Tile()
	}

	if err != nil {
		log.Warn("Error updating state: ", err)
	}
}

// updateClients updates the list of tracked clients with the most up to date list of clients.
func (tr *X11Tracker[T]) updateClients() {
	clientList, _ := ewmh.ClientListStackingGet(tr.X)

	for _, wid := range clientList {
		c := tr.getClass(wid)
		if tr.isWindowHidden(wid) || tr.shouldIgnore(c.Class) {
			continue
		}

		tr.startTrackingWindow(wid)
	}

	// Remove tracking of windows not in client list
	for trackedWid := range tr.clients {
		found := slices.Contains(clientList, trackedWid)

		if !found {
			tr.stopTrackingWindow(trackedWid)
		}
	}

}

func (tr *X11Tracker[T]) newClient(wid xproto.Window) X11Client {
	win := xwindow.New(tr.X, wid)

	workspaceNum, err := ewmh.WmDesktopGet(tr.X, wid)
	if err != nil {
		workspaceNum = tr.currentWorkpaceNum
	}

	savedGeom, err := win.DecorGeometry()
	if err != nil {
		log.Info(err)
	}

	var hasDecoration bool
	mh, err := motif.WmHintsGet(tr.X, wid)
	if err != nil {
		hasDecoration = true
	} else {
		hasDecoration = motif.Decor(mh)
	}

	return X11Client{
		id:           newX11ClientIdFromWid(wid),
		window:       win,
		workspaceNum: workspaceNum,
		savedProp: Prop{
			Geom:       savedGeom,
			decoration: hasDecoration,
		},

		X: tr.X,
	}
}

// isWindowHidden returns true if the window has been minimized.
func (tr *X11Tracker[T]) isWindowHidden(w xproto.Window) bool {
	states, _ := ewmh.WmStateGet(tr.X, w)
	return slices.Contains(states, "_NET_WM_STATE_HIDDEN")
}

func (tr *X11Tracker[T]) getClass(w xproto.Window) *icccm.WmClass {
	c, err := icccm.WmClassGet(tr.X, w)
	if err != nil {
		log.Warn(err)
	}

	return c
}

func (tr *X11Tracker[T]) shouldIgnore(clientClass string) bool {
	isClassToIgnore := func(class string) bool {
		return strings.EqualFold(clientClass, class)
	}
	return slices.ContainsFunc(tr.classesToIgnore, isClassToIgnore)
}

/* Client tracking */

func (tr *X11Tracker[T]) IsTracked(wid xproto.Window) bool {
	_, tracked := tr.clients[wid]
	return tracked
}

// Adds window to Tracked Clients and layouts.
func (tr *X11Tracker[T]) startTrackingWindow(wid xproto.Window) {
	if tr.IsTracked(wid) {
		return
	}

	c := tr.newClient(wid)
	if c.workspaceNum > tr.workspaceCount {
		return
	}
	tr.attachHandlers(&c)

	tr.clients[c.window.Id] = c
	ws := tr.workspaces[c.workspaceNum]
	ws.AddClient(c)

}

func (tr *X11Tracker[T]) stopTrackingWindow(wid xproto.Window) {
	c, ok := tr.clients[wid]
	if ok {
		ws := tr.workspaces[c.workspaceNum]
		ws.RemoveClient(c)
		xevent.Detach(tr.X, wid)
		delete(tr.clients, wid)
	}
}

/* Client handlers */

func (tr *X11Tracker[T]) attachHandlers(c *X11Client) {
	c.window.Listen(xproto.EventMaskPropertyChange)

	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		if aname, _ := xprop.AtomName(tr.X, ev.Atom); aname == "_NET_WM_STATE" {
			tr.handleMinimizedClient(c)
		}
	}).Connect(tr.X, c.window.Id)

	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		if aname, _ := xprop.AtomName(tr.X, ev.Atom); aname == "_NET_WM_DESKTOP" {
			tr.handleDesktopChange(c)
		}
	}).Connect(tr.X, c.window.Id)
}

func (tr *X11Tracker[T]) handleMinimizedClient(c *X11Client) {
	states, _ := ewmh.WmStateGet(tr.X, c.window.Id)
	for _, state := range states {
		if state == "_NET_WM_STATE_HIDDEN" {
			tr.workspaces[c.workspaceNum].RemoveClient(*c)
			tr.stopTrackingWindow(c.window.Id)
			tr.workspaces[c.workspaceNum].Tile()
		}
	}
}

func (tr *X11Tracker[T]) handleDesktopChange(c *X11Client) {
	newWorkspaceNum, _ := ewmh.WmDesktopGet(tr.X, c.window.Id)
	oldWorkspaceNum := c.workspaceNum

	tr.workspaces[oldWorkspaceNum].RemoveClient(*c)
	tr.workspaces[newWorkspaceNum].AddClient(*c)

	c.workspaceNum = newWorkspaceNum
	if tr.workspaces[oldWorkspaceNum].IsTiling() {
		tr.workspaces[oldWorkspaceNum].Tile()
	}

	if tr.workspaces[newWorkspaceNum].IsTiling() {
		tr.workspaces[newWorkspaceNum].Tile()
	} else {
		c.Restore()
	}
}
