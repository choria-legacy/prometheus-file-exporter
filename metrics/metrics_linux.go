package metrics

import (
	"github.com/rjeczalik/notify"
)

func (m *Metrics) startNotify() error {
	return notify.Watch(m.path, m.events, notify.InDelete, notify.InCloseWrite, notify.InMovedTo)
}
