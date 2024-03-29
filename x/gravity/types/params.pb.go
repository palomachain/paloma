// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: palomachain/paloma/gravity/params.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/gogoproto/gogoproto"
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

// The slashing fractions for the various gravity related slashing conditions.
// The first three refer to not submitting a particular message, the third for
// submitting a different claim for the same Ethereum event
type Params struct {
	ContractSourceHash           string                      `protobuf:"bytes,1,opt,name=contract_source_hash,json=contractSourceHash,proto3" json:"contract_source_hash,omitempty"`
	BridgeEthereumAddress        string                      `protobuf:"bytes,2,opt,name=bridge_ethereum_address,json=bridgeEthereumAddress,proto3" json:"bridge_ethereum_address,omitempty"`
	BridgeChainId                uint64                      `protobuf:"varint,3,opt,name=bridge_chain_id,json=bridgeChainId,proto3" json:"bridge_chain_id,omitempty"`
	SignedBatchesWindow          uint64                      `protobuf:"varint,4,opt,name=signed_batches_window,json=signedBatchesWindow,proto3" json:"signed_batches_window,omitempty"`
	TargetBatchTimeout           uint64                      `protobuf:"varint,5,opt,name=target_batch_timeout,json=targetBatchTimeout,proto3" json:"target_batch_timeout,omitempty"`
	AverageBlockTime             uint64                      `protobuf:"varint,6,opt,name=average_block_time,json=averageBlockTime,proto3" json:"average_block_time,omitempty"`
	AverageEthereumBlockTime     uint64                      `protobuf:"varint,7,opt,name=average_ethereum_block_time,json=averageEthereumBlockTime,proto3" json:"average_ethereum_block_time,omitempty"`
	SlashFractionBatch           cosmossdk_io_math.LegacyDec `protobuf:"bytes,8,opt,name=slash_fraction_batch,json=slashFractionBatch,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"slash_fraction_batch"`
	SlashFractionBadEthSignature cosmossdk_io_math.LegacyDec `protobuf:"bytes,9,opt,name=slash_fraction_bad_eth_signature,json=slashFractionBadEthSignature,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"slash_fraction_bad_eth_signature"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef38bc2ba159fcc8, []int{0}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetContractSourceHash() string {
	if m != nil {
		return m.ContractSourceHash
	}
	return ""
}

func (m *Params) GetBridgeEthereumAddress() string {
	if m != nil {
		return m.BridgeEthereumAddress
	}
	return ""
}

func (m *Params) GetBridgeChainId() uint64 {
	if m != nil {
		return m.BridgeChainId
	}
	return 0
}

func (m *Params) GetSignedBatchesWindow() uint64 {
	if m != nil {
		return m.SignedBatchesWindow
	}
	return 0
}

func (m *Params) GetTargetBatchTimeout() uint64 {
	if m != nil {
		return m.TargetBatchTimeout
	}
	return 0
}

func (m *Params) GetAverageBlockTime() uint64 {
	if m != nil {
		return m.AverageBlockTime
	}
	return 0
}

func (m *Params) GetAverageEthereumBlockTime() uint64 {
	if m != nil {
		return m.AverageEthereumBlockTime
	}
	return 0
}

func init() {
	proto.RegisterType((*Params)(nil), "palomachain.paloma.gravity.Params")
}

func init() {
	proto.RegisterFile("palomachain/paloma/gravity/params.proto", fileDescriptor_ef38bc2ba159fcc8)
}

var fileDescriptor_ef38bc2ba159fcc8 = []byte{
	// 465 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x4f, 0x8b, 0xd4, 0x30,
	0x18, 0x87, 0x5b, 0x1d, 0x47, 0x37, 0x28, 0x4a, 0x9c, 0xc5, 0x3a, 0x2b, 0xdd, 0xa2, 0xa0, 0x73,
	0xd0, 0x56, 0x14, 0x3c, 0x08, 0x1e, 0x1c, 0x5d, 0xff, 0x80, 0x07, 0x99, 0x55, 0x04, 0x2f, 0xe1,
	0x6d, 0x12, 0x93, 0x30, 0xd3, 0x66, 0x48, 0xd2, 0x5d, 0xe7, 0xe6, 0x47, 0xf0, 0xdb, 0xf8, 0x15,
	0xf6, 0xb8, 0x47, 0x11, 0x59, 0x64, 0xe6, 0x8b, 0x48, 0x93, 0xce, 0xba, 0xa8, 0x87, 0xbd, 0xa5,
	0x79, 0x7e, 0xcf, 0x9b, 0xbc, 0xe9, 0x8b, 0xee, 0xcc, 0x61, 0xa6, 0x2b, 0xa0, 0x12, 0x54, 0x5d,
	0x84, 0x75, 0x21, 0x0c, 0xec, 0x29, 0xb7, 0x28, 0xe6, 0x60, 0xa0, 0xb2, 0xf9, 0xdc, 0x68, 0xa7,
	0xf1, 0xf0, 0x44, 0x30, 0x0f, 0xeb, 0xbc, 0x0b, 0x0e, 0x07, 0x42, 0x0b, 0xed, 0x63, 0x45, 0xbb,
	0x0a, 0xc6, 0xf0, 0x3a, 0xd5, 0xb6, 0xd2, 0x96, 0x04, 0x10, 0x3e, 0x02, 0xba, 0xf9, 0xad, 0x87,
	0xfa, 0x6f, 0x7d, 0x75, 0x7c, 0x1f, 0x0d, 0xa8, 0xae, 0x9d, 0x01, 0xea, 0x88, 0xd5, 0x8d, 0xa1,
	0x9c, 0x48, 0xb0, 0x32, 0x89, 0xb3, 0x78, 0xb4, 0x31, 0xc1, 0x6b, 0xb6, 0xeb, 0xd1, 0x2b, 0xb0,
	0x12, 0x3f, 0x42, 0xd7, 0x4a, 0xa3, 0x98, 0xe0, 0x84, 0x3b, 0xc9, 0x0d, 0x6f, 0x2a, 0x02, 0x8c,
	0x19, 0x6e, 0x6d, 0x72, 0xc6, 0x4b, 0x9b, 0x01, 0xef, 0x74, 0xf4, 0x69, 0x80, 0xf8, 0x36, 0xba,
	0xdc, 0x79, 0xbe, 0x09, 0xa2, 0x58, 0x72, 0x36, 0x8b, 0x47, 0xbd, 0xc9, 0xa5, 0xb0, 0xfd, 0xac,
	0xdd, 0x7d, 0xcd, 0xf0, 0x03, 0xb4, 0x69, 0x95, 0xa8, 0x39, 0x23, 0x25, 0x38, 0x2a, 0xb9, 0x25,
	0xfb, 0xaa, 0x66, 0x7a, 0x3f, 0xe9, 0xf9, 0xf4, 0xd5, 0x00, 0xc7, 0x81, 0x7d, 0xf0, 0xa8, 0xed,
	0xc2, 0x81, 0x11, 0xdc, 0x05, 0x87, 0x38, 0x55, 0x71, 0xdd, 0xb8, 0xe4, 0x9c, 0x57, 0x70, 0x60,
	0x5e, 0x79, 0x17, 0x08, 0xbe, 0x8b, 0x30, 0xec, 0x71, 0x03, 0x82, 0x93, 0x72, 0xa6, 0xe9, 0xd4,
	0x2b, 0x49, 0xdf, 0xe7, 0xaf, 0x74, 0x64, 0xdc, 0x82, 0x56, 0xc0, 0x4f, 0xd0, 0xd6, 0x3a, 0x7d,
	0xdc, 0xf4, 0x09, 0xed, 0xbc, 0xd7, 0x92, 0x2e, 0xb2, 0x6e, 0xfc, 0x8f, 0xfe, 0x1e, 0x0d, 0xec,
	0x0c, 0xac, 0x24, 0x9f, 0xda, 0xb7, 0x54, 0xba, 0x0e, 0xd7, 0x4c, 0x2e, 0x64, 0xf1, 0xe8, 0xe2,
	0xf8, 0xd6, 0xc1, 0xd1, 0x76, 0xf4, 0xe3, 0x68, 0x7b, 0x2b, 0xfc, 0x23, 0xcb, 0xa6, 0xb9, 0xd2,
	0x45, 0x05, 0x4e, 0xe6, 0x6f, 0xb8, 0x00, 0xba, 0x78, 0xce, 0xe9, 0x04, 0xfb, 0x02, 0x2f, 0x3a,
	0xdf, 0xb7, 0x82, 0xa7, 0x28, 0xfb, 0xa7, 0x2c, 0x6b, 0x2f, 0x48, 0xda, 0x37, 0x02, 0xd7, 0x18,
	0x9e, 0x6c, 0x9c, 0xfe, 0x88, 0x1b, 0x7f, 0x1d, 0xc1, 0x76, 0x9c, 0xdc, 0x5d, 0x17, 0x7a, 0xdc,
	0xfb, 0xf2, 0x33, 0x8b, 0xc6, 0x2f, 0x0f, 0x96, 0x69, 0x7c, 0xb8, 0x4c, 0xe3, 0x5f, 0xcb, 0x34,
	0xfe, 0xba, 0x4a, 0xa3, 0xc3, 0x55, 0x1a, 0x7d, 0x5f, 0xa5, 0xd1, 0xc7, 0x7b, 0x42, 0x39, 0xd9,
	0x94, 0x39, 0xd5, 0x55, 0xf1, 0x9f, 0xa1, 0xfe, 0x7c, 0x3c, 0xd6, 0x6e, 0x31, 0xe7, 0xb6, 0xec,
	0xfb, 0x49, 0x7c, 0xf8, 0x3b, 0x00, 0x00, 0xff, 0xff, 0x1f, 0x9f, 0x63, 0x68, 0x01, 0x03, 0x00,
	0x00,
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.SlashFractionBadEthSignature.Size()
		i -= size
		if _, err := m.SlashFractionBadEthSignature.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x4a
	{
		size := m.SlashFractionBatch.Size()
		i -= size
		if _, err := m.SlashFractionBatch.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x42
	if m.AverageEthereumBlockTime != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.AverageEthereumBlockTime))
		i--
		dAtA[i] = 0x38
	}
	if m.AverageBlockTime != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.AverageBlockTime))
		i--
		dAtA[i] = 0x30
	}
	if m.TargetBatchTimeout != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.TargetBatchTimeout))
		i--
		dAtA[i] = 0x28
	}
	if m.SignedBatchesWindow != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.SignedBatchesWindow))
		i--
		dAtA[i] = 0x20
	}
	if m.BridgeChainId != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.BridgeChainId))
		i--
		dAtA[i] = 0x18
	}
	if len(m.BridgeEthereumAddress) > 0 {
		i -= len(m.BridgeEthereumAddress)
		copy(dAtA[i:], m.BridgeEthereumAddress)
		i = encodeVarintParams(dAtA, i, uint64(len(m.BridgeEthereumAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ContractSourceHash) > 0 {
		i -= len(m.ContractSourceHash)
		copy(dAtA[i:], m.ContractSourceHash)
		i = encodeVarintParams(dAtA, i, uint64(len(m.ContractSourceHash)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintParams(dAtA []byte, offset int, v uint64) int {
	offset -= sovParams(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ContractSourceHash)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
	l = len(m.BridgeEthereumAddress)
	if l > 0 {
		n += 1 + l + sovParams(uint64(l))
	}
	if m.BridgeChainId != 0 {
		n += 1 + sovParams(uint64(m.BridgeChainId))
	}
	if m.SignedBatchesWindow != 0 {
		n += 1 + sovParams(uint64(m.SignedBatchesWindow))
	}
	if m.TargetBatchTimeout != 0 {
		n += 1 + sovParams(uint64(m.TargetBatchTimeout))
	}
	if m.AverageBlockTime != 0 {
		n += 1 + sovParams(uint64(m.AverageBlockTime))
	}
	if m.AverageEthereumBlockTime != 0 {
		n += 1 + sovParams(uint64(m.AverageEthereumBlockTime))
	}
	l = m.SlashFractionBatch.Size()
	n += 1 + l + sovParams(uint64(l))
	l = m.SlashFractionBadEthSignature.Size()
	n += 1 + l + sovParams(uint64(l))
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozParams(x uint64) (n int) {
	return sovParams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowParams
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractSourceHash", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractSourceHash = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BridgeEthereumAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BridgeEthereumAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BridgeChainId", wireType)
			}
			m.BridgeChainId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BridgeChainId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SignedBatchesWindow", wireType)
			}
			m.SignedBatchesWindow = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SignedBatchesWindow |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TargetBatchTimeout", wireType)
			}
			m.TargetBatchTimeout = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TargetBatchTimeout |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AverageBlockTime", wireType)
			}
			m.AverageBlockTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AverageBlockTime |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AverageEthereumBlockTime", wireType)
			}
			m.AverageEthereumBlockTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AverageEthereumBlockTime |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SlashFractionBatch", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.SlashFractionBatch.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SlashFractionBadEthSignature", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.SlashFractionBadEthSignature.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipParams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthParams
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
func skipParams(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowParams
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
					return 0, ErrIntOverflowParams
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
					return 0, ErrIntOverflowParams
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
				return 0, ErrInvalidLengthParams
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupParams
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthParams
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthParams        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowParams          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupParams = fmt.Errorf("proto: unexpected end of group")
)
