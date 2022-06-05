package service

import "sync"

var DefaultLibraryLockManager = LibraryLockManager{
	LockLibraryIds: []uint{},
}

type LibraryLockManager struct {
	LockLibraryIds []uint
	sync.Mutex
}

func (m *LibraryLockManager) TryToLock(id uint) bool {
	m.Lock()
	defer m.Unlock()
	for _, lockId := range m.LockLibraryIds {
		if lockId == id {
			return false
		}
	}
	m.LockLibraryIds = append(m.LockLibraryIds, id)
	return true
}
func (m *LibraryLockManager) IsLock(id uint) bool {
	m.Lock()
	defer m.Unlock()
	for _, lockId := range m.LockLibraryIds {
		if lockId == id {
			return true
		}
	}
	return false
}

func (m *LibraryLockManager) UnlockLibrary(id uint) {
	m.Lock()
	defer m.Unlock()
	newList := []uint{}
	for _, libraryId := range m.LockLibraryIds {
		if libraryId != id {
			newList = append(newList, libraryId)
		}
	}
	m.LockLibraryIds = newList
	return
}
