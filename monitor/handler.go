package monitor

import (
	"fmt"
	"strings"

	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
)

type TextHandler struct {
	*core.DefaultInboundHandler
}

func NewTextHandler() *TextHandler {
	return &TextHandler{DefaultInboundHandler: core.NewDefaultInboundHandler()}
}

func (t *TextHandler) OnRead(ctx *core.ChannelContext, msg interface{}) {
	cmd := t.filter(msg)
	metric := getMetric(cmd)

	if metric == nil {
		metric = getMetric("help")
	}

	result := metric.run()
	ctx.Write(fmt.Sprintf("%s\n", result))
}

func (t *TextHandler) filter(text interface{}) string {
	log.Debugf("filter msg %+v", text)
	str, ok := text.(string)
	if ok {
		result := strings.Trim(str, " ")
		result = strings.Replace(result, "\n", "", -1)
		result = strings.Replace(result, "\t", "", -1)
		result = strings.Replace(result, "\r", "", -1)
		result = strings.Replace(result, "\br", "", -1)
		return result
	}
	return str
}
