// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.12.4
// source: chainrpc/chainnotifier.proto

package chainrpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ConfRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The transaction hash for which we should request a confirmation notification
	// for. If set to a hash of all zeros, then the confirmation notification will
	// be requested for the script instead.
	Txid []byte `protobuf:"bytes,1,opt,name=txid,proto3" json:"txid,omitempty"`
	// An output script within a transaction with the hash above which will be used
	// by light clients to match block filters. If the transaction hash is set to a
	// hash of all zeros, then a confirmation notification will be requested for
	// this script instead.
	Script []byte `protobuf:"bytes,2,opt,name=script,proto3" json:"script,omitempty"`
	// The number of desired confirmations the transaction/output script should
	// reach before dispatching a confirmation notification.
	NumConfs uint32 `protobuf:"varint,3,opt,name=num_confs,json=numConfs,proto3" json:"num_confs,omitempty"`
	// The earliest height in the chain for which the transaction/output script
	// could have been included in a block. This should in most cases be set to the
	// broadcast height of the transaction/output script.
	HeightHint uint32 `protobuf:"varint,4,opt,name=height_hint,json=heightHint,proto3" json:"height_hint,omitempty"`
}

func (x *ConfRequest) Reset() {
	*x = ConfRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_chainnotifier_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfRequest) ProtoMessage() {}

func (x *ConfRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_chainnotifier_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfRequest.ProtoReflect.Descriptor instead.
func (*ConfRequest) Descriptor() ([]byte, []int) {
	return file_chainrpc_chainnotifier_proto_rawDescGZIP(), []int{0}
}

func (x *ConfRequest) GetTxid() []byte {
	if x != nil {
		return x.Txid
	}
	return nil
}

func (x *ConfRequest) GetScript() []byte {
	if x != nil {
		return x.Script
	}
	return nil
}

func (x *ConfRequest) GetNumConfs() uint32 {
	if x != nil {
		return x.NumConfs
	}
	return 0
}

func (x *ConfRequest) GetHeightHint() uint32 {
	if x != nil {
		return x.HeightHint
	}
	return 0
}

type ConfDetails struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The raw bytes of the confirmed transaction.
	RawTx []byte `protobuf:"bytes,1,opt,name=raw_tx,json=rawTx,proto3" json:"raw_tx,omitempty"`
	// The hash of the block in which the confirmed transaction was included in.
	BlockHash []byte `protobuf:"bytes,2,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
	// The height of the block in which the confirmed transaction was included
	// in.
	BlockHeight uint32 `protobuf:"varint,3,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	// The index of the confirmed transaction within the transaction.
	TxIndex uint32 `protobuf:"varint,4,opt,name=tx_index,json=txIndex,proto3" json:"tx_index,omitempty"`
}

func (x *ConfDetails) Reset() {
	*x = ConfDetails{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_chainnotifier_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfDetails) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfDetails) ProtoMessage() {}

func (x *ConfDetails) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_chainnotifier_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfDetails.ProtoReflect.Descriptor instead.
func (*ConfDetails) Descriptor() ([]byte, []int) {
	return file_chainrpc_chainnotifier_proto_rawDescGZIP(), []int{1}
}

func (x *ConfDetails) GetRawTx() []byte {
	if x != nil {
		return x.RawTx
	}
	return nil
}

func (x *ConfDetails) GetBlockHash() []byte {
	if x != nil {
		return x.BlockHash
	}
	return nil
}

func (x *ConfDetails) GetBlockHeight() uint32 {
	if x != nil {
		return x.BlockHeight
	}
	return 0
}

func (x *ConfDetails) GetTxIndex() uint32 {
	if x != nil {
		return x.TxIndex
	}
	return 0
}

type Reorg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Reorg) Reset() {
	*x = Reorg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_chainnotifier_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Reorg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Reorg) ProtoMessage() {}

