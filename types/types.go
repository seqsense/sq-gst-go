package types

import "unsafe"

type GstElement struct {
	p unsafe.Pointer
}

func NewGstElement(p unsafe.Pointer) *GstElement {
	return &GstElement{p: p}
}

func (s *GstElement) UnsafePointer() unsafe.Pointer {
	return s.p
}
