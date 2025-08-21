package tools

import (
	"time"
)

// ExecuteGetCurrentTimeInTokyo mendapatkan waktu saat ini di Tokyo.
// Dibuat menjadi publik (Execute...)
func ExecuteGetCurrentTimeInTokyo(_ string) (string, error) {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", err
	}
	now := time.Now().In(loc)
	return now.Format("Monday, 02 January 2006, 15:04:05 JST"), nil
}

// ExecuteGetTimeIndonesia mendapatkan waktu saat ini di Jakarta.
// Dibuat menjadi publik (Execute...)
func ExecuteGetTimeIndonesia(_ string) (string, error) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return "", err
	}
	now := time.Now().In(loc)
	return now.Format("Monday, 02 January 2006, 15:04:05 WIB"), nil
}