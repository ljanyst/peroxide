// Copyright (c) 2022 Lukasz Janyst <lukasz@jany.st>
//
// This file is part of Peroxide.
//
// Peroxide is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Peroxide is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Peroxide.  If not, see <https://www.gnu.org/licenses/>.

package message

import (
	"bufio"
	"bytes"
	"io"

	"github.com/pkg/errors"
)

type Base64Sanitizer struct {
	r io.Reader
}

func NewBase64Sanitizer(r io.Reader) (*Base64Sanitizer, error) {
	reader := bufio.NewReader(r)
	var data []byte

	for {
		line, err := reader.ReadBytes('\n')

		if len(line) != 0 {
			line = bytes.TrimSpace(line)
			line = bytes.TrimSuffix(line, []byte("!"))
			data = append(data, line...)
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}

	return &Base64Sanitizer{bytes.NewReader(data)}, nil
}

func (c *Base64Sanitizer) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

type QuotedPrintableSanitizer struct {
	r io.Reader
}

func NewQuotedPrintableSanitizer(r io.Reader) (*QuotedPrintableSanitizer, error) {
	reader := bufio.NewReader(r)
	var data []byte

	for {
		line, err := reader.ReadBytes('\n')

		if len(line) != 0 {
			if len(line) == 1 && line[0] == '=' {
				continue
			}

			data = append(data, line...)
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}

	return &QuotedPrintableSanitizer{bytes.NewReader(data)}, nil
}

func (c *QuotedPrintableSanitizer) Read(b []byte) (int, error) {
	return c.r.Read(b)
}
