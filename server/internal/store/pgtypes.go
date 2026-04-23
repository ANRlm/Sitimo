package store

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func pgTextFromPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func pgTextFromString(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func pgInt4FromPtr(value *int) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*value), Valid: true}
}

func pgTimestamptzFromTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func pgNumericFromPtr(value *float64) (pgtype.Numeric, error) {
	if value == nil {
		return pgtype.Numeric{}, nil
	}
	return pgNumericFromFloat(*value)
}

func pgNumericFromFloat(value float64) (pgtype.Numeric, error) {
	var numeric pgtype.Numeric
	if err := numeric.Scan(strconv.FormatFloat(value, 'f', -1, 64)); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("scan numeric: %w", err)
	}
	return numeric, nil
}

func pgTextPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func pgTextValue(value pgtype.Text, fallback string) string {
	if !value.Valid || value.String == "" {
		return fallback
	}
	return value.String
}

func pgInt4Ptr(value pgtype.Int4) *int {
	if !value.Valid {
		return nil
	}
	current := int(value.Int32)
	return &current
}

func pgNumericPtr(value pgtype.Numeric) *float64 {
	if !value.Valid {
		return nil
	}
	current, err := value.Float64Value()
	if err != nil || !current.Valid {
		return nil
	}
	number := current.Float64
	return &number
}

func pgNumericValue(value pgtype.Numeric) float64 {
	current, err := value.Float64Value()
	if err != nil || !current.Valid {
		return 0
	}
	return current.Float64
}

func pgTimestamptzPtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	current := value.Time
	return &current
}

func pgTimestamptzValue(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}
