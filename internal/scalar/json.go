package scalar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type JSON json.RawMessage

func (value *JSON) UnmarshalGQL(input any) error {
	if text, ok := input.(string); ok {
		if !json.Valid([]byte(text)) {
			return fmt.Errorf("invalid JSON")
		}
		*value = JSON([]byte(text))
		return nil
	}

	encoded, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshal JSON scalar: %w", err)
	}
	*value = JSON(encoded)
	return nil
}

func (value JSON) MarshalGQL(writer io.Writer) {
	if len(value) == 0 || bytes.Equal(value, []byte("null")) {
		_, _ = writer.Write([]byte("null"))
		return
	}
	_, _ = writer.Write(value)
}