func (x *Reorg) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_chainnotifier_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Reorg.ProtoReflect.Descriptor instead.
func (*Reorg) Descriptor() ([]byte, []int) {
	return file_chainrpc_chainnotifier_proto_rawDescGZIP(), []int{2}
}

type ConfEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Event:
	//
	//	*ConfEvent_Conf
	//	*ConfEvent_Reorg
	Event isConfEvent_Event `protobuf_oneof:"event"`
}

func (x *ConfEvent) Reset() {
	*x = ConfEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_chainnotifier_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConfEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfEvent) ProtoMessage() {}

func (x *ConfEvent) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_chainnotifier_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConfEvent.ProtoReflect.Descriptor instead.
func (*ConfEvent) Descriptor() ([]byte, []int) {
	return file_chainrpc_chainnotifier_proto_rawDescGZIP(), []int{3}
}

func (m *ConfEvent) GetEvent() isConfEvent_Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (x *ConfEvent) GetConf() *ConfDetails {
	if x, ok := x.GetEvent().(*ConfEvent_Conf); ok {
		return x.Conf
	}
	return nil
}

func (x *ConfEvent) GetReorg() *Reorg {
	if x, ok := x.GetEvent().(*ConfEvent_Reorg); ok {
		return x.Reorg
	}
	return nil
}

type isConfEvent_Event interface {
	isConfEvent_Event()
}

type ConfEvent_Conf struct {
	// An event that includes the confirmation details of the request
	// (txid/ouput script).
	Conf *ConfDetails `protobuf:"bytes,1,opt,name=conf,proto3,oneof"`
}

type ConfEvent_Reorg struct {
	// An event send when the transaction of the request is reorged out of the
	// chain.
	Reorg *Reorg `protobuf:"bytes,2,opt,name=reorg,proto3,oneof"`
}

func (*ConfEvent_Conf) isConfEvent_Event() {}

func (*ConfEvent_Reorg) isConfEvent_Event() {}

type Outpoint struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The hash of the transaction.
	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	// The index of the output within the transaction.
	Index uint32 `protobuf:"varint,2,opt,name=index,proto3" json:"index,omitempty"`
}

func (x *Outpoint) Reset() {
	*x = Outpoint{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_chainnotifier_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Outpoint) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Outpoint) ProtoMessage() {}

func (x *Outpoint) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_chainnotifier_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Outpoint.ProtoReflect.Descriptor instead.
func (*Outpoint) Descriptor() ([]byte, []int) {
	return file_chainrpc_chainnotifier_proto_rawDescGZIP(), []int{4}
}

func (x *Outpoint) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

func (x *Outpoint) GetIndex() uint32 {
	if x != nil {
		return x.Index
	}
	return 0
}

type SpendRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The outpoint for which we should request a spend notification for. If set to
	// a zero outpoint, then the spend notification will be requested for the
	// script instead. A zero or nil outpoint is not supported for Taproot spends
	// because the output script cannot reliably be computed from the witness alone
	// and the spent output script is not always available in the rescan context.
	// So an outpoint must _always_ be specified when registering a spend
	// notification for a Taproot output.
	Outpoint *Outpoint `protobuf:"bytes,1,opt,name=outpoint,proto3" json:"outpoint,omitempty"`
	// The output script for the outpoint above. This will be used by light clients
	// to match block filters. If the outpoint is set to a zero outpoint, then a
	// spend notification will be requested for this script instead.
	Script []byte `protobuf:"bytes,2,opt,name=script,proto3" json:"script,omitempty"`
	// The earliest height in the chain for which the outpoint/output script could
	// have been spent. This should in most cases be set to the broadcast height of
	// the outpoint/output script.
	HeightHint uint32 `protobuf:"varint,3,opt,name=height_hint,json=heightHint,proto3" json:"height_hint,omitempty"`
}

func (x *SpendRequest) Reset() {
	*x = SpendRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_chainnotifier_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SpendRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SpendRequest) ProtoMessage() {}

