package render

import (
	"fmt"
	"strings"
)

// imagePlaceholderState accumulates image references in document order as
// they're encountered during rendering and emits the canonical placeholder
// form "<!-- IMG:N -->" for each occurrence.
//
// This protocol matches what internal/converter.ReplaceImagePlaceholders
// expects downstream: each placeholder maps to an ImageRef index.
type imagePlaceholderState struct {
	srcs []string // URL/path per placeholder index, document order
}

func newImageState() *imagePlaceholderState {
	return &imagePlaceholderState{}
}

// Add records a new image source and returns the HTML placeholder comment.
func (s *imagePlaceholderState) Add(src string) string {
	idx := len(s.srcs)
	s.srcs = append(s.srcs, strings.TrimSpace(src))
	return fmt.Sprintf("<!-- IMG:%d -->", idx)
}

// Sources returns the captured image sources in order. The caller should
// treat the slice as read-only.
func (s *imagePlaceholderState) Sources() []string {
	return s.srcs
}
