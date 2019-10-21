package main

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestParseMessage(t *testing.T) {
	day := strconv.Itoa(time.Now().Day())
	testCases := []struct {
		msg    string
		result *message
	}{
		{
			"1420 101299 Сходить в магазин",
			&message{ReminderAt: "2099-12-10 14:20", Txt: "Сходить в магазин"},
		},
		{
			"2359 Сходить в магазин",
			&message{ReminderAt: time.Now().Format("2006-01-02") + " 23:59", Txt: "Сходить в магазин"},
		},
		{
			"2359 " + day + " Сходить в магазин",
			&message{ReminderAt: time.Now().Format("2006-01-02") + " 23:59", Txt: "Сходить в магазин"},
		},
		{
			"2359 3112 Сходить в магазин",
			&message{ReminderAt: time.Now().Format("2006") + "-12-31 23:59", Txt: "Сходить в магазин"},
		},
		{
			"7000 Сходить в магазин",
			nil,
		},
		{
			"5520 Сходить в магазин",
			nil,
		},
		{
			"5520 45 Сходить в магазин",
			nil,
		},
		{
			"5520 101319 Сходить в магазин",
			nil,
		},
		{
			"5520 321119 Сходить в магазин",
			nil,
		},
	}

	for testNumber, testCase := range testCases {
		msg, err := parseMessage(testCase.msg)
		if !cmp.Equal(msg, testCase.result) {
			t.Errorf("#%d Expected %v, got %v\n", testNumber, testCase.result, msg)
		}
		if testCase.result == nil && err == nil {
			t.Errorf("#%d Expected error, got nil\n", testNumber)
		}
	}
}