func (x *SpendRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_chainnotifier_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SpendRequest.ProtoReflect.Descriptor instead.
func (*SpendRequest) Descriptor() ([]byte, []int) {
	return file_chainrpc_chainnotifier_proto_rawDescGZIP(), []int{5}
}

func (x *SpendRequest) GetOutpoint() *Outpoint {
	if x != nil {
		return x.Outpoint
	}
	return nil
}

func (x *SpendRequest) GetScript() []byte {
	if x != nil {
		return x.Script
	}
	return nil
}

func (x *SpendRequest) GetHeightHint() uint32 {
	if x != nil {
		return x.HeightHint
	}
	return 0
}

type SpendDetails struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The outpoint was that spent.
	SpendingOutpoint *Outpoint `protobuf:"bytes,1,opt,name=spending_outpoint,json=spendingOutpoint,proto3" json:"spending_outpoint,omitempty"`
	// The raw bytes of the spending transaction.
	RawSpendingTx []byte `protobuf:"bytes,2,opt,name=raw_spending_tx,json=rawSpendingTx,proto3" json:"raw_spending_tx,omitempty"`
	// The hash of the spending transaction.
	SpendingTxHash []byte `protobuf:"bytes,3,opt,name=spending_tx_hash,json=spendingTxHash,proto3" json:"spending_tx_hash,omitempty"`
	// The input of the spending transaction that fulfilled the spend request.
	SpendingInputIndex uint32 `protobuf:"varint,4,opt,name=spending_input_index,json=spendingInputIndex,proto3" json:"spending_input_index,omitempty"`
	// The height at which the spending transaction was included in a block.
	SpendingHeight uint32 `protobuf:"varint,5,opt,name=spending_height,json=spendingHeight,proto3" json:"spending_height,omitempty"`
}

func (x *SpendDetails) Reset() {
	*x = SpendDetails{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_chainnotifier_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SpendDetails) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SpendDetails) ProtoMessage() {}

func (x *SpendDetails) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_chainnotifier_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SpendDetails.ProtoReflect.Descriptor instead.
func (*SpendDetails) Descriptor() ([]byte, []int) {
	return file_chainrpc_chainnotifier_proto_rawDescGZIP(), []int{6}
}

func (x *SpendDetails) GetSpendingOutpoint() *Outpoint {
	if x != nil {
		return x.SpendingOutpoint
	}
	return nil
}

func (x *SpendDetails) GetRawSpendingTx() []byte {
	if x != nil {
		return x.RawSpendingTx
	}
	return nil
}

func (x *SpendDetails) GetSpendingTxHash() []byte {
	if x != nil {
		return x.SpendingTxHash
	}
	return nil
}

func (x *SpendDetails) GetSpendingInputIndex() uint32 {
	if x != nil {
		return x.SpendingInputIndex
	}
	return 0
}

func (x *SpendDetails) GetSpendingHeight() uint32 {
	if x != nil {
		return x.SpendingHeight
	}
	return 0
}

type SpendEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Event:
	//
	//	*SpendEvent_Spend
	//	*SpendEvent_Reorg
	Event isSpendEvent_Event `protobuf_oneof:"event"`
}

func (x *SpendEvent) Reset() {
	*x = SpendEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_chainnotifier_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SpendEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SpendEvent) ProtoMessage() {}

func (x *SpendEvent) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_chainnotifier_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SpendEvent.ProtoReflect.Descriptor instead.
func (*SpendEvent) Descriptor() ([]byte, []int) {
	return file_chainrpc_chainnotifier_proto_rawDescGZIP(), []int{7}
}

func (m *SpendEvent) GetEvent() isSpendEvent_Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (x *SpendEvent) GetSpend() *SpendDetails {
	if x, ok := x.GetEvent().(*SpendEvent_Spend); ok {
		return x.Spend
	}
	return nil
}

