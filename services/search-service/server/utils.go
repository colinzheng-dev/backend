package server

import (
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/services/search-service/model"
	"net/url"
)

func GetRegionType(qs url.Values, dst **string) error {

	s := qs.Get("region_type")
	var result string
	if s == "" {
		result := "country"
		*dst = &result
	} else {
		switch s {
		case "country":
			result = "country"
		case "states":
			result = "states"
		//case "all":
		//	result = "all"
		default:
			return errors.New("invalid region parameter")
		}
		*dst = &result
	}
	return nil
}

func (s *Server) LogError(action, msg string, err error, persist bool) {
	log.Error().Err(err).Msg(msg)
	if persist {
		e := model.ErrorLog{
			Action: action,
			Error:   err.Error(),
		}
		if err := s.db.CreateErrorLog(&e); err != nil {
			log.Error().Err(err).Msgf("ERROR occurred while being persisted on database. Action = %s, error = %s \n",e.Action, e.Error)
		}
	}
}