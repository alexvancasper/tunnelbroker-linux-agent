package doer

import (
	"encoding/json"
	"os/exec"
	"sync"

	"github.com/alexvancasper/TunnelBroker/agent/pkg/models"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Log *logrus.Logger
}

type NewTunnel struct {
	TunnelName         string
	IPv4Remote         string
	IPv4Local          string
	IPv6ServerEndpoint string
	PD                 string
}

func (h Handler) AddTunnel(wg *sync.WaitGroup, data []byte) {
	defer wg.Done()
	l := h.Log.WithFields(logrus.Fields{
		"function": "AddTunnel",
	})
	var tun models.Tunnel
	err := json.Unmarshal(data, &tun)
	if err != nil {
		l.Errorf("Tunnel unmarshalling error: %s", err)
	}
	ExecAddCmd(tun, h.Log)

}

func (h Handler) DeleteTunnel(wg *sync.WaitGroup, data []byte) {
	defer wg.Done()
	l := h.Log.WithFields(logrus.Fields{
		"function": "DeleteTunnel",
	})
	var tun models.Tunnel
	err := json.Unmarshal(data, &tun)
	if err != nil {
		l.Errorf("Tunnel unmarshalling error: %s", err)
	}
	l.Debugf("Tunnel delete info: %v\n", tun)
	ExecDeleteCmd(tun, h.Log)

}

func ExecAddCmd(tun models.Tunnel, log *logrus.Logger) {
	l := log.WithFields(logrus.Fields{
		"function": "ExecAddCmd",
	})
	c := exec.Command("/sbin/ip", "tunnel", "add", tun.TunnelName, "mode", "sit", "remote", tun.IPv4Remote, "local", tun.IPv4Local)
	l.Debugf("cmd1: %s", c.String())
	err := c.Run()
	if err != nil {
		l.Error(err)
		return
	}
	c = exec.Command("/sbin/ip", "link", "set", tun.TunnelName, "up")
	l.Debugf("cmd2: %s", c.String())
	err = c.Run()
	if err != nil {
		l.Error(err)
		return
	}
	c = exec.Command("/sbin/ip", "addr", "add", tun.IPv6ServerEndpoint, "dev", tun.TunnelName)
	l.Debugf("cmd3: %s", c.String())
	err = c.Run()
	if err != nil {
		l.Error(err)
		return
	}
	c = exec.Command("/sbin/ip", "-6", "route", "add", tun.PD, "dev", tun.TunnelName)
	l.Debugf("cmd4: %s", c.String())
	err = c.Run()
	if err != nil {
		l.Error(err)
		return
	}
}

func ExecDeleteCmd(tun models.Tunnel, log *logrus.Logger) {
	l := log.WithFields(logrus.Fields{
		"function": "ExecDeleteCmd",
	})
	c := exec.Command("/sbin/ip", "-6", "route", "del", tun.PD, "dev", tun.TunnelName)
	l.Debugf("cmd1: %s", c.String())
	err := c.Run()
	if err != nil {
		l.Error(err)
		return
	}

	c = exec.Command("/sbin/ip", "tunnel", "del", tun.TunnelName)
	l.Debugf("cmd2: %s", c.String())
	err = c.Run()
	if err != nil {
		l.Error(err)
		return
	}
}
