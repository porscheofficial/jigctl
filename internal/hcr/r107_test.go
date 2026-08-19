package hcr

import (
	"reflect"
	"testing"
	"time"
)

func TestApplyR107(t *testing.T) {
	currentDate := time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		exceptions []parsedException
		want       []Diagnostic
	}{
		{
			name:       "expired date is reported",
			exceptions: []parsedException{{Until: "2026-04-11"}},
			want: []Diagnostic{{
				File:    "record.md",
				Pointer: "/exceptions/0/until",
				Code:    "R-107",
				Message: "exception expired on 2026-04-11",
			}},
		},
		{
			name:       "date equal to current date is in force",
			exceptions: []parsedException{{Until: "2026-04-12"}},
		},
		{
			name:       "date after current date is in force",
			exceptions: []parsedException{{Until: "2026-04-13"}},
		},
		{
			name:       "missing date is permanent",
			exceptions: []parsedException{{}},
		},
		{
			name:       "impossible date is reported",
			exceptions: []parsedException{{Until: "2026-13-45"}},
			want: []Diagnostic{{
				File:    "record.md",
				Pointer: "/exceptions/0/until",
				Code:    "R-107",
				Message: "exception until 2026-13-45 is not a calendar date",
			}},
		},
		{
			name: "pointer identifies later exception",
			exceptions: []parsedException{
				{},
				{Until: "2026-04-11"},
			},
			want: []Diagnostic{{
				File:    "record.md",
				Pointer: "/exceptions/1/until",
				Code:    "R-107",
				Message: "exception expired on 2026-04-11",
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			emitters := []emitterRecord{{
				path: "record.md",
				meta: parsedMeta{Exceptions: test.exceptions},
			}}
			var diagnostics []Diagnostic

			// When
			applyR107(currentDate, emitters, &diagnostics)

			// Then
			if !reflect.DeepEqual(diagnostics, test.want) {
				t.Fatalf("diagnostics mismatch\nwant: %+v\ngot:  %+v", test.want, diagnostics)
			}
		})
	}
}

func TestUTCDateOf(t *testing.T) {
	tests := []struct {
		name    string
		instant time.Time
		want    time.Time
	}{
		{
			name:    "last second of a day keeps that day",
			instant: time.Date(2026, time.April, 12, 23, 59, 59, 0, time.UTC),
			want:    time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "already tomorrow east of UTC",
			instant: time.Date(2026, time.April, 13, 8, 0, 0, 0, time.FixedZone("UTC+9", 9*60*60)),
			want:    time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "still yesterday west of UTC",
			instant: time.Date(2026, time.April, 12, 16, 0, 0, 0, time.FixedZone("UTC-10", -10*60*60)),
			want:    time.Date(2026, time.April, 13, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			got := utcDateOf(test.instant)

			// Then
			if !got.Equal(test.want) {
				t.Errorf("utcDateOf(%s) = %s, want %s", test.instant, got, test.want)
			}
		})
	}
}
