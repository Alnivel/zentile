package backend

import (
	"fmt"
	"strconv"

	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/ewmh"
	"github.com/jezek/xgbutil/motif"
	"github.com/jezek/xgbutil/xrect"
	"github.com/jezek/xgbutil/xwindow"
	log "github.com/sirupsen/logrus"
)

type X11ClientId xproto.Window

func newX11ClientIdFromWid(w xproto.Window) X11ClientId {
	return X11ClientId(w)
}

func parseX11ClientId(str string) (X11ClientId, error) {
	val, err := strconv.ParseUint(str, 0, 32)
	return X11ClientId(val), err
}

func (id X11ClientId) Equals(otherId ClientId) bool {
	xid, isX11ClientId := otherId.(X11ClientId)
	return isX11ClientId && xproto.Window(xid) == xproto.Window(id)
}

func (id X11ClientId) String() string {
	return fmt.Sprintf("%#x", uint32(id))
}

type X11Client struct {
	id           X11ClientId
	window       *xwindow.Window
	workspaceNum uint // Desktop the client is currently in.
	savedProp    Prop // Properties that the client had, before it was tiled.

	X *xgbutil.XUtil
}

type Prop struct {
	Geom       xrect.Rect
	decoration bool
}

func (c X11Client) Id() ClientId {
	return c.id
}

func (c X11Client) name() string {
	name, err := ewmh.WmNameGet(c.X, c.window.Id)
	if err != nil {
		return ""
	}

	return name
}

func (c X11Client) String() string {
	return fmt.Sprintf("'%s' (%#x)", c.name(), uint32(c.id))
}

func (c X11Client) MoveResize(x, y, width, height int) {
	c.Unmaximize()

	dw, dh := c.DecorDimensions()
	err := c.window.WMMoveResize(x, y, width-dw, height-dh)

	if err != nil {
		log.Info("Error when moving ", c.name(), " ", err)
	}
}

// DecorDimensions returns the width and height occupied by window decorations
func (c X11Client) DecorDimensions() (width int, height int) {
	cGeom, err1 := xwindow.RawGeometry(c.X, xproto.Drawable(c.window.Id))
	pGeom, err2 := c.window.DecorGeometry()

	if err1 != nil || err2 != nil {
		return
	}

	width = pGeom.Width() - cGeom.Width()
	height = pGeom.Height() - cGeom.Height()
	return
}

func (c X11Client) Maximize() {
	ewmh.WmStateReq(c.X, c.window.Id, 1, "_NET_WM_STATE_MAXIMIZED_VERT")
	ewmh.WmStateReq(c.X, c.window.Id, 1, "_NET_WM_STATE_MAXIMIZED_HORZ")
}

func (c X11Client) Unmaximize() {
	ewmh.WmStateReq(c.X, c.window.Id, 0, "_NET_WM_STATE_MAXIMIZED_VERT")
	ewmh.WmStateReq(c.X, c.window.Id, 0, "_NET_WM_STATE_MAXIMIZED_HORZ")
}

func (c X11Client) Undecorate() {
	motif.WmHintsSet(c.X, c.window.Id,
		&motif.Hints{
			Flags:      motif.HintDecorations,
			Decoration: motif.DecorationNone,
		})
}

func (c X11Client) Decorate() {
	if !c.savedProp.decoration {
		return
	}

	motif.WmHintsSet(c.X, c.window.Id,
		&motif.Hints{
			Flags:      motif.HintDecorations,
			Decoration: motif.DecorationAll,
		})
}

// Restore resizes and decorates window to pre-tiling state.
func (c X11Client) Restore() {
	c.Decorate()
	geom := c.savedProp.Geom
	log.Info("Restoring ", c.name(), ": ", "X: ", geom.X(), " Y: ", geom.Y())
	c.MoveResize(geom.X(), geom.Y(), geom.Width(), geom.Height())
}

// Activate makes the client the currently active window
func (c X11Client) Activate() {
	ewmh.ActiveWindowReq(c.X, c.window.Id)
}
