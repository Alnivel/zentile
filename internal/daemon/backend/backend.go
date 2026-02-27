package backend

type backend interface {
	internalOnly()
}

func NewTrackerFor[T workspace](
	backend backend,
	classesToIgnore []string,
	workspaceFactory WorkspaceFactory[T],
) (
	Tracker[T],
	error,
) {

	switch concreteBackend := backend.(type) {
	case x11Backend:
		return newX11Tracker(concreteBackend, classesToIgnore, workspaceFactory)
	default:
		return nil, nil
	}
}

func NewMainLoopFor(
	backend backend,
) (
	beforeCommandPing chan struct{},
	afterCommandPing chan struct{},
	quitPing chan struct{},
) {

	switch concreteBackend := backend.(type) {
	case x11Backend:
		return newX11MainLoop(concreteBackend)
	default:
		return nil, nil, nil
	}
}

func NewKeybinderFor(
	backend backend,
) (
	Keybinder,
) {

	switch concreteBackend := backend.(type) {
	case x11Backend:
		return newX11Keybinder(concreteBackend)
	default:
		return nil
	}
}
