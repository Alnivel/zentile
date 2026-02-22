package daemon

import (
	"slices"
)

type Store struct {
	allowedMasters  int
	masters, slaves []Client
}

func buildStore() *Store {
	return &Store{allowedMasters: 1,
		masters: make([]Client, 0),
		slaves:  make([]Client, 0),
	}
}

func (st *Store) Add(c Client) {
	if len(st.masters) < st.allowedMasters {
		st.masters = append(st.masters, c)
	} else {
		st.slaves = append(st.slaves, c)
	}
}

func (st *Store) Remove(c Client) {
	for i, m := range st.masters {
		if m.window.Id == c.window.Id {
			if len(st.slaves) > 0 {
				st.masters[i] = st.slaves[0]
				st.slaves = st.slaves[1:]
			} else {
				st.masters = removeElement(st.masters, i)
			}
			return
		}
	}

	for i, s := range st.slaves {
		if s.window.Id == c.window.Id {
			st.slaves = removeElement(st.slaves, i)
			return
		}
	}
}

func removeElement(s []Client, i int) []Client {
	return append(s[:i], s[i+1:]...)
}

func (st *Store) IncMaster() {
	if len(st.slaves) > 1 {
		st.allowedMasters = st.allowedMasters + 1
		st.masters = append(st.masters, st.slaves[0])
		st.slaves = st.slaves[1:]
	}
}

func (st *Store) DecreaseMaster() {
	if len(st.masters) > 1 {
		st.allowedMasters = st.allowedMasters - 1
		mlen := len(st.masters)
		st.slaves = append([]Client{st.masters[mlen-1]}, st.slaves...)
		st.masters = st.masters[:mlen-1]
	}
}

func (st *Store) MakeMaster(c Client) bool {
	return st.MakeMasterById(c.Id)
}

func (st *Store) MakeMasterById(id ClientId) bool {
	for i, slave := range st.slaves {
		if slave.Id == id {
			st.masters[0], st.slaves[i] = st.slaves[i], st.masters[0]
			return true
		}
	}
	return false
}

func (st *Store) Swap(this_client Client, that_client Client) bool {
	return st.SwapById(this_client.Id, that_client.Id)
}

func (st *Store) SwapById(thisId ClientId, thatId ClientId) bool {
	var thisSlice []Client = nil
	var thatSlice []Client = nil

	var thisIndex = -1
	var thatIndex = -1
	

	for index, master := range st.masters {
		switch master.Id {
		case thisId:
			thisIndex = index
			thisSlice = st.masters
		case thatId:
			thatIndex = index
			thatSlice = st.masters
		}
	}

	for index, slave := range st.slaves {
		switch slave.Id {
		case thisId:
			thisIndex = index
			thisSlice = st.slaves
		case thatId:
			thatIndex = index
			thatSlice = st.slaves
		}
	}

	if thisSlice == nil || thatSlice == nil {
		return false
	}

	thisSlice[thisIndex], thatSlice[thatIndex] = thatSlice[thatIndex], thisSlice[thisIndex]
	return true
}

func (st *Store) All() []Client {
	return append(st.masters, st.slaves...)
}

func (st *Store) ClientRelative(relativeTo ClientId, offset int) (Client, bool) {
	clients := st.All()
	
	index := slices.IndexFunc(clients, func(c Client) bool {
		return c.Id == relativeTo
	})

	if index == -1 {
		return  Client{}, false
	}

	count := len(clients)
	resultIndex := (count + ((index + offset) % count)) % count

	return clients[resultIndex], true
}
