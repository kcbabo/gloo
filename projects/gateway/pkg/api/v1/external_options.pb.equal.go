// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/gloo/projects/gateway/api/v1/external_options.proto

package v1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	equality "github.com/solo-io/protoc-gen-ext/pkg/equality"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = bytes.Compare
	_ = strings.Compare
	_ = equality.Equalizer(nil)
	_ = proto.Message(nil)
)

// Equal function
func (m *VirtualHostOption) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*VirtualHostOption)
	if !ok {
		that2, ok := that.(VirtualHostOption)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetMetadata()).(equality.Equalizer); ok {
		if !h.Equal(target.GetMetadata()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetMetadata(), target.GetMetadata()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetOptions()).(equality.Equalizer); ok {
		if !h.Equal(target.GetOptions()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetOptions(), target.GetOptions()) {
			return false
		}
	}

	switch m.StatusOneof.(type) {

	case *VirtualHostOption_Status:
		if _, ok := target.StatusOneof.(*VirtualHostOption_Status); !ok {
			return false
		}

		if h, ok := interface{}(m.GetStatus()).(equality.Equalizer); ok {
			if !h.Equal(target.GetStatus()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetStatus(), target.GetStatus()) {
				return false
			}
		}

	case *VirtualHostOption_ReporterStatus:
		if _, ok := target.StatusOneof.(*VirtualHostOption_ReporterStatus); !ok {
			return false
		}

		if h, ok := interface{}(m.GetReporterStatus()).(equality.Equalizer); ok {
			if !h.Equal(target.GetReporterStatus()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetReporterStatus(), target.GetReporterStatus()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.StatusOneof != target.StatusOneof {
			return false
		}
	}

	return true
}

// Equal function
func (m *RouteOption) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*RouteOption)
	if !ok {
		that2, ok := that.(RouteOption)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetMetadata()).(equality.Equalizer); ok {
		if !h.Equal(target.GetMetadata()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetMetadata(), target.GetMetadata()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetOptions()).(equality.Equalizer); ok {
		if !h.Equal(target.GetOptions()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetOptions(), target.GetOptions()) {
			return false
		}
	}

	switch m.StatusOneof.(type) {

	case *RouteOption_Status:
		if _, ok := target.StatusOneof.(*RouteOption_Status); !ok {
			return false
		}

		if h, ok := interface{}(m.GetStatus()).(equality.Equalizer); ok {
			if !h.Equal(target.GetStatus()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetStatus(), target.GetStatus()) {
				return false
			}
		}

	case *RouteOption_ReporterStatus:
		if _, ok := target.StatusOneof.(*RouteOption_ReporterStatus); !ok {
			return false
		}

		if h, ok := interface{}(m.GetReporterStatus()).(equality.Equalizer); ok {
			if !h.Equal(target.GetReporterStatus()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetReporterStatus(), target.GetReporterStatus()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.StatusOneof != target.StatusOneof {
			return false
		}
	}

	return true
}
