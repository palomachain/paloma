// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: evm/add_chain_proposal.proto

package types

import (
	fmt "fmt"
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

type AddChainProposal struct {
	Title             string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description       string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	ChainReferenceID  string `protobuf:"bytes,3,opt,name=chainReferenceID,proto3" json:"chainReferenceID,omitempty"`
	ChainID           uint64 `protobuf:"varint,4,opt,name=chainID,proto3" json:"chainID,omitempty"`
	BlockHeight       uint64 `protobuf:"varint,5,opt,name=blockHeight,proto3" json:"blockHeight,omitempty"`
	BlockHashAtHeight string `protobuf:"bytes,6,opt,name=blockHashAtHeight,proto3" json:"blockHashAtHeight,omitempty"`
}

func (m *AddChainProposal) Reset()         { *m = AddChainProposal{} }
func (m *AddChainProposal) String() string { return proto.CompactTextString(m) }
func (*AddChainProposal) ProtoMessage()    {}
func (*AddChainProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_baa2ec73ba233b34, []int{0}
}
func (m *AddChainProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AddChainProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AddChainProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AddChainProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddChainProposal.Merge(m, src)
}
func (m *AddChainProposal) XXX_Size() int {
	return m.Size()
}
func (m *AddChainProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_AddChainProposal.DiscardUnknown(m)
}

var xxx_messageInfo_AddChainProposal proto.InternalMessageInfo

func (m *AddChainProposal) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *AddChainProposal) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *AddChainProposal) GetChainReferenceID() string {
	if m != nil {
		return m.ChainReferenceID
	}
	return ""
}

func (m *AddChainProposal) GetChainID() uint64 {
	if m != nil {
		return m.ChainID
	}
	return 0
}

func (m *AddChainProposal) GetBlockHeight() uint64 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *AddChainProposal) GetBlockHashAtHeight() string {
	if m != nil {
		return m.BlockHashAtHeight
	}
	return ""
}

func init() {
	proto.RegisterType((*AddChainProposal)(nil), "palomachain.paloma.evm.AddChainProposal")
}

func init() { proto.RegisterFile("evm/add_chain_proposal.proto", fileDescriptor_baa2ec73ba233b34) }

var fileDescriptor_baa2ec73ba233b34 = []byte{
	// 260 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x49, 0x2d, 0xcb, 0xd5,
	0x4f, 0x4c, 0x49, 0x89, 0x4f, 0xce, 0x48, 0xcc, 0xcc, 0x8b, 0x2f, 0x28, 0xca, 0x2f, 0xc8, 0x2f,
	0x4e, 0xcc, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x2b, 0x48, 0xcc, 0xc9, 0xcf, 0x4d,
	0x04, 0xcb, 0xe9, 0x41, 0xd8, 0x7a, 0xa9, 0x65, 0xb9, 0x4a, 0x0f, 0x18, 0xb9, 0x04, 0x1c, 0x53,
	0x52, 0x9c, 0x41, 0xe2, 0x01, 0x50, 0x2d, 0x42, 0x22, 0x5c, 0xac, 0x25, 0x99, 0x25, 0x39, 0xa9,
	0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x10, 0x8e, 0x90, 0x02, 0x17, 0x77, 0x4a, 0x6a, 0x71,
	0x72, 0x51, 0x66, 0x41, 0x49, 0x66, 0x7e, 0x9e, 0x04, 0x13, 0x58, 0x0e, 0x59, 0x48, 0x48, 0x8b,
	0x4b, 0x00, 0x6c, 0x41, 0x50, 0x6a, 0x5a, 0x6a, 0x51, 0x6a, 0x5e, 0x72, 0xaa, 0xa7, 0x8b, 0x04,
	0x33, 0x58, 0x19, 0x86, 0xb8, 0x90, 0x04, 0x17, 0x3b, 0x58, 0xcc, 0xd3, 0x45, 0x82, 0x45, 0x81,
	0x51, 0x83, 0x25, 0x08, 0xc6, 0x05, 0xd9, 0x93, 0x94, 0x93, 0x9f, 0x9c, 0xed, 0x91, 0x9a, 0x99,
	0x9e, 0x51, 0x22, 0xc1, 0x0a, 0x96, 0x45, 0x16, 0x12, 0xd2, 0xe1, 0x12, 0x84, 0x70, 0x13, 0x8b,
	0x33, 0x1c, 0x4b, 0xa0, 0xea, 0xd8, 0xc0, 0x16, 0x61, 0x4a, 0x38, 0x39, 0x9f, 0x78, 0x24, 0xc7,
	0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83, 0x47, 0x72, 0x8c, 0x13, 0x1e, 0xcb, 0x31, 0x5c, 0x78, 0x2c,
	0xc7, 0x70, 0xe3, 0xb1, 0x1c, 0x43, 0x94, 0x66, 0x7a, 0x66, 0x49, 0x46, 0x69, 0x92, 0x5e, 0x72,
	0x7e, 0xae, 0x3e, 0x52, 0xf8, 0x40, 0xd9, 0xfa, 0x15, 0xfa, 0xa0, 0x20, 0x2d, 0xa9, 0x2c, 0x48,
	0x2d, 0x4e, 0x62, 0x03, 0x07, 0xa3, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x4f, 0xaf, 0x6d, 0x2e,
	0x66, 0x01, 0x00, 0x00,
}