func (x *SpendEvent) GetReorg() *Reorg {
	if x, ok := x.GetEvent().(*SpendEvent_Reorg); ok {
		return x.Reorg
	}
	return nil
}

type isSpendEvent_Event interface {
	isSpendEvent_Event()
}

type SpendEvent_Spend struct {
	// An event that includes the details of the spending transaction of the
	// request (outpoint/output script).
	Spend *SpendDetails `protobuf:"bytes,1,opt,name=spend,proto3,oneof"`
}

type SpendEvent_Reorg struct {
	// An event sent when the spending transaction of the request was
	// reorged out of the chain.
	Reorg *Reorg `protobuf:"bytes,2,opt,name=reorg,proto3,oneof"`
}

func (*SpendEvent_Spend) isSpendEvent_Event() {}

func (*SpendEvent_Reorg) isSpendEvent_Event() {}

type BlockEpoch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The hash of the block.
	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	// The height of the block.
	Height uint32 `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
}

func (x *BlockEpoch) Reset() {
	*x = BlockEpoch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chainrpc_chainnotifier_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockEpoch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockEpoch) ProtoMessage() {}

func (x *BlockEpoch) ProtoReflect() protoreflect.Message {
	mi := &file_chainrpc_chainnotifier_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockEpoch.ProtoReflect.Descriptor instead.
func (*BlockEpoch) Descriptor() ([]byte, []int) {
	return file_chainrpc_chainnotifier_proto_rawDescGZIP(), []int{8}
}

func (x *BlockEpoch) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

func (x *BlockEpoch) GetHeight() uint32 {
	if x != nil {
		return x.Height
	}
	return 0
}

var File_chainrpc_chainnotifier_proto protoreflect.FileDescriptor

var file_chainrpc_chainnotifier_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08,
	0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x22, 0x77, 0x0a, 0x0b, 0x43, 0x6f, 0x6e, 0x66,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x78, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x74, 0x78, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x6e, 0x75, 0x6d, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x73,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x6e, 0x75, 0x6d, 0x43, 0x6f, 0x6e, 0x66, 0x73,
	0x12, 0x1f, 0x0a, 0x0b, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x5f, 0x68, 0x69, 0x6e, 0x74, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x48, 0x69, 0x6e,
	0x74, 0x22, 0x81, 0x01, 0x0a, 0x0b, 0x43, 0x6f, 0x6e, 0x66, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c,
	0x73, 0x12, 0x15, 0x0a, 0x06, 0x72, 0x61, 0x77, 0x5f, 0x74, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x05, 0x72, 0x61, 0x77, 0x54, 0x78, 0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x74, 0x78,
	0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x74, 0x78,
	0x49, 0x6e, 0x64, 0x65, 0x78, 0x22, 0x07, 0x0a, 0x05, 0x52, 0x65, 0x6f, 0x72, 0x67, 0x22, 0x6a,
	0x0a, 0x09, 0x43, 0x6f, 0x6e, 0x66, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x2b, 0x0a, 0x04, 0x63,
	0x6f, 0x6e, 0x66, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x68, 0x61, 0x69,
	0x6e, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73,
	0x48, 0x00, 0x52, 0x04, 0x63, 0x6f, 0x6e, 0x66, 0x12, 0x27, 0x0a, 0x05, 0x72, 0x65, 0x6f, 0x72,
	0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72,
	0x70, 0x63, 0x2e, 0x52, 0x65, 0x6f, 0x72, 0x67, 0x48, 0x00, 0x52, 0x05, 0x72, 0x65, 0x6f, 0x72,
	0x67, 0x42, 0x07, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x34, 0x0a, 0x08, 0x4f, 0x75,
	0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6e,
	0x64, 0x65, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78,
	0x22, 0x77, 0x0a, 0x0c, 0x53, 0x70, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x2e, 0x0a, 0x08, 0x6f, 0x75, 0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2e, 0x4f, 0x75,
	0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x08, 0x6f, 0x75, 0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x06, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x68, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x5f, 0x68, 0x69, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x68,
	0x65, 0x69, 0x67, 0x68, 0x74, 0x48, 0x69, 0x6e, 0x74, 0x22, 0xfc, 0x01, 0x0a, 0x0c, 0x53, 0x70,
	0x65, 0x6e, 0x64, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x12, 0x3f, 0x0a, 0x11, 0x73, 0x70,
	0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x5f, 0x6f, 0x75, 0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63,
	0x2e, 0x4f, 0x75, 0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x10, 0x73, 0x70, 0x65, 0x6e, 0x64,
	0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x26, 0x0a, 0x0f, 0x72,
	0x61, 0x77, 0x5f, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x78, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x72, 0x61, 0x77, 0x53, 0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e,
	0x67, 0x54, 0x78, 0x12, 0x28, 0x0a, 0x10, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x5f,
	0x74, 0x78, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x73,
	0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x54, 0x78, 0x48, 0x61, 0x73, 0x68, 0x12, 0x30, 0x0a,
	0x14, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x5f,
	0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x12, 0x73, 0x70, 0x65,
	0x6e, 0x64, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12,
	0x27, 0x0a, 0x0f, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x5f, 0x68, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0e, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x69,
	0x6e, 0x67, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x22, 0x6e, 0x0a, 0x0a, 0x53, 0x70, 0x65, 0x6e,
	0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x2e, 0x0a, 0x05, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63,
	0x2e, 0x53, 0x70, 0x65, 0x6e, 0x64, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x48, 0x00, 0x52,
	0x05, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x12, 0x27, 0x0a, 0x05, 0x72, 0x65, 0x6f, 0x72, 0x67, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63,
	0x2e, 0x52, 0x65, 0x6f, 0x72, 0x67, 0x48, 0x00, 0x52, 0x05, 0x72, 0x65, 0x6f, 0x72, 0x67, 0x42,
	0x07, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x38, 0x0a, 0x0a, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x68, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x32, 0xe7, 0x01, 0x0a, 0x0d, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x4e, 0x6f, 0x74, 0x69,
	0x66, 0x69, 0x65, 0x72, 0x12, 0x49, 0x0a, 0x19, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x4e, 0x74, 0x66,
	0x6e, 0x12, 0x15, 0x2e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f, 0x6e,
	0x66, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x30, 0x01, 0x12,
	0x43, 0x0a, 0x11, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x53, 0x70, 0x65, 0x6e, 0x64,
	0x4e, 0x74, 0x66, 0x6e, 0x12, 0x16, 0x2e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2e,
	0x53, 0x70, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x63,
	0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x70, 0x65, 0x6e, 0x64, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x30, 0x01, 0x12, 0x46, 0x0a, 0x16, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x4e, 0x74, 0x66, 0x6e, 0x12, 0x14,
	0x2e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x45,
	0x70, 0x6f, 0x63, 0x68, 0x1a, 0x14, 0x2e, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x2e,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x30, 0x01, 0x42, 0x30, 0x5a, 0x2e,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x69, 0x6e, 0x64, 0x72, 0x61,
	0x2d, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x69, 0x6e, 0x64, 0x72, 0x61, 0x2f, 0x70, 0x6b, 0x67, 0x2f,
	0x6c, 0x6e, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x72, 0x70, 0x63, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chainrpc_chainnotifier_proto_rawDescOnce sync.Once
	file_chainrpc_chainnotifier_proto_rawDescData = file_chainrpc_chainnotifier_proto_rawDesc
)

func file_chainrpc_chainnotifier_proto_rawDescGZIP() []byte {
	file_chainrpc_chainnotifier_proto_rawDescOnce.Do(func() {
		file_chainrpc_chainnotifier_proto_rawDescData = protoimpl.X.CompressGZIP(file_chainrpc_chainnotifier_proto_rawDescData)
	})
	return file_chainrpc_chainnotifier_proto_rawDescData
}

var file_chainrpc_chainnotifier_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_chainrpc_chainnotifier_proto_goTypes = []interface{}{
	(*ConfRequest)(nil),  // 0: chainrpc.ConfRequest
	(*ConfDetails)(nil),  // 1: chainrpc.ConfDetails
	(*Reorg)(nil),        // 2: chainrpc.Reorg
	(*ConfEvent)(nil),    // 3: chainrpc.ConfEvent
	(*Outpoint)(nil),     // 4: chainrpc.Outpoint
	(*SpendRequest)(nil), // 5: chainrpc.SpendRequest
	(*SpendDetails)(nil), // 6: chainrpc.SpendDetails
	(*SpendEvent)(nil),   // 7: chainrpc.SpendEvent
	(*BlockEpoch)(nil),   // 8: chainrpc.BlockEpoch
}
var file_chainrpc_chainnotifier_proto_depIdxs = []int32{
	1, // 0: chainrpc.ConfEvent.conf:type_name -> chainrpc.ConfDetails
	2, // 1: chainrpc.ConfEvent.reorg:type_name -> chainrpc.Reorg
	4, // 2: chainrpc.SpendRequest.outpoint:type_name -> chainrpc.Outpoint
	4, // 3: chainrpc.SpendDetails.spending_outpoint:type_name -> chainrpc.Outpoint
	6, // 4: chainrpc.SpendEvent.spend:type_name -> chainrpc.SpendDetails
	2, // 5: chainrpc.SpendEvent.reorg:type_name -> chainrpc.Reorg
	0, // 6: chainrpc.ChainNotifier.RegisterConfirmationsNtfn:input_type -> chainrpc.ConfRequest
	5, // 7: chainrpc.ChainNotifier.RegisterSpendNtfn:input_type -> chainrpc.SpendRequest
	8, // 8: chainrpc.ChainNotifier.RegisterBlockEpochNtfn:input_type -> chainrpc.BlockEpoch
	3, // 9: chainrpc.ChainNotifier.RegisterConfirmationsNtfn:output_type -> chainrpc.ConfEvent
	7, // 10: chainrpc.ChainNotifier.RegisterSpendNtfn:output_type -> chainrpc.SpendEvent
	8, // 11: chainrpc.ChainNotifier.RegisterBlockEpochNtfn:output_type -> chainrpc.BlockEpoch
	9, // [9:12] is the sub-list for method output_type
	6, // [6:9] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_chainrpc_chainnotifier_proto_init() }
func file_chainrpc_chainnotifier_proto_init() {
	if File_chainrpc_chainnotifier_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_chainrpc_chainnotifier_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConfRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_chainnotifier_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConfDetails); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_chainnotifier_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Reorg); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_chainnotifier_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConfEvent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_chainnotifier_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Outpoint); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_chainnotifier_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SpendRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_chainnotifier_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SpendDetails); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_chainnotifier_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SpendEvent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chainrpc_chainnotifier_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockEpoch); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_chainrpc_chainnotifier_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*ConfEvent_Conf)(nil),
		(*ConfEvent_Reorg)(nil),
	}
	file_chainrpc_chainnotifier_proto_msgTypes[7].OneofWrappers = []interface{}{
		(*SpendEvent_Spend)(nil),
		(*SpendEvent_Reorg)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_chainrpc_chainnotifier_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_chainrpc_chainnotifier_proto_goTypes,
		DependencyIndexes: file_chainrpc_chainnotifier_proto_depIdxs,
		MessageInfos:      file_chainrpc_chainnotifier_proto_msgTypes,
	}.Build()
	File_chainrpc_chainnotifier_proto = out.File
	file_chainrpc_chainnotifier_proto_rawDesc = nil
	file_chainrpc_chainnotifier_proto_goTypes = nil
	file_chainrpc_chainnotifier_proto_depIdxs = nil
}
