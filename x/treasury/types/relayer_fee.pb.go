// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: palomachain/paloma/treasury/relayer_fee.proto

package types

import (
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	_ "github.com/cosmos/cosmos-proto"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = proto.Marshal
	_ = fmt.Errorf
	_ = math.Inf
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type RelayerFee struct {
	Validator string     `protobuf:"bytes,1,opt,name=validator,proto3" json:"validator,omitempty"`
	Fee       types.Coin `protobuf:"bytes,2,opt,name=fee,proto3" json:"fee"`
}

func (m *RelayerFee) Reset()         { *m = RelayerFee{} }
func (m *RelayerFee) String() string { return proto.CompactTextString(m) }
func (*RelayerFee) ProtoMessage()    {}
func (*RelayerFee) Descriptor() ([]byte, []int) {
	return fileDescriptor_4593a6c78cf76824, []int{0}
}

func (m *RelayerFee) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *RelayerFee) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RelayerFee.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *RelayerFee) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RelayerFee.Merge(m, src)
}

func (m *RelayerFee) XXX_Size() int {
	return m.Size()
}

func (m *RelayerFee) XXX_DiscardUnknown() {
	xxx_messageInfo_RelayerFee.DiscardUnknown(m)
}

var xxx_messageInfo_RelayerFee proto.InternalMessageInfo

func (m *RelayerFee) GetValidator() string {
	if m != nil {
		return m.Validator
	}
	return ""
}

func (m *RelayerFee) GetFee() types.Coin {
	if m != nil {
		return m.Fee
	}
	return types.Coin{}
}

func init() {
	proto.RegisterType((*RelayerFee)(nil), "palomachain.paloma.treasury.RelayerFee")
}

func init() {
	proto.RegisterFile("palomachain/paloma/treasury/relayer_fee.proto", fileDescriptor_4593a6c78cf76824)
}

var fileDescriptor_4593a6c78cf76824 = []byte{
	// 307 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xd2, 0x2d, 0x48, 0xcc, 0xc9,
	0xcf, 0x4d, 0x4c, 0xce, 0x48, 0xcc, 0xcc, 0xd3, 0x87, 0xb0, 0xf5, 0x4b, 0x8a, 0x52, 0x13, 0x8b,
	0x4b, 0x8b, 0x2a, 0xf5, 0x8b, 0x52, 0x73, 0x12, 0x2b, 0x53, 0x8b, 0xe2, 0xd3, 0x52, 0x53, 0xf5,
	0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0xa4, 0x91, 0x94, 0xeb, 0x41, 0xd8, 0x7a, 0x30, 0xe5, 0x52,
	0x92, 0xc9, 0xf9, 0xc5, 0xb9, 0xf9, 0xc5, 0xf1, 0x60, 0xa5, 0xfa, 0x10, 0x0e, 0x44, 0x9f, 0x94,
	0x48, 0x7a, 0x7e, 0x7a, 0x3e, 0x44, 0x1c, 0xc4, 0x82, 0x8a, 0x0a, 0x26, 0xe6, 0x66, 0xe6, 0xe5,
	0xeb, 0x83, 0x49, 0xa8, 0x90, 0x1c, 0x44, 0x9b, 0x7e, 0x52, 0x62, 0x71, 0xaa, 0x7e, 0x99, 0x61,
	0x52, 0x6a, 0x49, 0xa2, 0xa1, 0x7e, 0x72, 0x3e, 0xc8, 0x36, 0x90, 0xbc, 0x52, 0x2f, 0x23, 0x17,
	0x57, 0x10, 0xc4, 0x59, 0x6e, 0xa9, 0xa9, 0x42, 0x66, 0x5c, 0x9c, 0x65, 0x89, 0x39, 0x99, 0x29,
	0x89, 0x25, 0xf9, 0x45, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x4e, 0x12, 0x97, 0xb6, 0xe8, 0x8a,
	0x40, 0x2d, 0x77, 0x4c, 0x49, 0x29, 0x4a, 0x2d, 0x2e, 0x0e, 0x2e, 0x29, 0xca, 0xcc, 0x4b, 0x0f,
	0x42, 0x28, 0x15, 0x72, 0xe7, 0x62, 0x4e, 0x4b, 0x4d, 0x95, 0x60, 0x52, 0x60, 0xd4, 0xe0, 0x36,
	0x92, 0xd4, 0x83, 0x2a, 0x07, 0x59, 0xaa, 0x07, 0xb5, 0x54, 0xcf, 0x39, 0x3f, 0x33, 0xcf, 0x49,
	0xea, 0xc4, 0x3d, 0x79, 0x86, 0x59, 0xcf, 0x37, 0x68, 0xf1, 0xe4, 0xa4, 0xa6, 0x27, 0x26, 0x57,
	0xc6, 0x83, 0x9c, 0x52, 0xbc, 0xe2, 0xf9, 0x06, 0x2d, 0xc6, 0x20, 0x90, 0x09, 0x4e, 0x1e, 0x27,
	0x1e, 0xc9, 0x31, 0x5e, 0x78, 0x24, 0xc7, 0xf8, 0xe0, 0x91, 0x1c, 0xe3, 0x84, 0xc7, 0x72, 0x0c,
	0x17, 0x1e, 0xcb, 0x31, 0xdc, 0x78, 0x2c, 0xc7, 0x10, 0xa5, 0x97, 0x9e, 0x59, 0x92, 0x51, 0x9a,
	0xa4, 0x97, 0x9c, 0x9f, 0xab, 0x8f, 0x25, 0x90, 0x2b, 0x10, 0xc1, 0x5c, 0x52, 0x59, 0x90, 0x5a,
	0x9c, 0xc4, 0x06, 0xf6, 0xa0, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x35, 0x8d, 0x35, 0x8f, 0x92,
	0x01, 0x00, 0x00,
}

func (m *RelayerFee) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RelayerFee) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RelayerFee) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Fee.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintRelayerFee(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Validator) > 0 {
		i -= len(m.Validator)
		copy(dAtA[i:], m.Validator)
		i = encodeVarintRelayerFee(dAtA, i, uint64(len(m.Validator)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintRelayerFee(dAtA []byte, offset int, v uint64) int {
	offset -= sovRelayerFee(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func (m *RelayerFee) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Validator)
	if l > 0 {
		n += 1 + l + sovRelayerFee(uint64(l))
	}
	l = m.Fee.Size()
	n += 1 + l + sovRelayerFee(uint64(l))
	return n
}

func sovRelayerFee(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

func sozRelayerFee(x uint64) (n int) {
	return sovRelayerFee(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}

func (m *RelayerFee) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowRelayerFee
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
			return fmt.Errorf("proto: RelayerFee: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RelayerFee: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validator", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelayerFee
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
				return ErrInvalidLengthRelayerFee
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthRelayerFee
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Validator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fee", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowRelayerFee
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
				return ErrInvalidLengthRelayerFee
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthRelayerFee
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Fee.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipRelayerFee(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthRelayerFee
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

func skipRelayerFee(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowRelayerFee
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
					return 0, ErrIntOverflowRelayerFee
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
					return 0, ErrIntOverflowRelayerFee
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
				return 0, ErrInvalidLengthRelayerFee
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupRelayerFee
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthRelayerFee
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthRelayerFee        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowRelayerFee          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupRelayerFee = fmt.Errorf("proto: unexpected end of group")
)