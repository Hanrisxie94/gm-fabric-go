// Copyright 2017 Decipher Technology Studios LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flatjson

import (
	"bufio"
	"fmt"
	"io"
)

// Writer holds intermediate content for producing JSON
type Writer struct {
	writer    *bufio.Writer
	count     int
	prevKey   string
	prevValue string
}

// New creates a Flat JSON Writer
func New(writer io.Writer) (*Writer, error) {
	w := &Writer{
		writer: bufio.NewWriter(writer),
	}

	if _, err := w.writer.WriteString("{\n"); err != nil {
		return nil, err
	}

	return w, nil
}

// Write stores a value as JSON
func (w *Writer) Write(key string, value interface{}) error {
	if err := w.writePrev(); err != nil {
		return err
	}

	w.prevKey = key

	switch n := value.(type) {
	case string:
		w.prevValue = fmt.Sprintf("\"%s\"", n)
	case int, int32, int64, uint64:
		w.prevValue = fmt.Sprintf("%d", n)
	case float32, float64:
		w.prevValue = fmt.Sprintf("%f", n)
	default:
		return fmt.Errorf("unwriteable type %T; %v", value, value)
	}

	return nil
}

// Flush writes the last line of JSON (if any) and the closing bracket
func (w *Writer) Flush() error {
	var err error

	if w.count > 0 {
		_, err = w.writer.WriteString(fmt.Sprintf("\t\"%s\": %s\n",
			w.prevKey, w.prevValue))
		if err != nil {
			return err
		}
	}

	if _, err = w.writer.WriteString("}\n"); err != nil {
		return err
	}

	return w.writer.Flush()
}

func (w *Writer) writePrev() error {
	if w.count > 0 {
		_, err := w.writer.WriteString(fmt.Sprintf("\t\"%s\": %s,\n",
			w.prevKey, w.prevValue))
		if err != nil {
			return err
		}
	}
	w.count++
	return nil
}
