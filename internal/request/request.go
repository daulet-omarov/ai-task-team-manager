package request

import (
	"encoding/json"
	"net/http"

	"github.com/daulet-omarov/ai-task-team-manager/internal/validator"
)

func DecodeAndValidate(r *http.Request, dst interface{}) error {

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	if err := validator.Validate.Struct(dst); err != nil {
		return err
	}

	return nil
}
