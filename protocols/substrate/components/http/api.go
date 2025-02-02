package http

import (
	"fmt"
	"time"

	goHttp "net/http"

	iface "github.com/taubyte/go-interfaces/services/substrate/components/http"
	http "github.com/taubyte/http"
	"github.com/taubyte/tau/protocols/substrate/components/http/common"
	"github.com/taubyte/tau/vm/counter"
	"github.com/taubyte/tau/vm/helpers"
	"github.com/taubyte/tau/vm/lookup"
)

func (s *Service) handle(w goHttp.ResponseWriter, r *goHttp.Request) error {
	startTime := time.Now()
	matcher := common.New(helpers.ExtractHost(r.Host), r.URL.Path, r.Method)

	pickServiceables, err := lookup.Lookup(s, matcher)
	if err != nil {
		return fmt.Errorf("http serviceable lookup failed with: %s", err)
	}

	if len(pickServiceables) != 1 {
		return fmt.Errorf("lookup returned %d serviceables, expected 1", len(pickServiceables))
	}

	pick, ok := pickServiceables[0].(iface.Serviceable)
	if !ok {
		return fmt.Errorf("matched serviceable is not a http serviceable")
	}

	if err := pick.Ready(); err != nil {
		return counter.ErrorWrapper(pick, startTime, time.Time{}, fmt.Errorf("HTTP serviceable is not ready with: %s", err))
	}

	coldStartDoneTime, err := pick.Handle(w, r, matcher)

	return counter.ErrorWrapper(pick, startTime, coldStartDoneTime, err)
}

func (s *Service) attach() {
	s.Http().LowLevel(&http.LowLevelDefinition{
		PathPrefix: "/",
		Handler: func(w goHttp.ResponseWriter, r *goHttp.Request) {
			err := s.handle(w, r)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(500)
			}
		},
	})
}