func (m *AddChainProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AddChainProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AddChainProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.BlockHashAtHeight) > 0 {
		i -= len(m.BlockHashAtHeight)
		copy(dAtA[i:], m.BlockHashAtHeight)
		i = encodeVarintAddChainProposal(dAtA, i, uint64(len(m.BlockHashAtHeight)))
		i--
		dAtA[i] = 0x32
	}
	if m.BlockHeight != 0 {
		i = encodeVarintAddChainProposal(dAtA, i, uint64(m.BlockHeight))
		i--
		dAtA[i] = 0x28
	}
	if m.ChainID != 0 {
		i = encodeVarintAddChainProposal(dAtA, i, uint64(m.ChainID))
		i--
		dAtA[i] = 0x20
	}
	if len(m.ChainReferenceID) > 0 {
		i -= len(m.ChainReferenceID)
		copy(dAtA[i:], m.ChainReferenceID)
		i = encodeVarintAddChainProposal(dAtA, i, uint64(len(m.ChainReferenceID)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintAddChainProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintAddChainProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintAddChainProposal(dAtA []byte, offset int, v uint64) int {
	offset -= sovAddChainProposal(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *AddChainProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovAddChainProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovAddChainProposal(uint64(l))
	}
	l = len(m.ChainReferenceID)
	if l > 0 {
		n += 1 + l + sovAddChainProposal(uint64(l))
	}
	if m.ChainID != 0 {
		n += 1 + sovAddChainProposal(uint64(m.ChainID))
	}
	if m.BlockHeight != 0 {
		n += 1 + sovAddChainProposal(uint64(m.BlockHeight))
	}
	l = len(m.BlockHashAtHeight)
	if l > 0 {
		n += 1 + l + sovAddChainProposal(uint64(l))
	}
	return n
}

func sovAddChainProposal(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozAddChainProposal(x uint64) (n int) {
	return sovAddChainProposal(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *AddChainProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAddChainProposal
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
			return fmt.Errorf("proto: AddChainProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AddChainProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAddChainProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAddChainProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAddChainProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAddChainProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAddChainProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAddChainProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainReferenceID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAddChainProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAddChainProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAddChainProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChainReferenceID = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainID", wireType)
			}
			m.ChainID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAddChainProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ChainID |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockHeight", wireType)
			}
			m.BlockHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAddChainProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BlockHeight |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockHashAtHeight", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAddChainProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAddChainProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAddChainProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BlockHashAtHeight = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAddChainProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAddChainProposal
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
func skipAddChainProposal(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAddChainProposal
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
					return 0, ErrIntOverflowAddChainProposal
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
					return 0, ErrIntOverflowAddChainProposal
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
				return 0, ErrInvalidLengthAddChainProposal
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupAddChainProposal
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthAddChainProposal
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthAddChainProposal        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAddChainProposal          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupAddChainProposal = fmt.Errorf("proto: unexpected end of group")
)
