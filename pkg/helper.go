package pkg

import (
	"database/sql"
	"encoding/base64"
	"math/rand"
	"strconv"
)

func NullStringToString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func GetSerialId(n int) string {
	t := "0000000"
	if len(strconv.Itoa(n+1)) == len(strconv.Itoa(n)) {
		return t[len(strconv.Itoa(n)):] + strconv.Itoa(n+1)
	}
	return t[len(strconv.Itoa(n))+1:] + strconv.Itoa(n+1)
}

func GenerateOTP() int {

	return rand.Intn(900000) + 100000
}

func GenerateIdentifier() string {
	b := make([]byte, 8) // 8 bytes = 11 characters base64
	_, err := rand.Read(b)
	if err != nil {
		panic(err) 
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
