package init

import (
	"net/http"

	"github.com/mycontroller-org/backend/v2/cmd/core/app/handler"
	handlerInit "github.com/mycontroller-org/backend/v2/cmd/core/start_handler"
	"go.uber.org/zap"
)

var HANDLER http.Handler

func InitHandler() {
	if HANDLER == nil {
		httpHandler, err := handler.GetHandler()
		if err != nil {
			zap.L().Fatal("Error on getting handler", zap.Error(err))
		}
		HANDLER = httpHandler
		handlerInit.StartHandler(httpHandler)
		return
	}
	zap.L().Info("handler init service is called multiple times")
}
