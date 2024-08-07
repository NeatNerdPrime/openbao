// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package physical

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrNonUTF8      = errors.New("key contains invalid UTF-8 characters")
	ErrNonPrintable = errors.New("key contains non-printable characters")
)

// StorageEncoding is used to add errors into underlying physical requests
type StorageEncoding struct {
	Backend
}

// Verify StorageEncoding satisfies the correct interfaces
var (
	_ Backend = (*StorageEncoding)(nil)
)

// NewStorageEncoding returns a wrapped physical backend and verifies the key
// encoding
func NewStorageEncoding(b Backend) Backend {
	return &StorageEncoding{
		Backend: b,
	}
}

func (e *StorageEncoding) containsNonPrintableChars(key string) bool {
	idx := strings.IndexFunc(key, func(c rune) bool {
		return !unicode.IsPrint(c)
	})

	return idx != -1
}

func (e *StorageEncoding) Put(ctx context.Context, entry *Entry) error {
	if !utf8.ValidString(entry.Key) {
		return ErrNonUTF8
	}

	if e.containsNonPrintableChars(entry.Key) {
		return ErrNonPrintable
	}

	return e.Backend.Put(ctx, entry)
}

func (e *StorageEncoding) Delete(ctx context.Context, key string) error {
	if !utf8.ValidString(key) {
		return ErrNonUTF8
	}

	if e.containsNonPrintableChars(key) {
		return ErrNonPrintable
	}

	return e.Backend.Delete(ctx, key)
}

func (e *StorageEncoding) Purge(ctx context.Context) {
	if purgeable, ok := e.Backend.(ToggleablePurgemonster); ok {
		purgeable.Purge(ctx)
	}
}

func (e *StorageEncoding) SetEnabled(enabled bool) {
	if purgeable, ok := e.Backend.(ToggleablePurgemonster); ok {
		purgeable.SetEnabled(enabled)
	}
}
