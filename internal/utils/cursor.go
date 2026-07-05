package utils

import (
	"encoding/base64"
	"encoding/json"
)

type Cursor struct {
	CreatedAt int64  `json:"t"`
	ID        string `json:"i"`
}

// EncodeCursor converts timestamp and ID into a Base64 cursor string
func EncodeCursor(createdAt int64, id string) string {
	c := Cursor{
		CreatedAt: createdAt,
		ID:        id,
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// DecodeCursor parses a Base64 cursor string back into timestamp and ID
func DecodeCursor(cursorStr string) (int64, string, error) {
	bytes, err := base64.URLEncoding.DecodeString(cursorStr)
	if err != nil {
		return 0, "", err
	}

	var c Cursor
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return 0, "", err
	}

	return c.CreatedAt, c.ID, nil
}
