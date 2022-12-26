package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// error - обработчик-хелпер, пишущий ошибки.
func (c *Controller) error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	reqID := middleware.GetReqID(r.Context())
	type errorJSON struct {
		Error      string
		StatusCode int
	}
	e := errorJSON{
		Error:      err.Error(),
		StatusCode: statusCode,
	}
	prefix := "[ERROR]"
	if reqID != "" {
		prefix = fmt.Sprintf("[%s] [ERROR]", reqID)
	}
	b, errMarshal := json.Marshal(e)
	if errMarshal != nil {
		msg := fmt.Sprintf("Failed to marshal error - %s, StatusCode: 500", err.Error())
		w.Write([]byte(msg))
		log.Info().Str(prefix, msg)
		return
	}
	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(statusCode)
	w.Write(b)
	log.Info().Str(prefix, string(b))
}
