package backend

import (
	"fmt"

	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/ewmh"
)

type x11Backend struct {
	X *xgbutil.XUtil
}

func (x11Backend) internalOnly() {}

func NewX11Backend() (backend, error) {
	X, err := xgbutil.NewConn()
	if err != nil {
		return nil, err
	}

	_, err = ewmh.GetEwmhWM(X)
	if err != nil {
		return nil, fmt.Errorf("Window manager is not EWMH complaint: %w", err)
	}

	return x11Backend{X}, nil
}
