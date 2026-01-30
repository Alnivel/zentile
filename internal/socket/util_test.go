package socket

import (
	"bytes"
	"errors"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"
)

func Test_readSplitSeq(t *testing.T) {
	normalData := []string{"Never", "gonna", "give", "you", "up"}

	tests := []struct {
		name string

		input   []string
		sep     []byte
		want    []string
		wantErr error
	}{
		// Normal use cases
		{"NewlineSeparator", normalData, []byte("\n"), normalData, io.EOF},
		{"NullByteSeparator", normalData, []byte{0}, normalData, io.EOF},
		{"MulticharSeparator", normalData, []byte("|separator|"), normalData, io.EOF},
		// Empty inputs
		{"NoDataSent", []string{}, []byte{0}, []string{}, io.EOF},
		{"EmptySegments", []string{"", "", ""}, []byte{0}, []string{"", "", ""}, io.EOF},
		{"EmptySeparator", normalData, []byte{}, []string{}, InvalidSeparatorError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pipeIn, pipeOut := net.Pipe()
			defer pipeOut.Close()

			// Simulate sending data
			go func(pipeIn net.Conn) {
				defer pipeIn.Close()
				buf := new(bytes.Buffer)
				for _, split := range tt.input {
					buf.Write([]byte(split))
					buf.Write(tt.sep)
				}
				pipeIn.Write(buf.Bytes())
			}(pipeIn)

			// Collect results
			got := make([]string, 0)
			var gotErr error = nil

			for split, splitErr := range readSplitSeq(pipeOut, tt.sep) {
				if splitErr != nil {
					gotErr = splitErr
					break
				}

				got = append(got, string(split))
			}

			// Verify
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got result %v, want %v", got, tt.want)
			}
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("got error %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_readSplitSeq_longData(t *testing.T) {
	// The read buffer and the max allowed size of slice are (hardcoded) 512 bytes
	const readBufferLength int = 512
	longStr := strings.Repeat("A", readBufferLength+20)
	// The read buffer is capped at 512 bytes too
	maxLengthData := []string{"One", "Two", longStr[:readBufferLength], "Four"}

	tooLongData := []string{"One", "Two", longStr[:], "Four"}
	// The iterator still should read out strings before the too long one
	tooLongOutput := tooLongData[:2]

	tests := []struct {
		name string

		input   []string
		sep     []byte
		want    []string
		wantErr error
	}{
		// Test behaviour when things get split between reads
		{
			name:    "MaxSplitLength",
			input:   maxLengthData,
			sep:     []byte{0},
			want:    maxLengthData,
			wantErr: io.EOF,
		},
		{
			name:    "SplitTooLong",
			input:   tooLongData,
			sep:     []byte{0},
			want:    tooLongOutput,
			wantErr: SplitTooLongError,
		},
		{
			name:    "SeparatorInNextRead",
			input:   []string{longStr[:readBufferLength], "end"},
			sep:     []byte("|separator|"),
			want:    []string{longStr[:readBufferLength], "end"},
			wantErr: io.EOF,
		},
		{
			name:    "SeparatorSplitBetweenReads",
			input:   []string{longStr[:readBufferLength-10], "end"},
			sep:     []byte("|separator|"),
			want:    []string{longStr[:readBufferLength-10], "end"},
			wantErr: io.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pipeIn, pipeOut := net.Pipe()
			defer pipeOut.Close()

			// Simulate sending data
			go func() {
				defer pipeIn.Close()
				buf := new(bytes.Buffer)
				for _, split := range tt.input {
					buf.Write([]byte(split))
					buf.Write(tt.sep)
				}
				pipeIn.Write(buf.Bytes())
			}()

			// Collect results
			got := make([]string, 0)
			var gotErr error = nil

			for split, splitErr := range readSplitSeq(pipeOut, tt.sep) {
				if splitErr != nil {
					gotErr = splitErr
					break
				}

				got = append(got, string(split))
			}

			// Verify
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got result %v, want %v", got, tt.want)
			}
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("got error %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}
