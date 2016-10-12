package shadowsocks

import (
	"log"
	"net"
	"sync"
)

type PortListener struct {
	password string
	listener net.Listener
}

type PasswdManager struct {
	sync.Mutex
	PortListener map[string]*PortListener
}

func (pm *PasswdManager) Add(port, password string, listener net.Listener) {
	pm.Lock()
	pm.PortListener[port] = &PortListener{password, listener}
	pm.Unlock()
}

func (pm *PasswdManager) Get(port string) (pl *PortListener, ok bool) {
	pm.Lock()
	pl, ok = pm.PortListener[port]
	pm.Unlock()
	return
}

func (pm *PasswdManager) Del(port string) {
	pl, ok := pm.Get(port)
	if !ok {
		return
	}
	pl.listener.Close()
	pm.Lock()
	delete(pm.PortListener, port)
	pm.Unlock()
}

func (pm *PasswdManager) UpdatePortPasswd(port, password string, auth bool) {
	pl, ok := pm.Get(port)
	if !ok {
		log.Printf("new port %s added\n", port)
	} else {
		if pl.password == password {
			return
		}
		log.Printf("closing port %s to update password\n", port)
		pl.listener.Close()
	}
	// run will add the new port listener to passwdManager.
	// So there maybe concurrent access to passwdManager and we need lock to protect it.
	// go run(port, password, auth)
}
