package backend

type ClientId interface {
	Equals(ClientId) bool
	String() string
}

type Client interface {
	Id() ClientId

	Activate()

	Decorate()
	Undecorate()
	DecorDimensions() (width int, height int)

	Maximize()
	Unmaximize()

	MoveResize(x, y, width, height int)
	Restore()

	String() string
}

type workspace interface {
	AddClient(c Client)
	RemoveClient(c Client)

	IsTiling() bool
	Tile()
}

type Tracker[T workspace] interface {
	ParseClientId(string) (ClientId, error)

	Client(id ClientId) (client Client, exists bool)
	ActiveClient() (client Client, exists bool)

	CurentWorkspaceNum() uint
	WorkspaceCount() uint
	Workspace(index uint) T
	ActiveWorkspace() T

	WorkAreaDimensions(num uint) (x, y, width, height int)

	StartTracking()
	Sync()
}

type Keybinder interface {
	Bind(keyStr string, callback func())
}
