// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: palomachain/paloma/paloma/light_node_client.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type LightNodeClient struct {
	ClientAddress string    `protobuf:"bytes,1,opt,name=client_address,json=clientAddress,proto3" json:"client_address,omitempty"`
	ActivatedAt   time.Time `protobuf:"bytes,2,opt,name=activated_at,json=activatedAt,proto3,stdtime" json:"activated_at"`
	LastAuthAt    time.Time `protobuf:"bytes,3,opt,name=last_auth_at,json=lastAuthAt,proto3,stdtime" json:"last_auth_at"`
}

func (m *LightNodeClient) Reset()         { *m = LightNodeClient{} }
func (m *LightNodeClient) String() string { return proto.CompactTextString(m) }
func (*LightNodeClient) ProtoMessage()    {}
func (*LightNodeClient) Descriptor() ([]byte, []int) {
	return fileDescriptor_f112d4742a1730b5, []int{0}
}
func (m *LightNodeClient) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *LightNodeClient) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_LightNodeClient.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *LightNodeClient) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LightNodeClient.Merge(m, src)
}
func (m *LightNodeClient) XXX_Size() int {
	return m.Size()
}
func (m *LightNodeClient) XXX_DiscardUnknown() {
	xxx_messageInfo_LightNodeClient.DiscardUnknown(m)
}

var xxx_messageInfo_LightNodeClient proto.InternalMessageInfo

func (m *LightNodeClient) GetClientAddress() string {
	if m != nil {
		return m.ClientAddress
	}
	return ""
}

func (m *LightNodeClient) GetActivatedAt() time.Time {
	if m != nil {
		return m.ActivatedAt
	}
	return time.Time{}
}

func (m *LightNodeClient) GetLastAuthAt() time.Time {
	if m != nil {
		return m.LastAuthAt
	}
	return time.Time{}
}

func init() {
	proto.RegisterType((*LightNodeClient)(nil), "palomachain.paloma.paloma.LightNodeClient")
}

func init() {
	proto.RegisterFile("palomachain/paloma/paloma/light_node_client.proto", fileDescriptor_f112d4742a1730b5)
}

var fileDescriptor_f112d4742a1730b5 = []byte{
	// 335 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x91, 0xc1, 0x4a, 0xfb, 0x40,
	0x10, 0xc6, 0xb3, 0xff, 0x3f, 0x88, 0xa6, 0x55, 0x21, 0xf4, 0xd0, 0xf6, 0xb0, 0x2d, 0x9e, 0x7a,
	0x31, 0x4b, 0xeb, 0x03, 0x48, 0x2a, 0x28, 0x82, 0x78, 0xa8, 0x9e, 0xbc, 0x84, 0x4d, 0xb2, 0x6e,
	0x16, 0x92, 0x4c, 0xe8, 0x4e, 0x8a, 0xbe, 0x45, 0x1f, 0xc6, 0x87, 0xe8, 0xb1, 0x78, 0xf2, 0xa4,
	0xd2, 0xbc, 0x88, 0x24, 0x9b, 0x88, 0x07, 0x2f, 0x9e, 0xf6, 0x9b, 0x99, 0xef, 0x9b, 0xf9, 0xc1,
	0xda, 0xd3, 0x9c, 0x27, 0x90, 0xf2, 0x30, 0xe6, 0x2a, 0x63, 0x46, 0xb7, 0x4f, 0xa2, 0x64, 0x8c,
	0x7e, 0x06, 0x91, 0xf0, 0xc3, 0x44, 0x89, 0x0c, 0xdd, 0x7c, 0x09, 0x08, 0xce, 0xe0, 0x47, 0xc4,
	0x35, 0xba, 0x79, 0x86, 0x34, 0x04, 0x9d, 0x82, 0x66, 0x01, 0xd7, 0x82, 0xad, 0xa6, 0x81, 0x40,
	0x3e, 0x65, 0x21, 0x54, 0xbe, 0x2a, 0x3a, 0x1c, 0x98, 0xb9, 0x5f, 0x57, 0xcc, 0x14, 0xcd, 0xa8,
	0x27, 0x41, 0x82, 0xe9, 0x57, 0xaa, 0xe9, 0x8e, 0x24, 0x80, 0x4c, 0x04, 0xab, 0xab, 0xa0, 0x78,
	0x64, 0xa8, 0x52, 0xa1, 0x91, 0xa7, 0xb9, 0x31, 0x9c, 0x94, 0xc4, 0x3e, 0xbe, 0xa9, 0x40, 0x6f,
	0x21, 0x12, 0x17, 0x35, 0xa6, 0x73, 0x6e, 0x1f, 0x19, 0x60, 0x9f, 0x47, 0xd1, 0x52, 0x68, 0xdd,
	0x27, 0x63, 0x32, 0x39, 0x98, 0xf7, 0x5f, 0x5f, 0x4e, 0x7b, 0xcd, 0x51, 0xcf, 0x4c, 0xee, 0x70,
	0xa9, 0x32, 0xb9, 0x38, 0x34, 0xfe, 0xa6, 0xe9, 0x5c, 0xd9, 0x5d, 0x1e, 0xa2, 0x5a, 0x71, 0x14,
	0x91, 0xcf, 0xb1, 0xff, 0x6f, 0x4c, 0x26, 0x9d, 0xd9, 0xd0, 0x35, 0x30, 0x6e, 0x0b, 0xe3, 0xde,
	0xb7, 0x30, 0xf3, 0xfd, 0xcd, 0xfb, 0xc8, 0x5a, 0x7f, 0x8c, 0xc8, 0xa2, 0xf3, 0x9d, 0xf4, 0xd0,
	0xb9, 0xb4, 0xbb, 0x09, 0xd7, 0xe8, 0xf3, 0x02, 0xe3, 0x6a, 0xd1, 0xff, 0x3f, 0x2c, 0xb2, 0xab,
	0xa4, 0x57, 0x60, 0xec, 0xe1, 0xfc, 0x7a, 0xb3, 0xa3, 0x64, 0xbb, 0xa3, 0xe4, 0x73, 0x47, 0xc9,
	0xba, 0xa4, 0xd6, 0xb6, 0xa4, 0xd6, 0x5b, 0x49, 0xad, 0x07, 0x26, 0x15, 0xc6, 0x45, 0xe0, 0x86,
	0x90, 0xb2, 0x5f, 0xbe, 0x72, 0x35, 0x63, 0x4f, 0xad, 0xc6, 0xe7, 0x5c, 0xe8, 0x60, 0xaf, 0x3e,
	0x7a, 0xf6, 0x15, 0x00, 0x00, 0xff, 0xff, 0xce, 0xbd, 0xb7, 0xd1, 0xf9, 0x01, 0x00, 0x00,
}

