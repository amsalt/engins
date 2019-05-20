package engins

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/amsalt/engins/components"
	"github.com/amsalt/log"
)

func Run(component ...components.Component) {
	// register and run components.
	components.Register(component...)
	components.Run()
	log.Infof("engins start success")

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	sig := <-c
	log.Infof("engins closing down (signal: %v)", sig)
	components.Stop()
}
