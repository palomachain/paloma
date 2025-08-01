// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: palomachain/paloma/skyway/light_node_sale_contracts_proposal.proto

package types

import (
	fmt "fmt"
	proto "github.com/cosmos/gogoproto/proto"
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

type SetLightNodeSaleContractsProposal struct {
	Title                  string                   `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description            string                   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	LightNodeSaleContracts []*LightNodeSaleContract `protobuf:"bytes,3,rep,name=light_node_sale_contracts,json=lightNodeSaleContracts,proto3" json:"light_node_sale_contracts,omitempty"`
}

func (m *SetLightNodeSaleContractsProposal) Reset()         { *m = SetLightNodeSaleContractsProposal{} }
func (m *SetLightNodeSaleContractsProposal) String() string { return proto.CompactTextString(m) }
func (*SetLightNodeSaleContractsProposal) ProtoMessage()    {}
func (*SetLightNodeSaleContractsProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_bcf7977a19185ea8, []int{0}
}
func (m *SetLightNodeSaleContractsProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SetLightNodeSaleContractsProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SetLightNodeSaleContractsProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SetLightNodeSaleContractsProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetLightNodeSaleContractsProposal.Merge(m, src)
}
func (m *SetLightNodeSaleContractsProposal) XXX_Size() int {
	return m.Size()
}
func (m *SetLightNodeSaleContractsProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_SetLightNodeSaleContractsProposal.DiscardUnknown(m)
}

var xxx_messageInfo_SetLightNodeSaleContractsProposal proto.InternalMessageInfo

func (m *SetLightNodeSaleContractsProposal) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *SetLightNodeSaleContractsProposal) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *SetLightNodeSaleContractsProposal) GetLightNodeSaleContracts() []*LightNodeSaleContract {
	if m != nil {
		return m.LightNodeSaleContracts
	}
	return nil
}

func init() {
	proto.RegisterType((*SetLightNodeSaleContractsProposal)(nil), "palomachain.paloma.skyway.SetLightNodeSaleContractsProposal")
}

func init() {
	proto.RegisterFile("palomachain/paloma/skyway/light_node_sale_contracts_proposal.proto", fileDescriptor_bcf7977a19185ea8)
}

var fileDescriptor_bcf7977a19185ea8 = []byte{
	// 256 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x72, 0x2a, 0x48, 0xcc, 0xc9,
	0xcf, 0x4d, 0x4c, 0xce, 0x48, 0xcc, 0xcc, 0xd3, 0x87, 0xb0, 0xf5, 0x8b, 0xb3, 0x2b, 0xcb, 0x13,
	0x2b, 0xf5, 0x73, 0x32, 0xd3, 0x33, 0x4a, 0xe2, 0xf3, 0xf2, 0x53, 0x52, 0xe3, 0x8b, 0x13, 0x73,
	0x52, 0xe3, 0x93, 0xf3, 0xf3, 0x4a, 0x8a, 0x12, 0x93, 0x4b, 0x8a, 0xe3, 0x0b, 0x8a, 0xf2, 0x0b,
	0xf2, 0x8b, 0x13, 0x73, 0xf4, 0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0x24, 0x91, 0xcc, 0xd0, 0x83,
	0xb0, 0xf5, 0x20, 0x66, 0x48, 0x59, 0x90, 0x6e, 0x3c, 0xc4, 0x50, 0xa5, 0x13, 0x8c, 0x5c, 0x8a,
	0xc1, 0xa9, 0x25, 0x3e, 0x20, 0x55, 0x7e, 0xf9, 0x29, 0xa9, 0xc1, 0x89, 0x39, 0xa9, 0xce, 0x30,
	0x17, 0x04, 0x40, 0x1d, 0x20, 0x24, 0xc2, 0xc5, 0x5a, 0x92, 0x59, 0x92, 0x93, 0x2a, 0xc1, 0xa8,
	0xc0, 0xa8, 0xc1, 0x19, 0x04, 0xe1, 0x08, 0x29, 0x70, 0x71, 0xa7, 0xa4, 0x16, 0x27, 0x17, 0x65,
	0x16, 0x94, 0x64, 0xe6, 0xe7, 0x49, 0x30, 0x81, 0xe5, 0x90, 0x85, 0x84, 0xb2, 0xb9, 0x24, 0x71,
	0x7a, 0x4f, 0x82, 0x59, 0x81, 0x59, 0x83, 0xdb, 0xc8, 0x40, 0x0f, 0xa7, 0xb7, 0xf4, 0xb0, 0xba,
	0x2a, 0x48, 0x2c, 0x07, 0xab, 0x63, 0x9d, 0x3c, 0x4f, 0x3c, 0x92, 0x63, 0xbc, 0xf0, 0x48, 0x8e,
	0xf1, 0xc1, 0x23, 0x39, 0xc6, 0x09, 0x8f, 0xe5, 0x18, 0x2e, 0x3c, 0x96, 0x63, 0xb8, 0xf1, 0x58,
	0x8e, 0x21, 0x4a, 0x3f, 0x3d, 0xb3, 0x24, 0xa3, 0x34, 0x49, 0x2f, 0x39, 0x3f, 0x57, 0x1f, 0x4b,
	0x48, 0x95, 0x19, 0xe9, 0x57, 0xc0, 0x82, 0xab, 0xa4, 0xb2, 0x20, 0xb5, 0x38, 0x89, 0x0d, 0x1c,
	0x38, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x84, 0x6e, 0x01, 0x48, 0xb7, 0x01, 0x00, 0x00,
}

func (m *SetLightNodeSaleContractsProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SetLightNodeSaleContractsProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SetLightNodeSaleContractsProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.LightNodeSaleContracts) > 0 {
		for iNdEx := len(m.LightNodeSaleContracts) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.LightNodeSaleContracts[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintLightNodeSaleContractsProposal(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintLightNodeSaleContractsProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintLightNodeSaleContractsProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintLightNodeSaleContractsProposal(dAtA []byte, offset int, v uint64) int {
	offset -= sovLightNodeSaleContractsProposal(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *SetLightNodeSaleContractsProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovLightNodeSaleContractsProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovLightNodeSaleContractsProposal(uint64(l))
	}
	if len(m.LightNodeSaleContracts) > 0 {
		for _, e := range m.LightNodeSaleContracts {
			l = e.Size()
			n += 1 + l + sovLightNodeSaleContractsProposal(uint64(l))
		}
	}
	return n
}

func sovLightNodeSaleContractsProposal(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozLightNodeSaleContractsProposal(x uint64) (n int) {
	return sovLightNodeSaleContractsProposal(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SetLightNodeSaleContractsProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLightNodeSaleContractsProposal
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
			return fmt.Errorf("proto: SetLightNodeSaleContractsProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SetLightNodeSaleContractsProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLightNodeSaleContractsProposal
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
				return ErrInvalidLengthLightNodeSaleContractsProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthLightNodeSaleContractsProposal
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
					return ErrIntOverflowLightNodeSaleContractsProposal
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
				return ErrInvalidLengthLightNodeSaleContractsProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthLightNodeSaleContractsProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LightNodeSaleContracts", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLightNodeSaleContractsProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthLightNodeSaleContractsProposal
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLightNodeSaleContractsProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.LightNodeSaleContracts = append(m.LightNodeSaleContracts, &LightNodeSaleContract{})
			if err := m.LightNodeSaleContracts[len(m.LightNodeSaleContracts)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLightNodeSaleContractsProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLightNodeSaleContractsProposal
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
func skipLightNodeSaleContractsProposal(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLightNodeSaleContractsProposal
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
					return 0, ErrIntOverflowLightNodeSaleContractsProposal
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
					return 0, ErrIntOverflowLightNodeSaleContractsProposal
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
				return 0, ErrInvalidLengthLightNodeSaleContractsProposal
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupLightNodeSaleContractsProposal
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthLightNodeSaleContractsProposal
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthLightNodeSaleContractsProposal        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLightNodeSaleContractsProposal          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupLightNodeSaleContractsProposal = fmt.Errorf("proto: unexpected end of group")
)
