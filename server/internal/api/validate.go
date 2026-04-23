package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = func() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	return v
}()

type validationDetail struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type validationError struct {
	Code    string             `json:"code"`
	Details []validationDetail `json:"details"`
}

func (e *validationError) Error() string {
	return fmt.Sprintf("validation failed: %d errors", len(e.Details))
}

func decodeAndValidate[T any](w http.ResponseWriter, r *http.Request, target *T) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_json", err)
		return false
	}
	if err := validate.Struct(target); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			details := make([]validationDetail, 0, len(ve))
			for _, fe := range ve {
				details = append(details, validationDetail{
					Field:   fieldName(fe),
					Rule:    fe.Tag(),
					Message: fieldMessage(fe),
				})
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(envelope{
				Error: &apiError{Code: "validation_failed", Message: "请求参数校验失败"},
				Data:  details,
			})
			return false
		}
		respondError(w, http.StatusBadRequest, "validation_failed", err)
		return false
	}
	return true
}

func fieldName(fe validator.FieldError) string {
	field := fe.Field()
	if len(field) == 0 {
		return field
	}
	return strings.ToLower(field[:1]) + field[1:]
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " 不能为空"
	case "max":
		return fe.Field() + " 超过最大长度 " + fe.Param()
	case "min":
		return fe.Field() + " 小于最小值 " + fe.Param()
	case "oneof":
		return fe.Field() + " 必须是以下之一：" + fe.Param()
	case "gte":
		return fe.Field() + " 不能小于 " + fe.Param()
	case "lte":
		return fe.Field() + " 不能大于 " + fe.Param()
	default:
		return fe.Field() + " 校验失败（" + fe.Tag() + "）"
	}
}