func (m *LightNodeClient) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LightNodeClient) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *LightNodeClient) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.LastAuthAt, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LastAuthAt):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintLightNodeClient(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x1a
	n2, err2 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(m.ActivatedAt, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(m.ActivatedAt):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintLightNodeClient(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x12
	if len(m.ClientAddress) > 0 {
		i -= len(m.ClientAddress)
		copy(dAtA[i:], m.ClientAddress)
		i = encodeVarintLightNodeClient(dAtA, i, uint64(len(m.ClientAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintLightNodeClient(dAtA []byte, offset int, v uint64) int {
	offset -= sovLightNodeClient(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *LightNodeClient) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ClientAddress)
	if l > 0 {
		n += 1 + l + sovLightNodeClient(uint64(l))
	}
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.ActivatedAt)
	n += 1 + l + sovLightNodeClient(uint64(l))
	l = github_com_cosmos_gogoproto_types.SizeOfStdTime(m.LastAuthAt)
	n += 1 + l + sovLightNodeClient(uint64(l))
	return n
}

func sovLightNodeClient(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozLightNodeClient(x uint64) (n int) {
	return sovLightNodeClient(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *LightNodeClient) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLightNodeClient
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
			return fmt.Errorf("proto: LightNodeClient: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LightNodeClient: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClientAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLightNodeClient
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
				return ErrInvalidLengthLightNodeClient
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthLightNodeClient
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ClientAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ActivatedAt", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLightNodeClient
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
				return ErrInvalidLengthLightNodeClient
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLightNodeClient
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.ActivatedAt, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastAuthAt", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLightNodeClient
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
				return ErrInvalidLengthLightNodeClient
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLightNodeClient
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(&m.LastAuthAt, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLightNodeClient(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLightNodeClient
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
func skipLightNodeClient(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLightNodeClient
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
					return 0, ErrIntOverflowLightNodeClient
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
					return 0, ErrIntOverflowLightNodeClient
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
				return 0, ErrInvalidLengthLightNodeClient
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupLightNodeClient
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthLightNodeClient
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthLightNodeClient        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLightNodeClient          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupLightNodeClient = fmt.Errorf("proto: unexpected end of group")
)