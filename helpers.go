package main

import (
	"fmt"
	"strings"
	"time"
)

func reverseDate(s string) string {
	runes := []rune(s)
	lr := len(runes)
	if lr > 2 {
		runes[0], runes[lr-2] = runes[lr-2], runes[0]
		runes[1], runes[lr-1] = runes[lr-1], runes[1]
	}
	return string(runes)
}

func validDate(date string) error {
	t, err := time.Parse("2006-01-02 15:04", date)
	if err != nil {
		return fmt.Errorf("Неверный формат даты")
	}
	if t.Sub(time.Now()) <= 0 {
		return fmt.Errorf("Заданная дата уже прошла")
	}
	return nil
}

func insRepetitiveVal(str, val string, period int) string {
	lstr := len(str)
	chunks := make([]string, 0, lstr/2+1)
	for i := 0; i < lstr; i += period {
		if i+period > lstr {
			chunks = append(chunks, str[i:])
		} else {
			chunks = append(chunks, str[i:(i+period)])
		}
	}
	return strings.Join(chunks, val)
}
