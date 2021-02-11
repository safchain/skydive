// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: flow/layers/vrrpv2.proto

package layers

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// LayerVRRPv2 wrapper to generate extra layer
type VRRPv2 struct {
	Contents     []byte `protobuf:"bytes,1,opt,name=contents,proto3" json:"contents,omitempty"`
	Payload      []byte `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	Version      uint8  `protobuf:"varint,3,opt,name=version,proto3,casttype=uint8" json:"version,omitempty"`
	VirtualRtrID uint8  `protobuf:"varint,4,opt,name=virtual_rtr_id,json=virtualRtrId,proto3,casttype=uint8" json:"VirtualRtrID,omitempty"`
	Priority     uint8  `protobuf:"varint,5,opt,name=priority,proto3,casttype=uint8" json:"priority,omitempty"`
	CountIPAddr  uint8  `protobuf:"varint,6,opt,name=count_ipaddr,json=countIpaddr,proto3,casttype=uint8" json:"CountIPAddr,omitempty"`
	AdverInt     uint8  `protobuf:"varint,7,opt,name=adver_int,json=adverInt,proto3,casttype=uint8" json:"adver_int,omitempty"`
	Checksum     uint16 `protobuf:"varint,8,opt,name=checksum,proto3,casttype=uint16" json:"checksum,omitempty"`
}

func (m *VRRPv2) Reset()         { *m = VRRPv2{} }
func (m *VRRPv2) String() string { return proto.CompactTextString(m) }
func (*VRRPv2) ProtoMessage()    {}
func (*VRRPv2) Descriptor() ([]byte, []int) {
	return fileDescriptor_b54c53acd43306bc, []int{0}
}
func (m *VRRPv2) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *VRRPv2) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_VRRPv2.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *VRRPv2) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VRRPv2.Merge(m, src)
}
func (m *VRRPv2) XXX_Size() int {
	return m.ProtoSize()
}
func (m *VRRPv2) XXX_DiscardUnknown() {
	xxx_messageInfo_VRRPv2.DiscardUnknown(m)
}

var xxx_messageInfo_VRRPv2 proto.InternalMessageInfo

func init() {
	proto.RegisterType((*VRRPv2)(nil), "layers.VRRPv2")
}

func init() { proto.RegisterFile("flow/layers/vrrpv2.proto", fileDescriptor_b54c53acd43306bc) }

var fileDescriptor_b54c53acd43306bc = []byte{
	// 388 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x92, 0x4f, 0xcb, 0xd3, 0x30,
	0x1c, 0xc7, 0x5b, 0x9f, 0xe7, 0xe9, 0xba, 0xac, 0x7a, 0x08, 0x2a, 0x61, 0x87, 0x76, 0x28, 0x8e,
	0x1d, 0x74, 0x65, 0x1b, 0x88, 0x78, 0x73, 0x7a, 0x29, 0x78, 0x18, 0x45, 0x06, 0x7a, 0x29, 0x5d,
	0x13, 0xb7, 0xb8, 0xb6, 0x29, 0x69, 0x1a, 0xe9, 0x3b, 0xf0, 0xa8, 0xef, 0x60, 0x2f, 0xc7, 0xe3,
	0x8e, 0x9e, 0x8a, 0xb4, 0xb7, 0xbd, 0x84, 0x9d, 0x64, 0x19, 0xfb, 0x83, 0x7a, 0xcb, 0xf7, 0xf7,
	0xf9, 0xf0, 0xfd, 0x05, 0x12, 0x80, 0x3e, 0xc7, 0xec, 0xab, 0x1b, 0x87, 0x25, 0xe1, 0xb9, 0x2b,
	0x39, 0xcf, 0xe4, 0x78, 0x98, 0x71, 0x26, 0x18, 0x34, 0x8e, 0xc3, 0xee, 0xc3, 0x25, 0x5b, 0x32,
	0x35, 0x72, 0x0f, 0xa7, 0x23, 0x7d, 0xf2, 0xe3, 0x06, 0x18, 0x73, 0xdf, 0x9f, 0xc9, 0x31, 0xec,
	0x02, 0x33, 0x62, 0xa9, 0x20, 0xa9, 0xc8, 0x91, 0xde, 0xd3, 0x07, 0x96, 0x7f, 0xce, 0x10, 0x81,
	0x56, 0x16, 0x96, 0x31, 0x0b, 0x31, 0xba, 0xa7, 0xd0, 0x29, 0xc2, 0xa7, 0xa0, 0x25, 0x09, 0xcf,
	0x29, 0x4b, 0xd1, 0x4d, 0x4f, 0x1f, 0xdc, 0x9f, 0xb6, 0xf7, 0x95, 0x73, 0x57, 0xd0, 0x54, 0xbc,
	0xf2, 0x4f, 0x04, 0x7e, 0x04, 0x0f, 0x24, 0xe5, 0xa2, 0x08, 0xe3, 0x80, 0x0b, 0x1e, 0x50, 0x8c,
	0x6e, 0x95, 0x3b, 0xa9, 0x2b, 0xc7, 0x9a, 0x1f, 0x89, 0x2f, 0xb8, 0xf7, 0x6e, 0x57, 0x39, 0x8f,
	0xaf, 0xf3, 0x73, 0x96, 0x50, 0x41, 0x92, 0x4c, 0x94, 0x97, 0x56, 0x4b, 0x5e, 0x04, 0x0c, 0x9f,
	0x01, 0x33, 0xe3, 0x94, 0x71, 0x2a, 0x4a, 0x74, 0xf7, 0xf7, 0x05, 0xce, 0x08, 0x7e, 0x00, 0x56,
	0xc4, 0x8a, 0x54, 0x04, 0x34, 0x0b, 0x31, 0xe6, 0xc8, 0x50, 0xea, 0xa8, 0xae, 0x9c, 0xce, 0xdb,
	0xc3, 0xdc, 0x9b, 0xbd, 0xc1, 0x98, 0xef, 0x2a, 0xe7, 0xd1, 0x55, 0xfc, 0xdf, 0xf6, 0x8e, 0xaa,
	0xf1, 0x54, 0x0b, 0xec, 0x83, 0x76, 0x88, 0x25, 0xe1, 0x01, 0x4d, 0x05, 0x6a, 0xfd, 0xb3, 0x5d,
	0x31, 0x2f, 0x15, 0xb0, 0x0f, 0xcc, 0x68, 0x45, 0xa2, 0x75, 0x5e, 0x24, 0xc8, 0x54, 0x1a, 0xd8,
	0x57, 0x8e, 0x71, 0xd0, 0x46, 0x2f, 0xfd, 0x33, 0x7b, 0x7d, 0xfb, 0x6d, 0xe3, 0x68, 0xd3, 0xf7,
	0x3f, 0x6b, 0x5b, 0xdf, 0xd6, 0xb6, 0xfe, 0xbb, 0xb6, 0xb5, 0xef, 0x8d, 0xad, 0x6d, 0x1a, 0x5b,
	0xdf, 0x36, 0xb6, 0xf6, 0xab, 0xb1, 0xb5, 0x4f, 0xc3, 0x25, 0x15, 0xab, 0x62, 0x31, 0x8c, 0x58,
	0xe2, 0xe6, 0xeb, 0x12, 0x53, 0x49, 0x5e, 0x64, 0x9c, 0x7d, 0x21, 0x91, 0x38, 0x65, 0xf7, 0xea,
	0x33, 0x2c, 0x0c, 0xf5, 0xd0, 0x93, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x13, 0xbe, 0xc0, 0xe7,
	0x22, 0x02, 0x00, 0x00,
}

func (m *VRRPv2) Marshal() (dAtA []byte, err error) {
	size := m.ProtoSize()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VRRPv2) MarshalTo(dAtA []byte) (int, error) {
	size := m.ProtoSize()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *VRRPv2) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Checksum != 0 {
		i = encodeVarintVrrpv2(dAtA, i, uint64(m.Checksum))
		i--
		dAtA[i] = 0x40
	}
	if m.AdverInt != 0 {
		i = encodeVarintVrrpv2(dAtA, i, uint64(m.AdverInt))
		i--
		dAtA[i] = 0x38
	}
	if m.CountIPAddr != 0 {
		i = encodeVarintVrrpv2(dAtA, i, uint64(m.CountIPAddr))
		i--
		dAtA[i] = 0x30
	}
	if m.Priority != 0 {
		i = encodeVarintVrrpv2(dAtA, i, uint64(m.Priority))
		i--
		dAtA[i] = 0x28
	}
	if m.VirtualRtrID != 0 {
		i = encodeVarintVrrpv2(dAtA, i, uint64(m.VirtualRtrID))
		i--
		dAtA[i] = 0x20
	}
	if m.Version != 0 {
		i = encodeVarintVrrpv2(dAtA, i, uint64(m.Version))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Payload) > 0 {
		i -= len(m.Payload)
		copy(dAtA[i:], m.Payload)
		i = encodeVarintVrrpv2(dAtA, i, uint64(len(m.Payload)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Contents) > 0 {
		i -= len(m.Contents)
		copy(dAtA[i:], m.Contents)
		i = encodeVarintVrrpv2(dAtA, i, uint64(len(m.Contents)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintVrrpv2(dAtA []byte, offset int, v uint64) int {
	offset -= sovVrrpv2(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *VRRPv2) ProtoSize() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Contents)
	if l > 0 {
		n += 1 + l + sovVrrpv2(uint64(l))
	}
	l = len(m.Payload)
	if l > 0 {
		n += 1 + l + sovVrrpv2(uint64(l))
	}
	if m.Version != 0 {
		n += 1 + sovVrrpv2(uint64(m.Version))
	}
	if m.VirtualRtrID != 0 {
		n += 1 + sovVrrpv2(uint64(m.VirtualRtrID))
	}
	if m.Priority != 0 {
		n += 1 + sovVrrpv2(uint64(m.Priority))
	}
	if m.CountIPAddr != 0 {
		n += 1 + sovVrrpv2(uint64(m.CountIPAddr))
	}
	if m.AdverInt != 0 {
		n += 1 + sovVrrpv2(uint64(m.AdverInt))
	}
	if m.Checksum != 0 {
		n += 1 + sovVrrpv2(uint64(m.Checksum))
	}
	return n
}

func sovVrrpv2(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozVrrpv2(x uint64) (n int) {
	return sovVrrpv2(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *VRRPv2) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowVrrpv2
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: VRRPv2: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VRRPv2: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Contents", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVrrpv2
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthVrrpv2
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthVrrpv2
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Contents = append(m.Contents[:0], dAtA[iNdEx:postIndex]...)
			if m.Contents == nil {
				m.Contents = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Payload", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVrrpv2
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthVrrpv2
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthVrrpv2
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Payload = append(m.Payload[:0], dAtA[iNdEx:postIndex]...)
			if m.Payload == nil {
				m.Payload = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
			}
			m.Version = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVrrpv2
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Version |= uint8(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field VirtualRtrID", wireType)
			}
			m.VirtualRtrID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVrrpv2
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.VirtualRtrID |= uint8(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Priority", wireType)
			}
			m.Priority = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVrrpv2
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Priority |= uint8(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CountIPAddr", wireType)
			}
			m.CountIPAddr = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVrrpv2
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CountIPAddr |= uint8(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AdverInt", wireType)
			}
			m.AdverInt = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVrrpv2
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AdverInt |= uint8(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Checksum", wireType)
			}
			m.Checksum = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVrrpv2
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Checksum |= uint16(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipVrrpv2(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthVrrpv2
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipVrrpv2(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowVrrpv2
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowVrrpv2
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowVrrpv2
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthVrrpv2
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupVrrpv2
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthVrrpv2
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthVrrpv2        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowVrrpv2          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupVrrpv2 = fmt.Errorf("proto: unexpected end of group")
)
