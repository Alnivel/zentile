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

func (st *Store) Add(cleint Client) {
	if len(st.masters) < st.allowedMasters {
		st.masters = append(st.masters, cleint)
	} else {
		st.slaves = append(st.slaves, cleint)
	}
}

func (st *Store) Remove(client Client) {
	for i, m := range st.masters {
		if m == client {
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
		if s == client {
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
	for i, slave := range st.slaves {
		if slave == c {
			st.masters[0], st.slaves[i] = st.slaves[i], st.masters[0]
			return true
		}
	}
	return false
}

func (st *Store) Swap(this Client, that Client) bool {
	var thisSlice []Client = nil
	var thatSlice []Client = nil

	var thisIndex = -1
	var thatIndex = -1
	

	for index, master := range st.masters {
		switch master {
		case this:
			thisIndex = index
			thisSlice = st.masters
		case that:
			thatIndex = index
			thatSlice = st.masters
		}
	}

	for index, slave := range st.slaves {
		switch slave {
		case this:
			thisIndex = index
			thisSlice = st.slaves
		case that:
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

func (st *Store) ClientRelative(relativeTo Client, offset int) (Client, bool) {
	clients := st.All()
	
	index := slices.Index(clients, relativeTo)

	if index == -1 {
		return  nil, false
	}

	count := len(clients)
	resultIndex := (count + ((index + offset) % count)) % count

	return clients[resultIndex], true
}
