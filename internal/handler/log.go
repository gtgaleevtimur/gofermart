package handler

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (c *Controller) log(r *http.Request, msg string) {
	reqID := middleware.GetReqID(r.Context())
	if reqID != "" {
		reqID = "[" + reqID + "]"
	}
	url := fmt.Sprintf(`"%s %s%s%s"`, r.Method, "http://", r.Host, r.URL)
	log.Info().Str("reqID", reqID).
		Str("url", url).
		Msg(msg)
}
