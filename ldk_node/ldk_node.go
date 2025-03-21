package ldk_node

// #include <ldk_node.h>
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync/atomic"
	"unsafe"
)

type RustBuffer = C.RustBuffer

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() int
	Capacity() int
}

func RustBufferFromExternal(b RustBufferI) RustBuffer {
	return RustBuffer{
		capacity: C.int(b.Capacity()),
		len:      C.int(b.Len()),
		data:     (*C.uchar)(b.Data()),
	}
}

func (cb RustBuffer) Capacity() int {
	return int(cb.capacity)
}

func (cb RustBuffer) Len() int {
	return int(cb.len)
}

func (cb RustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.data)
}

func (cb RustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.data), C.int(cb.len))
	return bytes.NewReader(b)
}

func (cb RustBuffer) Free() {
	rustCall(func(status *C.RustCallStatus) bool {
		C.ffi_ldk_node_rustbuffer_free(cb, status)
		return false
	})
}

func (cb RustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.data), C.int(cb.len))
}

func stringToRustBuffer(str string) RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) RustBuffer {
	if len(b) == 0 {
		return RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes{
		len:  C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}

	return rustCall(func(status *C.RustCallStatus) RustBuffer {
		return C.ffi_ldk_node_rustbuffer_from_bytes(foreign, status)
	})
}

type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) RustBuffer
}

type FfiConverter[GoType any, FfiType any] interface {
	Lift(value FfiType) GoType
	Lower(value GoType) FfiType
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

type FfiRustBufConverter[GoType any, FfiType any] interface {
	FfiConverter[GoType, FfiType]
	BufReader[GoType]
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.Write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return bytesToRustBuffer(bytes)
}

func LiftFromRustBuffer[GoType any](bufReader BufReader[GoType], rbuf RustBufferI) GoType {
	defer rbuf.Free()
	reader := rbuf.AsReader()
	item := bufReader.Read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}

func rustCallWithError[U any](converter BufLifter[error], callback func(*C.RustCallStatus) U) (U, error) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)

	return returnValue, err
}

func checkCallStatus(converter BufLifter[error], status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		return converter.Lift(status.errorBuf)
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a rustbuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(status.errorBuf)))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError(nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
}

func writeInt8(writer io.Writer, value int8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint8(writer io.Writer, value uint8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt16(writer io.Writer, value int16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint16(writer io.Writer, value uint16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt32(writer io.Writer, value int32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint32(writer io.Writer, value uint32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt64(writer io.Writer, value int64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint64(writer io.Writer, value uint64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat32(writer io.Writer, value float32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat64(writer io.Writer, value float64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func readInt8(reader io.Reader) int8 {
	var result int8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint8(reader io.Reader) uint8 {
	var result uint8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt16(reader io.Reader) int16 {
	var result int16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint16(reader io.Reader) uint16 {
	var result uint16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt32(reader io.Reader) int32 {
	var result int32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint32(reader io.Reader) uint32 {
	var result uint32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt64(reader io.Reader) int64 {
	var result int64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint64(reader io.Reader) uint64 {
	var result uint64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat32(reader io.Reader) float32 {
	var result float32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat64(reader io.Reader) float64 {
	var result float64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func init() {

	uniffiCheckChecksums()
}

func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 24
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_ldk_node_uniffi_contract_version(uniffiStatus)
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("ldk_node: UniFFI contract version mismatch")
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_func_default_config(uniffiStatus)
		})
		if checksum != 62308 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_func_default_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_func_generate_entropy_mnemonic(uniffiStatus)
		})
		if checksum != 7251 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_func_generate_entropy_mnemonic: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_claim_for_hash(uniffiStatus)
		})
		if checksum != 40017 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_claim_for_hash: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_fail_for_hash(uniffiStatus)
		})
		if checksum != 22118 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_fail_for_hash: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_receive(uniffiStatus)
		})
		if checksum != 44074 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_receive: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_receive_for_hash(uniffiStatus)
		})
		if checksum != 4239 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_receive_for_hash: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_receive_variable_amount(uniffiStatus)
		})
		if checksum != 50172 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_receive_variable_amount: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_receive_variable_amount_for_hash(uniffiStatus)
		})
		if checksum != 60371 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_receive_variable_amount_for_hash: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_receive_variable_amount_via_jit_channel(uniffiStatus)
		})
		if checksum != 6695 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_receive_variable_amount_via_jit_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_receive_via_jit_channel(uniffiStatus)
		})
		if checksum != 10006 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_receive_via_jit_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_send(uniffiStatus)
		})
		if checksum != 20666 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_send: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_send_probes(uniffiStatus)
		})
		if checksum != 1481 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_send_probes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_send_probes_using_amount(uniffiStatus)
		})
		if checksum != 40103 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_send_probes_using_amount: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt11payment_send_using_amount(uniffiStatus)
		})
		if checksum != 63159 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt11payment_send_using_amount: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt12payment_initiate_refund(uniffiStatus)
		})
		if checksum != 57799 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt12payment_initiate_refund: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt12payment_receive(uniffiStatus)
		})
		if checksum != 18429 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt12payment_receive: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt12payment_receive_variable_amount(uniffiStatus)
		})
		if checksum != 29252 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt12payment_receive_variable_amount: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt12payment_request_refund_payment(uniffiStatus)
		})
		if checksum != 50892 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt12payment_request_refund_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt12payment_send(uniffiStatus)
		})
		if checksum != 64402 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt12payment_send: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_bolt12payment_send_using_amount(uniffiStatus)
		})
		if checksum != 26048 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_bolt12payment_send_using_amount: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_build(uniffiStatus)
		})
		if checksum != 46255 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_build: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_build_with_fs_store(uniffiStatus)
		})
		if checksum != 15423 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_build_with_fs_store: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_build_with_vss_store(uniffiStatus)
		})
		if checksum != 10256 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_build_with_vss_store: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_build_with_vss_store_and_fixed_headers(uniffiStatus)
		})
		if checksum != 17368 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_build_with_vss_store_and_fixed_headers: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_migrate_storage(uniffiStatus)
		})
		if checksum != 57962 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_migrate_storage: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_reset_state(uniffiStatus)
		})
		if checksum != 15117 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_reset_state: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_restore_encoded_channel_monitors(uniffiStatus)
		})
		if checksum != 34188 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_restore_encoded_channel_monitors: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_chain_source_bitcoind_rpc(uniffiStatus)
		})
		if checksum != 2111 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_chain_source_bitcoind_rpc: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_chain_source_esplora(uniffiStatus)
		})
		if checksum != 21678 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_chain_source_esplora: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_entropy_bip39_mnemonic(uniffiStatus)
		})
		if checksum != 35659 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_entropy_bip39_mnemonic: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_entropy_seed_bytes(uniffiStatus)
		})
		if checksum != 26795 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_entropy_seed_bytes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_entropy_seed_path(uniffiStatus)
		})
		if checksum != 64056 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_entropy_seed_path: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_gossip_source_p2p(uniffiStatus)
		})
		if checksum != 9279 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_gossip_source_p2p: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_gossip_source_rgs(uniffiStatus)
		})
		if checksum != 64312 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_gossip_source_rgs: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_liquidity_source_lsps2(uniffiStatus)
		})
		if checksum != 26412 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_liquidity_source_lsps2: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_listening_addresses(uniffiStatus)
		})
		if checksum != 18689 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_listening_addresses: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_network(uniffiStatus)
		})
		if checksum != 38526 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_network: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_node_alias(uniffiStatus)
		})
		if checksum != 52925 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_node_alias: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_builder_set_storage_dir_path(uniffiStatus)
		})
		if checksum != 59019 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_builder_set_storage_dir_path: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_networkgraph_channel(uniffiStatus)
		})
		if checksum != 19532 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_networkgraph_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_networkgraph_list_channels(uniffiStatus)
		})
		if checksum != 13583 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_networkgraph_list_channels: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_networkgraph_list_nodes(uniffiStatus)
		})
		if checksum != 21167 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_networkgraph_list_nodes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_networkgraph_node(uniffiStatus)
		})
		if checksum != 16507 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_networkgraph_node: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_bolt11_payment(uniffiStatus)
		})
		if checksum != 41402 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_bolt11_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_bolt12_payment(uniffiStatus)
		})
		if checksum != 49254 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_bolt12_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_close_channel(uniffiStatus)
		})
		if checksum != 1420 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_close_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_config(uniffiStatus)
		})
		if checksum != 15339 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_connect(uniffiStatus)
		})
		if checksum != 15352 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_connect: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_disconnect(uniffiStatus)
		})
		if checksum != 47760 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_disconnect: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_event_handled(uniffiStatus)
		})
		if checksum != 47939 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_event_handled: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_force_close_all_channels_without_broadcasting_txn(uniffiStatus)
		})
		if checksum != 12970 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_force_close_all_channels_without_broadcasting_txn: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_force_close_channel(uniffiStatus)
		})
		if checksum != 48609 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_force_close_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_get_encoded_channel_monitors(uniffiStatus)
		})
		if checksum != 8831 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_get_encoded_channel_monitors: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_list_balances(uniffiStatus)
		})
		if checksum != 24919 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_list_balances: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_list_channels(uniffiStatus)
		})
		if checksum != 62491 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_list_channels: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_list_payments(uniffiStatus)
		})
		if checksum != 47765 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_list_payments: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_list_peers(uniffiStatus)
		})
		if checksum != 12947 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_list_peers: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_listening_addresses(uniffiStatus)
		})
		if checksum != 55483 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_listening_addresses: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_network_graph(uniffiStatus)
		})
		if checksum != 2695 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_network_graph: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_next_event(uniffiStatus)
		})
		if checksum != 46767 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_next_event: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_node_alias(uniffiStatus)
		})
		if checksum != 26583 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_node_alias: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_node_id(uniffiStatus)
		})
		if checksum != 34585 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_node_id: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_onchain_payment(uniffiStatus)
		})
		if checksum != 6092 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_onchain_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_open_announced_channel(uniffiStatus)
		})
		if checksum != 609 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_open_announced_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_open_channel(uniffiStatus)
		})
		if checksum != 57394 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_open_channel: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_payment(uniffiStatus)
		})
		if checksum != 59271 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_remove_payment(uniffiStatus)
		})
		if checksum != 8539 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_remove_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_sign_message(uniffiStatus)
		})
		if checksum != 35420 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_sign_message: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_spontaneous_payment(uniffiStatus)
		})
		if checksum != 37403 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_spontaneous_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_start(uniffiStatus)
		})
		if checksum != 21524 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_start: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_status(uniffiStatus)
		})
		if checksum != 46945 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_status: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_stop(uniffiStatus)
		})
		if checksum != 12389 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_stop: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_sync_wallets(uniffiStatus)
		})
		if checksum != 29385 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_sync_wallets: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_unified_qr_payment(uniffiStatus)
		})
		if checksum != 9837 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_unified_qr_payment: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_update_channel_config(uniffiStatus)
		})
		if checksum != 42796 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_update_channel_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_update_fee_estimates(uniffiStatus)
		})
		if checksum != 52795 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_update_fee_estimates: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_verify_signature(uniffiStatus)
		})
		if checksum != 56945 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_verify_signature: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_node_wait_next_event(uniffiStatus)
		})
		if checksum != 30900 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_node_wait_next_event: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_onchainpayment_new_address(uniffiStatus)
		})
		if checksum != 23077 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_onchainpayment_new_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_onchainpayment_send_all_to_address(uniffiStatus)
		})
		if checksum != 35766 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_onchainpayment_send_all_to_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_onchainpayment_send_to_address(uniffiStatus)
		})
		if checksum != 3083 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_onchainpayment_send_to_address: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_spontaneouspayment_send(uniffiStatus)
		})
		if checksum != 50859 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_spontaneouspayment_send: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_spontaneouspayment_send_probes(uniffiStatus)
		})
		if checksum != 32884 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_spontaneouspayment_send_probes: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_unifiedqrpayment_receive(uniffiStatus)
		})
		if checksum != 18681 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_unifiedqrpayment_receive: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_method_unifiedqrpayment_send(uniffiStatus)
		})
		if checksum != 1473 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_method_unifiedqrpayment_send: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_constructor_builder_from_config(uniffiStatus)
		})
		if checksum != 56443 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_constructor_builder_from_config: UniFFI API checksum mismatch")
		}
	}
	{
		checksum := rustCall(func(uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_ldk_node_checksum_constructor_builder_new(uniffiStatus)
		})
		if checksum != 48442 {
			// If this happens try cleaning and rebuilding your project
			panic("ldk_node: uniffi_ldk_node_checksum_constructor_builder_new: UniFFI API checksum mismatch")
		}
	}
}

type FfiConverterUint8 struct{}

var FfiConverterUint8INSTANCE = FfiConverterUint8{}

func (FfiConverterUint8) Lower(value uint8) C.uint8_t {
	return C.uint8_t(value)
}

func (FfiConverterUint8) Write(writer io.Writer, value uint8) {
	writeUint8(writer, value)
}

func (FfiConverterUint8) Lift(value C.uint8_t) uint8 {
	return uint8(value)
}

func (FfiConverterUint8) Read(reader io.Reader) uint8 {
	return readUint8(reader)
}

type FfiDestroyerUint8 struct{}

func (FfiDestroyerUint8) Destroy(_ uint8) {}

type FfiConverterUint16 struct{}

var FfiConverterUint16INSTANCE = FfiConverterUint16{}

func (FfiConverterUint16) Lower(value uint16) C.uint16_t {
	return C.uint16_t(value)
}

func (FfiConverterUint16) Write(writer io.Writer, value uint16) {
	writeUint16(writer, value)
}

func (FfiConverterUint16) Lift(value C.uint16_t) uint16 {
	return uint16(value)
}

func (FfiConverterUint16) Read(reader io.Reader) uint16 {
	return readUint16(reader)
}

type FfiDestroyerUint16 struct{}

func (FfiDestroyerUint16) Destroy(_ uint16) {}

type FfiConverterUint32 struct{}

var FfiConverterUint32INSTANCE = FfiConverterUint32{}

func (FfiConverterUint32) Lower(value uint32) C.uint32_t {
	return C.uint32_t(value)
}

func (FfiConverterUint32) Write(writer io.Writer, value uint32) {
	writeUint32(writer, value)
}

func (FfiConverterUint32) Lift(value C.uint32_t) uint32 {
	return uint32(value)
}

func (FfiConverterUint32) Read(reader io.Reader) uint32 {
	return readUint32(reader)
}

type FfiDestroyerUint32 struct{}

func (FfiDestroyerUint32) Destroy(_ uint32) {}

type FfiConverterUint64 struct{}

var FfiConverterUint64INSTANCE = FfiConverterUint64{}

func (FfiConverterUint64) Lower(value uint64) C.uint64_t {
	return C.uint64_t(value)
}

func (FfiConverterUint64) Write(writer io.Writer, value uint64) {
	writeUint64(writer, value)
}

func (FfiConverterUint64) Lift(value C.uint64_t) uint64 {
	return uint64(value)
}

func (FfiConverterUint64) Read(reader io.Reader) uint64 {
	return readUint64(reader)
}

type FfiDestroyerUint64 struct{}

func (FfiDestroyerUint64) Destroy(_ uint64) {}

type FfiConverterBool struct{}

var FfiConverterBoolINSTANCE = FfiConverterBool{}

func (FfiConverterBool) Lower(value bool) C.int8_t {
	if value {
		return C.int8_t(1)
	}
	return C.int8_t(0)
}

func (FfiConverterBool) Write(writer io.Writer, value bool) {
	if value {
		writeInt8(writer, 1)
	} else {
		writeInt8(writer, 0)
	}
}

func (FfiConverterBool) Lift(value C.int8_t) bool {
	return value != 0
}

func (FfiConverterBool) Read(reader io.Reader) bool {
	return readInt8(reader) != 0
}

type FfiDestroyerBool struct{}

func (FfiDestroyerBool) Destroy(_ bool) {}

type FfiConverterString struct{}

var FfiConverterStringINSTANCE = FfiConverterString{}

func (FfiConverterString) Lift(rb RustBufferI) string {
	defer rb.Free()
	reader := rb.AsReader()
	b, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("reading reader: %w", err))
	}
	return string(b)
}

func (FfiConverterString) Read(reader io.Reader) string {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) RustBuffer {
	return stringToRustBuffer(value)
}

func (FfiConverterString) Write(writer io.Writer, value string) {
	if len(value) > math.MaxInt32 {
		panic("String is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := io.WriteString(writer, value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing string, expected %d, written %d", len(value), write_length))
	}
}

type FfiDestroyerString struct{}

func (FfiDestroyerString) Destroy(_ string) {}

// Below is an implementation of synchronization requirements outlined in the link.
// https://github.com/mozilla/uniffi-rs/blob/0dc031132d9493ca812c3af6e7dd60ad2ea95bf0/uniffi_bindgen/src/bindings/kotlin/templates/ObjectRuntime.kt#L31

type FfiObject struct {
	pointer      unsafe.Pointer
	callCounter  atomic.Int64
	freeFunction func(unsafe.Pointer, *C.RustCallStatus)
	destroyed    atomic.Bool
}

func newFfiObject(pointer unsafe.Pointer, freeFunction func(unsafe.Pointer, *C.RustCallStatus)) FfiObject {
	return FfiObject{
		pointer:      pointer,
		freeFunction: freeFunction,
	}
}

func (ffiObject *FfiObject) incrementPointer(debugName string) unsafe.Pointer {
	for {
		counter := ffiObject.callCounter.Load()
		if counter <= -1 {
			panic(fmt.Errorf("%v object has already been destroyed", debugName))
		}
		if counter == math.MaxInt64 {
			panic(fmt.Errorf("%v object call counter would overflow", debugName))
		}
		if ffiObject.callCounter.CompareAndSwap(counter, counter+1) {
			break
		}
	}

	return ffiObject.pointer
}

func (ffiObject *FfiObject) decrementPointer() {
	if ffiObject.callCounter.Add(-1) == -1 {
		ffiObject.freeRustArcPtr()
	}
}

func (ffiObject *FfiObject) destroy() {
	if ffiObject.destroyed.CompareAndSwap(false, true) {
		if ffiObject.callCounter.Add(-1) == -1 {
			ffiObject.freeRustArcPtr()
		}
	}
}

func (ffiObject *FfiObject) freeRustArcPtr() {
	rustCall(func(status *C.RustCallStatus) int32 {
		ffiObject.freeFunction(ffiObject.pointer, status)
		return 0
	})
}

type Bolt11Payment struct {
	ffiObject FfiObject
}

func (_self *Bolt11Payment) ClaimForHash(paymentHash PaymentHash, claimableAmountMsat uint64, preimage PaymentPreimage) error {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_bolt11payment_claim_for_hash(
			_pointer, FfiConverterTypePaymentHashINSTANCE.Lower(paymentHash), FfiConverterUint64INSTANCE.Lower(claimableAmountMsat), FfiConverterTypePaymentPreimageINSTANCE.Lower(preimage), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Bolt11Payment) FailForHash(paymentHash PaymentHash) error {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_bolt11payment_fail_for_hash(
			_pointer, FfiConverterTypePaymentHashINSTANCE.Lower(paymentHash), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Bolt11Payment) Receive(amountMsat uint64, description string, expirySecs uint32) (Bolt11Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_receive(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterStringINSTANCE.Lower(description), FfiConverterUint32INSTANCE.Lower(expirySecs), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt11Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt11InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) ReceiveForHash(amountMsat uint64, description string, expirySecs uint32, paymentHash PaymentHash) (Bolt11Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_receive_for_hash(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterStringINSTANCE.Lower(description), FfiConverterUint32INSTANCE.Lower(expirySecs), FfiConverterTypePaymentHashINSTANCE.Lower(paymentHash), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt11Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt11InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) ReceiveVariableAmount(description string, expirySecs uint32) (Bolt11Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_receive_variable_amount(
			_pointer, FfiConverterStringINSTANCE.Lower(description), FfiConverterUint32INSTANCE.Lower(expirySecs), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt11Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt11InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) ReceiveVariableAmountForHash(description string, expirySecs uint32, paymentHash PaymentHash) (Bolt11Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_receive_variable_amount_for_hash(
			_pointer, FfiConverterStringINSTANCE.Lower(description), FfiConverterUint32INSTANCE.Lower(expirySecs), FfiConverterTypePaymentHashINSTANCE.Lower(paymentHash), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt11Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt11InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) ReceiveVariableAmountViaJitChannel(description string, expirySecs uint32, maxProportionalLspFeeLimitPpmMsat *uint64) (Bolt11Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_receive_variable_amount_via_jit_channel(
			_pointer, FfiConverterStringINSTANCE.Lower(description), FfiConverterUint32INSTANCE.Lower(expirySecs), FfiConverterOptionalUint64INSTANCE.Lower(maxProportionalLspFeeLimitPpmMsat), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt11Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt11InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) ReceiveViaJitChannel(amountMsat uint64, description string, expirySecs uint32, maxLspFeeLimitMsat *uint64) (Bolt11Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_receive_via_jit_channel(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterStringINSTANCE.Lower(description), FfiConverterUint32INSTANCE.Lower(expirySecs), FfiConverterOptionalUint64INSTANCE.Lower(maxLspFeeLimitMsat), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt11Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt11InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) Send(invoice Bolt11Invoice, sendingParameters *SendingParameters) (PaymentId, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_send(
			_pointer, FfiConverterTypeBolt11InvoiceINSTANCE.Lower(invoice), FfiConverterOptionalTypeSendingParametersINSTANCE.Lower(sendingParameters), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PaymentId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePaymentIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt11Payment) SendProbes(invoice Bolt11Invoice) error {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_bolt11payment_send_probes(
			_pointer, FfiConverterTypeBolt11InvoiceINSTANCE.Lower(invoice), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Bolt11Payment) SendProbesUsingAmount(invoice Bolt11Invoice, amountMsat uint64) error {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_bolt11payment_send_probes_using_amount(
			_pointer, FfiConverterTypeBolt11InvoiceINSTANCE.Lower(invoice), FfiConverterUint64INSTANCE.Lower(amountMsat), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Bolt11Payment) SendUsingAmount(invoice Bolt11Invoice, amountMsat uint64, sendingParameters *SendingParameters) (PaymentId, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt11Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt11payment_send_using_amount(
			_pointer, FfiConverterTypeBolt11InvoiceINSTANCE.Lower(invoice), FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterOptionalTypeSendingParametersINSTANCE.Lower(sendingParameters), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PaymentId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePaymentIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (object *Bolt11Payment) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterBolt11Payment struct{}

var FfiConverterBolt11PaymentINSTANCE = FfiConverterBolt11Payment{}

func (c FfiConverterBolt11Payment) Lift(pointer unsafe.Pointer) *Bolt11Payment {
	result := &Bolt11Payment{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_bolt11payment(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*Bolt11Payment).Destroy)
	return result
}

func (c FfiConverterBolt11Payment) Read(reader io.Reader) *Bolt11Payment {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterBolt11Payment) Lower(value *Bolt11Payment) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Bolt11Payment")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterBolt11Payment) Write(writer io.Writer, value *Bolt11Payment) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerBolt11Payment struct{}

func (_ FfiDestroyerBolt11Payment) Destroy(value *Bolt11Payment) {
	value.Destroy()
}

type Bolt12Payment struct {
	ffiObject FfiObject
}

func (_self *Bolt12Payment) InitiateRefund(amountMsat uint64, expirySecs uint32, quantity *uint64, payerNote *string) (Refund, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt12Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt12payment_initiate_refund(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterUint32INSTANCE.Lower(expirySecs), FfiConverterOptionalUint64INSTANCE.Lower(quantity), FfiConverterOptionalStringINSTANCE.Lower(payerNote), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Refund
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeRefundINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt12Payment) Receive(amountMsat uint64, description string, expirySecs *uint32, quantity *uint64) (Offer, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt12Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt12payment_receive(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterStringINSTANCE.Lower(description), FfiConverterOptionalUint32INSTANCE.Lower(expirySecs), FfiConverterOptionalUint64INSTANCE.Lower(quantity), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Offer
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeOfferINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt12Payment) ReceiveVariableAmount(description string, expirySecs *uint32) (Offer, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt12Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt12payment_receive_variable_amount(
			_pointer, FfiConverterStringINSTANCE.Lower(description), FfiConverterOptionalUint32INSTANCE.Lower(expirySecs), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Offer
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeOfferINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt12Payment) RequestRefundPayment(refund Refund) (Bolt12Invoice, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt12Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt12payment_request_refund_payment(
			_pointer, FfiConverterTypeRefundINSTANCE.Lower(refund), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Bolt12Invoice
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeBolt12InvoiceINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt12Payment) Send(offer Offer, quantity *uint64, payerNote *string) (PaymentId, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt12Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt12payment_send(
			_pointer, FfiConverterTypeOfferINSTANCE.Lower(offer), FfiConverterOptionalUint64INSTANCE.Lower(quantity), FfiConverterOptionalStringINSTANCE.Lower(payerNote), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PaymentId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePaymentIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Bolt12Payment) SendUsingAmount(offer Offer, amountMsat uint64, quantity *uint64, payerNote *string) (PaymentId, error) {
	_pointer := _self.ffiObject.incrementPointer("*Bolt12Payment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_bolt12payment_send_using_amount(
			_pointer, FfiConverterTypeOfferINSTANCE.Lower(offer), FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterOptionalUint64INSTANCE.Lower(quantity), FfiConverterOptionalStringINSTANCE.Lower(payerNote), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PaymentId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePaymentIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (object *Bolt12Payment) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterBolt12Payment struct{}

var FfiConverterBolt12PaymentINSTANCE = FfiConverterBolt12Payment{}

func (c FfiConverterBolt12Payment) Lift(pointer unsafe.Pointer) *Bolt12Payment {
	result := &Bolt12Payment{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_bolt12payment(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*Bolt12Payment).Destroy)
	return result
}

func (c FfiConverterBolt12Payment) Read(reader io.Reader) *Bolt12Payment {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterBolt12Payment) Lower(value *Bolt12Payment) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Bolt12Payment")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterBolt12Payment) Write(writer io.Writer, value *Bolt12Payment) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerBolt12Payment struct{}

func (_ FfiDestroyerBolt12Payment) Destroy(value *Bolt12Payment) {
	value.Destroy()
}

type Builder struct {
	ffiObject FfiObject
}

func NewBuilder() *Builder {
	return FfiConverterBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_constructor_builder_new(_uniffiStatus)
	}))
}

func BuilderFromConfig(config Config) *Builder {
	return FfiConverterBuilderINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_constructor_builder_from_config(FfiConverterTypeConfigINSTANCE.Lower(config), _uniffiStatus)
	}))
}

func (_self *Builder) Build() (*Node, error) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_builder_build(
			_pointer, _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *Node
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterNodeINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Builder) BuildWithFsStore() (*Node, error) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_builder_build_with_fs_store(
			_pointer, _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *Node
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterNodeINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Builder) BuildWithVssStore(vssUrl string, storeId string, lnurlAuthServerUrl string, fixedHeaders map[string]string) (*Node, error) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_builder_build_with_vss_store(
			_pointer, FfiConverterStringINSTANCE.Lower(vssUrl), FfiConverterStringINSTANCE.Lower(storeId), FfiConverterStringINSTANCE.Lower(lnurlAuthServerUrl), FfiConverterMapStringStringINSTANCE.Lower(fixedHeaders), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *Node
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterNodeINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Builder) BuildWithVssStoreAndFixedHeaders(vssUrl string, storeId string, fixedHeaders map[string]string) (*Node, error) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_builder_build_with_vss_store_and_fixed_headers(
			_pointer, FfiConverterStringINSTANCE.Lower(vssUrl), FfiConverterStringINSTANCE.Lower(storeId), FfiConverterMapStringStringINSTANCE.Lower(fixedHeaders), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue *Node
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterNodeINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Builder) MigrateStorage(what MigrateStorage) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_migrate_storage(
			_pointer, FfiConverterTypeMigrateStorageINSTANCE.Lower(what), _uniffiStatus)
		return false
	})
}

func (_self *Builder) ResetState(what ResetState) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_reset_state(
			_pointer, FfiConverterTypeResetStateINSTANCE.Lower(what), _uniffiStatus)
		return false
	})
}

func (_self *Builder) RestoreEncodedChannelMonitors(monitors []KeyValue) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_restore_encoded_channel_monitors(
			_pointer, FfiConverterSequenceTypeKeyValueINSTANCE.Lower(monitors), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetChainSourceBitcoindRpc(rpcHost string, rpcPort uint16, rpcUser string, rpcPassword string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_chain_source_bitcoind_rpc(
			_pointer, FfiConverterStringINSTANCE.Lower(rpcHost), FfiConverterUint16INSTANCE.Lower(rpcPort), FfiConverterStringINSTANCE.Lower(rpcUser), FfiConverterStringINSTANCE.Lower(rpcPassword), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetChainSourceEsplora(serverUrl string, config *EsploraSyncConfig) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_chain_source_esplora(
			_pointer, FfiConverterStringINSTANCE.Lower(serverUrl), FfiConverterOptionalTypeEsploraSyncConfigINSTANCE.Lower(config), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetEntropyBip39Mnemonic(mnemonic Mnemonic, passphrase *string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_entropy_bip39_mnemonic(
			_pointer, FfiConverterTypeMnemonicINSTANCE.Lower(mnemonic), FfiConverterOptionalStringINSTANCE.Lower(passphrase), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetEntropySeedBytes(seedBytes []uint8) error {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_entropy_seed_bytes(
			_pointer, FfiConverterSequenceUint8INSTANCE.Lower(seedBytes), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Builder) SetEntropySeedPath(seedPath string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_entropy_seed_path(
			_pointer, FfiConverterStringINSTANCE.Lower(seedPath), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetGossipSourceP2p() {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_gossip_source_p2p(
			_pointer, _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetGossipSourceRgs(rgsServerUrl string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_gossip_source_rgs(
			_pointer, FfiConverterStringINSTANCE.Lower(rgsServerUrl), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetLiquiditySourceLsps2(address SocketAddress, nodeId PublicKey, token *string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_liquidity_source_lsps2(
			_pointer, FfiConverterTypeSocketAddressINSTANCE.Lower(address), FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), FfiConverterOptionalStringINSTANCE.Lower(token), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetListeningAddresses(listeningAddresses []SocketAddress) error {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_listening_addresses(
			_pointer, FfiConverterSequenceTypeSocketAddressINSTANCE.Lower(listeningAddresses), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Builder) SetNetwork(network Network) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_network(
			_pointer, FfiConverterTypeNetworkINSTANCE.Lower(network), _uniffiStatus)
		return false
	})
}

func (_self *Builder) SetNodeAlias(nodeAlias string) error {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeBuildError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_node_alias(
			_pointer, FfiConverterStringINSTANCE.Lower(nodeAlias), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Builder) SetStorageDirPath(storageDirPath string) {
	_pointer := _self.ffiObject.incrementPointer("*Builder")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_builder_set_storage_dir_path(
			_pointer, FfiConverterStringINSTANCE.Lower(storageDirPath), _uniffiStatus)
		return false
	})
}

func (object *Builder) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterBuilder struct{}

var FfiConverterBuilderINSTANCE = FfiConverterBuilder{}

func (c FfiConverterBuilder) Lift(pointer unsafe.Pointer) *Builder {
	result := &Builder{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_builder(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*Builder).Destroy)
	return result
}

func (c FfiConverterBuilder) Read(reader io.Reader) *Builder {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterBuilder) Lower(value *Builder) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Builder")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterBuilder) Write(writer io.Writer, value *Builder) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerBuilder struct{}

func (_ FfiDestroyerBuilder) Destroy(value *Builder) {
	value.Destroy()
}

type NetworkGraph struct {
	ffiObject FfiObject
}

func (_self *NetworkGraph) Channel(shortChannelId uint64) *ChannelInfo {
	_pointer := _self.ffiObject.incrementPointer("*NetworkGraph")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalTypeChannelInfoINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_networkgraph_channel(
			_pointer, FfiConverterUint64INSTANCE.Lower(shortChannelId), _uniffiStatus)
	}))
}

func (_self *NetworkGraph) ListChannels() []uint64 {
	_pointer := _self.ffiObject.incrementPointer("*NetworkGraph")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceUint64INSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_networkgraph_list_channels(
			_pointer, _uniffiStatus)
	}))
}

func (_self *NetworkGraph) ListNodes() []NodeId {
	_pointer := _self.ffiObject.incrementPointer("*NetworkGraph")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTypeNodeIdINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_networkgraph_list_nodes(
			_pointer, _uniffiStatus)
	}))
}

func (_self *NetworkGraph) Node(nodeId NodeId) *NodeInfo {
	_pointer := _self.ffiObject.incrementPointer("*NetworkGraph")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalTypeNodeInfoINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_networkgraph_node(
			_pointer, FfiConverterTypeNodeIdINSTANCE.Lower(nodeId), _uniffiStatus)
	}))
}

func (object *NetworkGraph) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterNetworkGraph struct{}

var FfiConverterNetworkGraphINSTANCE = FfiConverterNetworkGraph{}

func (c FfiConverterNetworkGraph) Lift(pointer unsafe.Pointer) *NetworkGraph {
	result := &NetworkGraph{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_networkgraph(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*NetworkGraph).Destroy)
	return result
}

func (c FfiConverterNetworkGraph) Read(reader io.Reader) *NetworkGraph {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterNetworkGraph) Lower(value *NetworkGraph) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*NetworkGraph")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterNetworkGraph) Write(writer io.Writer, value *NetworkGraph) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerNetworkGraph struct{}

func (_ FfiDestroyerNetworkGraph) Destroy(value *NetworkGraph) {
	value.Destroy()
}

type Node struct {
	ffiObject FfiObject
}

func (_self *Node) Bolt11Payment() *Bolt11Payment {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBolt11PaymentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_node_bolt11_payment(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) Bolt12Payment() *Bolt12Payment {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBolt12PaymentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_node_bolt12_payment(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) CloseChannel(userChannelId UserChannelId, counterpartyNodeId PublicKey) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_close_channel(
			_pointer, FfiConverterTypeUserChannelIdINSTANCE.Lower(userChannelId), FfiConverterTypePublicKeyINSTANCE.Lower(counterpartyNodeId), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) Config() Config {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeConfigINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_config(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) Connect(nodeId PublicKey, address SocketAddress, persist bool) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_connect(
			_pointer, FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), FfiConverterTypeSocketAddressINSTANCE.Lower(address), FfiConverterBoolINSTANCE.Lower(persist), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) Disconnect(nodeId PublicKey) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_disconnect(
			_pointer, FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) EventHandled() {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_event_handled(
			_pointer, _uniffiStatus)
		return false
	})
}

func (_self *Node) ForceCloseAllChannelsWithoutBroadcastingTxn() {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	rustCall(func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_force_close_all_channels_without_broadcasting_txn(
			_pointer, _uniffiStatus)
		return false
	})
}

func (_self *Node) ForceCloseChannel(userChannelId UserChannelId, counterpartyNodeId PublicKey, reason *string) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_force_close_channel(
			_pointer, FfiConverterTypeUserChannelIdINSTANCE.Lower(userChannelId), FfiConverterTypePublicKeyINSTANCE.Lower(counterpartyNodeId), FfiConverterOptionalStringINSTANCE.Lower(reason), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) GetEncodedChannelMonitors() ([]KeyValue, error) {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_get_encoded_channel_monitors(
			_pointer, _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue []KeyValue
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterSequenceTypeKeyValueINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Node) ListBalances() BalanceDetails {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeBalanceDetailsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_list_balances(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) ListChannels() []ChannelDetails {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTypeChannelDetailsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_list_channels(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) ListPayments() []PaymentDetails {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTypePaymentDetailsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_list_payments(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) ListPeers() []PeerDetails {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSequenceTypePeerDetailsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_list_peers(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) ListeningAddresses() *[]SocketAddress {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalSequenceTypeSocketAddressINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_listening_addresses(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) NetworkGraph() *NetworkGraph {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterNetworkGraphINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_node_network_graph(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) NextEvent() *Event {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalTypeEventINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_next_event(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) NodeAlias() *NodeAlias {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalTypeNodeAliasINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_node_alias(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) NodeId() PublicKey {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypePublicKeyINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_node_id(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) OnchainPayment() *OnchainPayment {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOnchainPaymentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_node_onchain_payment(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) OpenAnnouncedChannel(nodeId PublicKey, address SocketAddress, channelAmountSats uint64, pushToCounterpartyMsat *uint64, channelConfig *ChannelConfig) (UserChannelId, error) {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_open_announced_channel(
			_pointer, FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), FfiConverterTypeSocketAddressINSTANCE.Lower(address), FfiConverterUint64INSTANCE.Lower(channelAmountSats), FfiConverterOptionalUint64INSTANCE.Lower(pushToCounterpartyMsat), FfiConverterOptionalTypeChannelConfigINSTANCE.Lower(channelConfig), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue UserChannelId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeUserChannelIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Node) OpenChannel(nodeId PublicKey, address SocketAddress, channelAmountSats uint64, pushToCounterpartyMsat *uint64, channelConfig *ChannelConfig) (UserChannelId, error) {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_open_channel(
			_pointer, FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), FfiConverterTypeSocketAddressINSTANCE.Lower(address), FfiConverterUint64INSTANCE.Lower(channelAmountSats), FfiConverterOptionalUint64INSTANCE.Lower(pushToCounterpartyMsat), FfiConverterOptionalTypeChannelConfigINSTANCE.Lower(channelConfig), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue UserChannelId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeUserChannelIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *Node) Payment(paymentId PaymentId) *PaymentDetails {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterOptionalTypePaymentDetailsINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_payment(
			_pointer, FfiConverterTypePaymentIdINSTANCE.Lower(paymentId), _uniffiStatus)
	}))
}

func (_self *Node) RemovePayment(paymentId PaymentId) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_remove_payment(
			_pointer, FfiConverterTypePaymentIdINSTANCE.Lower(paymentId), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) SignMessage(msg []uint8) string {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterStringINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_sign_message(
			_pointer, FfiConverterSequenceUint8INSTANCE.Lower(msg), _uniffiStatus)
	}))
}

func (_self *Node) SpontaneousPayment() *SpontaneousPayment {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterSpontaneousPaymentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_node_spontaneous_payment(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) Start() error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_start(
			_pointer, _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) Status() NodeStatus {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeNodeStatusINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_status(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) Stop() error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_stop(
			_pointer, _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) SyncWallets() error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_sync_wallets(
			_pointer, _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) UnifiedQrPayment() *UnifiedQrPayment {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterUnifiedQrPaymentINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) unsafe.Pointer {
		return C.uniffi_ldk_node_fn_method_node_unified_qr_payment(
			_pointer, _uniffiStatus)
	}))
}

func (_self *Node) UpdateChannelConfig(userChannelId UserChannelId, counterpartyNodeId PublicKey, channelConfig ChannelConfig) error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_update_channel_config(
			_pointer, FfiConverterTypeUserChannelIdINSTANCE.Lower(userChannelId), FfiConverterTypePublicKeyINSTANCE.Lower(counterpartyNodeId), FfiConverterTypeChannelConfigINSTANCE.Lower(channelConfig), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) UpdateFeeEstimates() error {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_node_update_fee_estimates(
			_pointer, _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (_self *Node) VerifySignature(msg []uint8, sig string, pkey PublicKey) bool {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterBoolINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) C.int8_t {
		return C.uniffi_ldk_node_fn_method_node_verify_signature(
			_pointer, FfiConverterSequenceUint8INSTANCE.Lower(msg), FfiConverterStringINSTANCE.Lower(sig), FfiConverterTypePublicKeyINSTANCE.Lower(pkey), _uniffiStatus)
	}))
}

func (_self *Node) WaitNextEvent() Event {
	_pointer := _self.ffiObject.incrementPointer("*Node")
	defer _self.ffiObject.decrementPointer()
	return FfiConverterTypeEventINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_node_wait_next_event(
			_pointer, _uniffiStatus)
	}))
}

func (object *Node) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterNode struct{}

var FfiConverterNodeINSTANCE = FfiConverterNode{}

func (c FfiConverterNode) Lift(pointer unsafe.Pointer) *Node {
	result := &Node{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_node(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*Node).Destroy)
	return result
}

func (c FfiConverterNode) Read(reader io.Reader) *Node {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterNode) Lower(value *Node) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*Node")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterNode) Write(writer io.Writer, value *Node) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerNode struct{}

func (_ FfiDestroyerNode) Destroy(value *Node) {
	value.Destroy()
}

type OnchainPayment struct {
	ffiObject FfiObject
}

func (_self *OnchainPayment) NewAddress() (Address, error) {
	_pointer := _self.ffiObject.incrementPointer("*OnchainPayment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_onchainpayment_new_address(
			_pointer, _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Address
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeAddressINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *OnchainPayment) SendAllToAddress(address Address) (Txid, error) {
	_pointer := _self.ffiObject.incrementPointer("*OnchainPayment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_onchainpayment_send_all_to_address(
			_pointer, FfiConverterTypeAddressINSTANCE.Lower(address), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Txid
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeTxidINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *OnchainPayment) SendToAddress(address Address, amountSats uint64) (Txid, error) {
	_pointer := _self.ffiObject.incrementPointer("*OnchainPayment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_onchainpayment_send_to_address(
			_pointer, FfiConverterTypeAddressINSTANCE.Lower(address), FfiConverterUint64INSTANCE.Lower(amountSats), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue Txid
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeTxidINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (object *OnchainPayment) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterOnchainPayment struct{}

var FfiConverterOnchainPaymentINSTANCE = FfiConverterOnchainPayment{}

func (c FfiConverterOnchainPayment) Lift(pointer unsafe.Pointer) *OnchainPayment {
	result := &OnchainPayment{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_onchainpayment(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*OnchainPayment).Destroy)
	return result
}

func (c FfiConverterOnchainPayment) Read(reader io.Reader) *OnchainPayment {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterOnchainPayment) Lower(value *OnchainPayment) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*OnchainPayment")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterOnchainPayment) Write(writer io.Writer, value *OnchainPayment) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerOnchainPayment struct{}

func (_ FfiDestroyerOnchainPayment) Destroy(value *OnchainPayment) {
	value.Destroy()
}

type SpontaneousPayment struct {
	ffiObject FfiObject
}

func (_self *SpontaneousPayment) Send(amountMsat uint64, nodeId PublicKey, sendingParameters *SendingParameters, customTlvs []TlvEntry, preimage *PaymentPreimage) (PaymentId, error) {
	_pointer := _self.ffiObject.incrementPointer("*SpontaneousPayment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_spontaneouspayment_send(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), FfiConverterOptionalTypeSendingParametersINSTANCE.Lower(sendingParameters), FfiConverterSequenceTypeTlvEntryINSTANCE.Lower(customTlvs), FfiConverterOptionalTypePaymentPreimageINSTANCE.Lower(preimage), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue PaymentId
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypePaymentIdINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *SpontaneousPayment) SendProbes(amountMsat uint64, nodeId PublicKey) error {
	_pointer := _self.ffiObject.incrementPointer("*SpontaneousPayment")
	defer _self.ffiObject.decrementPointer()
	_, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) bool {
		C.uniffi_ldk_node_fn_method_spontaneouspayment_send_probes(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountMsat), FfiConverterTypePublicKeyINSTANCE.Lower(nodeId), _uniffiStatus)
		return false
	})
	return _uniffiErr
}

func (object *SpontaneousPayment) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterSpontaneousPayment struct{}

var FfiConverterSpontaneousPaymentINSTANCE = FfiConverterSpontaneousPayment{}

func (c FfiConverterSpontaneousPayment) Lift(pointer unsafe.Pointer) *SpontaneousPayment {
	result := &SpontaneousPayment{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_spontaneouspayment(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*SpontaneousPayment).Destroy)
	return result
}

func (c FfiConverterSpontaneousPayment) Read(reader io.Reader) *SpontaneousPayment {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterSpontaneousPayment) Lower(value *SpontaneousPayment) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*SpontaneousPayment")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterSpontaneousPayment) Write(writer io.Writer, value *SpontaneousPayment) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerSpontaneousPayment struct{}

func (_ FfiDestroyerSpontaneousPayment) Destroy(value *SpontaneousPayment) {
	value.Destroy()
}

type UnifiedQrPayment struct {
	ffiObject FfiObject
}

func (_self *UnifiedQrPayment) Receive(amountSats uint64, message string, expirySec uint32) (string, error) {
	_pointer := _self.ffiObject.incrementPointer("*UnifiedQrPayment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_unifiedqrpayment_receive(
			_pointer, FfiConverterUint64INSTANCE.Lower(amountSats), FfiConverterStringINSTANCE.Lower(message), FfiConverterUint32INSTANCE.Lower(expirySec), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue string
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterStringINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (_self *UnifiedQrPayment) Send(uriStr string) (QrPaymentResult, error) {
	_pointer := _self.ffiObject.incrementPointer("*UnifiedQrPayment")
	defer _self.ffiObject.decrementPointer()
	_uniffiRV, _uniffiErr := rustCallWithError(FfiConverterTypeNodeError{}, func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_method_unifiedqrpayment_send(
			_pointer, FfiConverterStringINSTANCE.Lower(uriStr), _uniffiStatus)
	})
	if _uniffiErr != nil {
		var _uniffiDefaultValue QrPaymentResult
		return _uniffiDefaultValue, _uniffiErr
	} else {
		return FfiConverterTypeQrPaymentResultINSTANCE.Lift(_uniffiRV), _uniffiErr
	}
}

func (object *UnifiedQrPayment) Destroy() {
	runtime.SetFinalizer(object, nil)
	object.ffiObject.destroy()
}

type FfiConverterUnifiedQrPayment struct{}

var FfiConverterUnifiedQrPaymentINSTANCE = FfiConverterUnifiedQrPayment{}

func (c FfiConverterUnifiedQrPayment) Lift(pointer unsafe.Pointer) *UnifiedQrPayment {
	result := &UnifiedQrPayment{
		newFfiObject(
			pointer,
			func(pointer unsafe.Pointer, status *C.RustCallStatus) {
				C.uniffi_ldk_node_fn_free_unifiedqrpayment(pointer, status)
			}),
	}
	runtime.SetFinalizer(result, (*UnifiedQrPayment).Destroy)
	return result
}

func (c FfiConverterUnifiedQrPayment) Read(reader io.Reader) *UnifiedQrPayment {
	return c.Lift(unsafe.Pointer(uintptr(readUint64(reader))))
}

func (c FfiConverterUnifiedQrPayment) Lower(value *UnifiedQrPayment) unsafe.Pointer {
	// TODO: this is bad - all synchronization from ObjectRuntime.go is discarded here,
	// because the pointer will be decremented immediately after this function returns,
	// and someone will be left holding onto a non-locked pointer.
	pointer := value.ffiObject.incrementPointer("*UnifiedQrPayment")
	defer value.ffiObject.decrementPointer()
	return pointer
}

func (c FfiConverterUnifiedQrPayment) Write(writer io.Writer, value *UnifiedQrPayment) {
	writeUint64(writer, uint64(uintptr(c.Lower(value))))
}

type FfiDestroyerUnifiedQrPayment struct{}

func (_ FfiDestroyerUnifiedQrPayment) Destroy(value *UnifiedQrPayment) {
	value.Destroy()
}

type AnchorChannelsConfig struct {
	TrustedPeersNoReserve []PublicKey
	PerChannelReserveSats uint64
}

func (r *AnchorChannelsConfig) Destroy() {
	FfiDestroyerSequenceTypePublicKey{}.Destroy(r.TrustedPeersNoReserve)
	FfiDestroyerUint64{}.Destroy(r.PerChannelReserveSats)
}

type FfiConverterTypeAnchorChannelsConfig struct{}

var FfiConverterTypeAnchorChannelsConfigINSTANCE = FfiConverterTypeAnchorChannelsConfig{}

func (c FfiConverterTypeAnchorChannelsConfig) Lift(rb RustBufferI) AnchorChannelsConfig {
	return LiftFromRustBuffer[AnchorChannelsConfig](c, rb)
}

func (c FfiConverterTypeAnchorChannelsConfig) Read(reader io.Reader) AnchorChannelsConfig {
	return AnchorChannelsConfig{
		FfiConverterSequenceTypePublicKeyINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeAnchorChannelsConfig) Lower(value AnchorChannelsConfig) RustBuffer {
	return LowerIntoRustBuffer[AnchorChannelsConfig](c, value)
}

func (c FfiConverterTypeAnchorChannelsConfig) Write(writer io.Writer, value AnchorChannelsConfig) {
	FfiConverterSequenceTypePublicKeyINSTANCE.Write(writer, value.TrustedPeersNoReserve)
	FfiConverterUint64INSTANCE.Write(writer, value.PerChannelReserveSats)
}

type FfiDestroyerTypeAnchorChannelsConfig struct{}

func (_ FfiDestroyerTypeAnchorChannelsConfig) Destroy(value AnchorChannelsConfig) {
	value.Destroy()
}

type BalanceDetails struct {
	TotalOnchainBalanceSats            uint64
	SpendableOnchainBalanceSats        uint64
	TotalAnchorChannelsReserveSats     uint64
	TotalLightningBalanceSats          uint64
	LightningBalances                  []LightningBalance
	PendingBalancesFromChannelClosures []PendingSweepBalance
}

func (r *BalanceDetails) Destroy() {
	FfiDestroyerUint64{}.Destroy(r.TotalOnchainBalanceSats)
	FfiDestroyerUint64{}.Destroy(r.SpendableOnchainBalanceSats)
	FfiDestroyerUint64{}.Destroy(r.TotalAnchorChannelsReserveSats)
	FfiDestroyerUint64{}.Destroy(r.TotalLightningBalanceSats)
	FfiDestroyerSequenceTypeLightningBalance{}.Destroy(r.LightningBalances)
	FfiDestroyerSequenceTypePendingSweepBalance{}.Destroy(r.PendingBalancesFromChannelClosures)
}

type FfiConverterTypeBalanceDetails struct{}

var FfiConverterTypeBalanceDetailsINSTANCE = FfiConverterTypeBalanceDetails{}

func (c FfiConverterTypeBalanceDetails) Lift(rb RustBufferI) BalanceDetails {
	return LiftFromRustBuffer[BalanceDetails](c, rb)
}

func (c FfiConverterTypeBalanceDetails) Read(reader io.Reader) BalanceDetails {
	return BalanceDetails{
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterSequenceTypeLightningBalanceINSTANCE.Read(reader),
		FfiConverterSequenceTypePendingSweepBalanceINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeBalanceDetails) Lower(value BalanceDetails) RustBuffer {
	return LowerIntoRustBuffer[BalanceDetails](c, value)
}

func (c FfiConverterTypeBalanceDetails) Write(writer io.Writer, value BalanceDetails) {
	FfiConverterUint64INSTANCE.Write(writer, value.TotalOnchainBalanceSats)
	FfiConverterUint64INSTANCE.Write(writer, value.SpendableOnchainBalanceSats)
	FfiConverterUint64INSTANCE.Write(writer, value.TotalAnchorChannelsReserveSats)
	FfiConverterUint64INSTANCE.Write(writer, value.TotalLightningBalanceSats)
	FfiConverterSequenceTypeLightningBalanceINSTANCE.Write(writer, value.LightningBalances)
	FfiConverterSequenceTypePendingSweepBalanceINSTANCE.Write(writer, value.PendingBalancesFromChannelClosures)
}

type FfiDestroyerTypeBalanceDetails struct{}

func (_ FfiDestroyerTypeBalanceDetails) Destroy(value BalanceDetails) {
	value.Destroy()
}

type BestBlock struct {
	BlockHash BlockHash
	Height    uint32
}

func (r *BestBlock) Destroy() {
	FfiDestroyerTypeBlockHash{}.Destroy(r.BlockHash)
	FfiDestroyerUint32{}.Destroy(r.Height)
}

type FfiConverterTypeBestBlock struct{}

var FfiConverterTypeBestBlockINSTANCE = FfiConverterTypeBestBlock{}

func (c FfiConverterTypeBestBlock) Lift(rb RustBufferI) BestBlock {
	return LiftFromRustBuffer[BestBlock](c, rb)
}

func (c FfiConverterTypeBestBlock) Read(reader io.Reader) BestBlock {
	return BestBlock{
		FfiConverterTypeBlockHashINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeBestBlock) Lower(value BestBlock) RustBuffer {
	return LowerIntoRustBuffer[BestBlock](c, value)
}

func (c FfiConverterTypeBestBlock) Write(writer io.Writer, value BestBlock) {
	FfiConverterTypeBlockHashINSTANCE.Write(writer, value.BlockHash)
	FfiConverterUint32INSTANCE.Write(writer, value.Height)
}

type FfiDestroyerTypeBestBlock struct{}

func (_ FfiDestroyerTypeBestBlock) Destroy(value BestBlock) {
	value.Destroy()
}

type ChannelConfig struct {
	ForwardingFeeProportionalMillionths uint32
	ForwardingFeeBaseMsat               uint32
	CltvExpiryDelta                     uint16
	MaxDustHtlcExposure                 MaxDustHtlcExposure
	ForceCloseAvoidanceMaxFeeSatoshis   uint64
	AcceptUnderpayingHtlcs              bool
}

func (r *ChannelConfig) Destroy() {
	FfiDestroyerUint32{}.Destroy(r.ForwardingFeeProportionalMillionths)
	FfiDestroyerUint32{}.Destroy(r.ForwardingFeeBaseMsat)
	FfiDestroyerUint16{}.Destroy(r.CltvExpiryDelta)
	FfiDestroyerTypeMaxDustHtlcExposure{}.Destroy(r.MaxDustHtlcExposure)
	FfiDestroyerUint64{}.Destroy(r.ForceCloseAvoidanceMaxFeeSatoshis)
	FfiDestroyerBool{}.Destroy(r.AcceptUnderpayingHtlcs)
}

type FfiConverterTypeChannelConfig struct{}

var FfiConverterTypeChannelConfigINSTANCE = FfiConverterTypeChannelConfig{}

func (c FfiConverterTypeChannelConfig) Lift(rb RustBufferI) ChannelConfig {
	return LiftFromRustBuffer[ChannelConfig](c, rb)
}

func (c FfiConverterTypeChannelConfig) Read(reader io.Reader) ChannelConfig {
	return ChannelConfig{
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint16INSTANCE.Read(reader),
		FfiConverterTypeMaxDustHTLCExposureINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeChannelConfig) Lower(value ChannelConfig) RustBuffer {
	return LowerIntoRustBuffer[ChannelConfig](c, value)
}

func (c FfiConverterTypeChannelConfig) Write(writer io.Writer, value ChannelConfig) {
	FfiConverterUint32INSTANCE.Write(writer, value.ForwardingFeeProportionalMillionths)
	FfiConverterUint32INSTANCE.Write(writer, value.ForwardingFeeBaseMsat)
	FfiConverterUint16INSTANCE.Write(writer, value.CltvExpiryDelta)
	FfiConverterTypeMaxDustHTLCExposureINSTANCE.Write(writer, value.MaxDustHtlcExposure)
	FfiConverterUint64INSTANCE.Write(writer, value.ForceCloseAvoidanceMaxFeeSatoshis)
	FfiConverterBoolINSTANCE.Write(writer, value.AcceptUnderpayingHtlcs)
}

type FfiDestroyerTypeChannelConfig struct{}

func (_ FfiDestroyerTypeChannelConfig) Destroy(value ChannelConfig) {
	value.Destroy()
}

type ChannelDetails struct {
	ChannelId                                           ChannelId
	CounterpartyNodeId                                  PublicKey
	FundingTxo                                          *OutPoint
	ChannelType                                         *ChannelType
	ChannelValueSats                                    uint64
	UnspendablePunishmentReserve                        *uint64
	UserChannelId                                       UserChannelId
	FeerateSatPer1000Weight                             uint32
	OutboundCapacityMsat                                uint64
	InboundCapacityMsat                                 uint64
	ConfirmationsRequired                               *uint32
	Confirmations                                       *uint32
	IsOutbound                                          bool
	IsChannelReady                                      bool
	IsUsable                                            bool
	IsAnnounced                                         bool
	CltvExpiryDelta                                     *uint16
	CounterpartyUnspendablePunishmentReserve            uint64
	CounterpartyOutboundHtlcMinimumMsat                 *uint64
	CounterpartyOutboundHtlcMaximumMsat                 *uint64
	CounterpartyForwardingInfoFeeBaseMsat               *uint32
	CounterpartyForwardingInfoFeeProportionalMillionths *uint32
	CounterpartyForwardingInfoCltvExpiryDelta           *uint16
	NextOutboundHtlcLimitMsat                           uint64
	NextOutboundHtlcMinimumMsat                         uint64
	ForceCloseSpendDelay                                *uint16
	InboundHtlcMinimumMsat                              uint64
	InboundHtlcMaximumMsat                              *uint64
	Config                                              ChannelConfig
}

func (r *ChannelDetails) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(r.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(r.CounterpartyNodeId)
	FfiDestroyerOptionalTypeOutPoint{}.Destroy(r.FundingTxo)
	FfiDestroyerOptionalTypeChannelType{}.Destroy(r.ChannelType)
	FfiDestroyerUint64{}.Destroy(r.ChannelValueSats)
	FfiDestroyerOptionalUint64{}.Destroy(r.UnspendablePunishmentReserve)
	FfiDestroyerTypeUserChannelId{}.Destroy(r.UserChannelId)
	FfiDestroyerUint32{}.Destroy(r.FeerateSatPer1000Weight)
	FfiDestroyerUint64{}.Destroy(r.OutboundCapacityMsat)
	FfiDestroyerUint64{}.Destroy(r.InboundCapacityMsat)
	FfiDestroyerOptionalUint32{}.Destroy(r.ConfirmationsRequired)
	FfiDestroyerOptionalUint32{}.Destroy(r.Confirmations)
	FfiDestroyerBool{}.Destroy(r.IsOutbound)
	FfiDestroyerBool{}.Destroy(r.IsChannelReady)
	FfiDestroyerBool{}.Destroy(r.IsUsable)
	FfiDestroyerBool{}.Destroy(r.IsAnnounced)
	FfiDestroyerOptionalUint16{}.Destroy(r.CltvExpiryDelta)
	FfiDestroyerUint64{}.Destroy(r.CounterpartyUnspendablePunishmentReserve)
	FfiDestroyerOptionalUint64{}.Destroy(r.CounterpartyOutboundHtlcMinimumMsat)
	FfiDestroyerOptionalUint64{}.Destroy(r.CounterpartyOutboundHtlcMaximumMsat)
	FfiDestroyerOptionalUint32{}.Destroy(r.CounterpartyForwardingInfoFeeBaseMsat)
	FfiDestroyerOptionalUint32{}.Destroy(r.CounterpartyForwardingInfoFeeProportionalMillionths)
	FfiDestroyerOptionalUint16{}.Destroy(r.CounterpartyForwardingInfoCltvExpiryDelta)
	FfiDestroyerUint64{}.Destroy(r.NextOutboundHtlcLimitMsat)
	FfiDestroyerUint64{}.Destroy(r.NextOutboundHtlcMinimumMsat)
	FfiDestroyerOptionalUint16{}.Destroy(r.ForceCloseSpendDelay)
	FfiDestroyerUint64{}.Destroy(r.InboundHtlcMinimumMsat)
	FfiDestroyerOptionalUint64{}.Destroy(r.InboundHtlcMaximumMsat)
	FfiDestroyerTypeChannelConfig{}.Destroy(r.Config)
}

type FfiConverterTypeChannelDetails struct{}

var FfiConverterTypeChannelDetailsINSTANCE = FfiConverterTypeChannelDetails{}

func (c FfiConverterTypeChannelDetails) Lift(rb RustBufferI) ChannelDetails {
	return LiftFromRustBuffer[ChannelDetails](c, rb)
}

func (c FfiConverterTypeChannelDetails) Read(reader io.Reader) ChannelDetails {
	return ChannelDetails{
		FfiConverterTypeChannelIdINSTANCE.Read(reader),
		FfiConverterTypePublicKeyINSTANCE.Read(reader),
		FfiConverterOptionalTypeOutPointINSTANCE.Read(reader),
		FfiConverterOptionalTypeChannelTypeINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterTypeUserChannelIdINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterOptionalUint16INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterOptionalUint16INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint16INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterTypeChannelConfigINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeChannelDetails) Lower(value ChannelDetails) RustBuffer {
	return LowerIntoRustBuffer[ChannelDetails](c, value)
}

func (c FfiConverterTypeChannelDetails) Write(writer io.Writer, value ChannelDetails) {
	FfiConverterTypeChannelIdINSTANCE.Write(writer, value.ChannelId)
	FfiConverterTypePublicKeyINSTANCE.Write(writer, value.CounterpartyNodeId)
	FfiConverterOptionalTypeOutPointINSTANCE.Write(writer, value.FundingTxo)
	FfiConverterOptionalTypeChannelTypeINSTANCE.Write(writer, value.ChannelType)
	FfiConverterUint64INSTANCE.Write(writer, value.ChannelValueSats)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.UnspendablePunishmentReserve)
	FfiConverterTypeUserChannelIdINSTANCE.Write(writer, value.UserChannelId)
	FfiConverterUint32INSTANCE.Write(writer, value.FeerateSatPer1000Weight)
	FfiConverterUint64INSTANCE.Write(writer, value.OutboundCapacityMsat)
	FfiConverterUint64INSTANCE.Write(writer, value.InboundCapacityMsat)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.ConfirmationsRequired)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.Confirmations)
	FfiConverterBoolINSTANCE.Write(writer, value.IsOutbound)
	FfiConverterBoolINSTANCE.Write(writer, value.IsChannelReady)
	FfiConverterBoolINSTANCE.Write(writer, value.IsUsable)
	FfiConverterBoolINSTANCE.Write(writer, value.IsAnnounced)
	FfiConverterOptionalUint16INSTANCE.Write(writer, value.CltvExpiryDelta)
	FfiConverterUint64INSTANCE.Write(writer, value.CounterpartyUnspendablePunishmentReserve)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.CounterpartyOutboundHtlcMinimumMsat)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.CounterpartyOutboundHtlcMaximumMsat)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.CounterpartyForwardingInfoFeeBaseMsat)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.CounterpartyForwardingInfoFeeProportionalMillionths)
	FfiConverterOptionalUint16INSTANCE.Write(writer, value.CounterpartyForwardingInfoCltvExpiryDelta)
	FfiConverterUint64INSTANCE.Write(writer, value.NextOutboundHtlcLimitMsat)
	FfiConverterUint64INSTANCE.Write(writer, value.NextOutboundHtlcMinimumMsat)
	FfiConverterOptionalUint16INSTANCE.Write(writer, value.ForceCloseSpendDelay)
	FfiConverterUint64INSTANCE.Write(writer, value.InboundHtlcMinimumMsat)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.InboundHtlcMaximumMsat)
	FfiConverterTypeChannelConfigINSTANCE.Write(writer, value.Config)
}

type FfiDestroyerTypeChannelDetails struct{}

func (_ FfiDestroyerTypeChannelDetails) Destroy(value ChannelDetails) {
	value.Destroy()
}

type ChannelInfo struct {
	NodeOne      NodeId
	OneToTwo     *ChannelUpdateInfo
	NodeTwo      NodeId
	TwoToOne     *ChannelUpdateInfo
	CapacitySats *uint64
}

func (r *ChannelInfo) Destroy() {
	FfiDestroyerTypeNodeId{}.Destroy(r.NodeOne)
	FfiDestroyerOptionalTypeChannelUpdateInfo{}.Destroy(r.OneToTwo)
	FfiDestroyerTypeNodeId{}.Destroy(r.NodeTwo)
	FfiDestroyerOptionalTypeChannelUpdateInfo{}.Destroy(r.TwoToOne)
	FfiDestroyerOptionalUint64{}.Destroy(r.CapacitySats)
}

type FfiConverterTypeChannelInfo struct{}

var FfiConverterTypeChannelInfoINSTANCE = FfiConverterTypeChannelInfo{}

func (c FfiConverterTypeChannelInfo) Lift(rb RustBufferI) ChannelInfo {
	return LiftFromRustBuffer[ChannelInfo](c, rb)
}

func (c FfiConverterTypeChannelInfo) Read(reader io.Reader) ChannelInfo {
	return ChannelInfo{
		FfiConverterTypeNodeIdINSTANCE.Read(reader),
		FfiConverterOptionalTypeChannelUpdateInfoINSTANCE.Read(reader),
		FfiConverterTypeNodeIdINSTANCE.Read(reader),
		FfiConverterOptionalTypeChannelUpdateInfoINSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeChannelInfo) Lower(value ChannelInfo) RustBuffer {
	return LowerIntoRustBuffer[ChannelInfo](c, value)
}

func (c FfiConverterTypeChannelInfo) Write(writer io.Writer, value ChannelInfo) {
	FfiConverterTypeNodeIdINSTANCE.Write(writer, value.NodeOne)
	FfiConverterOptionalTypeChannelUpdateInfoINSTANCE.Write(writer, value.OneToTwo)
	FfiConverterTypeNodeIdINSTANCE.Write(writer, value.NodeTwo)
	FfiConverterOptionalTypeChannelUpdateInfoINSTANCE.Write(writer, value.TwoToOne)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.CapacitySats)
}

type FfiDestroyerTypeChannelInfo struct{}

func (_ FfiDestroyerTypeChannelInfo) Destroy(value ChannelInfo) {
	value.Destroy()
}

type ChannelUpdateInfo struct {
	LastUpdate      uint32
	Enabled         bool
	CltvExpiryDelta uint16
	HtlcMinimumMsat uint64
	HtlcMaximumMsat uint64
	Fees            RoutingFees
}

func (r *ChannelUpdateInfo) Destroy() {
	FfiDestroyerUint32{}.Destroy(r.LastUpdate)
	FfiDestroyerBool{}.Destroy(r.Enabled)
	FfiDestroyerUint16{}.Destroy(r.CltvExpiryDelta)
	FfiDestroyerUint64{}.Destroy(r.HtlcMinimumMsat)
	FfiDestroyerUint64{}.Destroy(r.HtlcMaximumMsat)
	FfiDestroyerTypeRoutingFees{}.Destroy(r.Fees)
}

type FfiConverterTypeChannelUpdateInfo struct{}

var FfiConverterTypeChannelUpdateInfoINSTANCE = FfiConverterTypeChannelUpdateInfo{}

func (c FfiConverterTypeChannelUpdateInfo) Lift(rb RustBufferI) ChannelUpdateInfo {
	return LiftFromRustBuffer[ChannelUpdateInfo](c, rb)
}

func (c FfiConverterTypeChannelUpdateInfo) Read(reader io.Reader) ChannelUpdateInfo {
	return ChannelUpdateInfo{
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterUint16INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterTypeRoutingFeesINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeChannelUpdateInfo) Lower(value ChannelUpdateInfo) RustBuffer {
	return LowerIntoRustBuffer[ChannelUpdateInfo](c, value)
}

func (c FfiConverterTypeChannelUpdateInfo) Write(writer io.Writer, value ChannelUpdateInfo) {
	FfiConverterUint32INSTANCE.Write(writer, value.LastUpdate)
	FfiConverterBoolINSTANCE.Write(writer, value.Enabled)
	FfiConverterUint16INSTANCE.Write(writer, value.CltvExpiryDelta)
	FfiConverterUint64INSTANCE.Write(writer, value.HtlcMinimumMsat)
	FfiConverterUint64INSTANCE.Write(writer, value.HtlcMaximumMsat)
	FfiConverterTypeRoutingFeesINSTANCE.Write(writer, value.Fees)
}

type FfiDestroyerTypeChannelUpdateInfo struct{}

func (_ FfiDestroyerTypeChannelUpdateInfo) Destroy(value ChannelUpdateInfo) {
	value.Destroy()
}

type Config struct {
	StorageDirPath                  string
	LogDirPath                      *string
	Network                         Network
	ListeningAddresses              *[]SocketAddress
	NodeAlias                       *NodeAlias
	TrustedPeers0conf               []PublicKey
	ProbingLiquidityLimitMultiplier uint64
	LogLevel                        LogLevel
	AnchorChannelsConfig            *AnchorChannelsConfig
	SendingParameters               *SendingParameters
	TransientNetworkGraph           bool
}

func (r *Config) Destroy() {
	FfiDestroyerString{}.Destroy(r.StorageDirPath)
	FfiDestroyerOptionalString{}.Destroy(r.LogDirPath)
	FfiDestroyerTypeNetwork{}.Destroy(r.Network)
	FfiDestroyerOptionalSequenceTypeSocketAddress{}.Destroy(r.ListeningAddresses)
	FfiDestroyerOptionalTypeNodeAlias{}.Destroy(r.NodeAlias)
	FfiDestroyerSequenceTypePublicKey{}.Destroy(r.TrustedPeers0conf)
	FfiDestroyerUint64{}.Destroy(r.ProbingLiquidityLimitMultiplier)
	FfiDestroyerTypeLogLevel{}.Destroy(r.LogLevel)
	FfiDestroyerOptionalTypeAnchorChannelsConfig{}.Destroy(r.AnchorChannelsConfig)
	FfiDestroyerOptionalTypeSendingParameters{}.Destroy(r.SendingParameters)
	FfiDestroyerBool{}.Destroy(r.TransientNetworkGraph)
}

type FfiConverterTypeConfig struct{}

var FfiConverterTypeConfigINSTANCE = FfiConverterTypeConfig{}

func (c FfiConverterTypeConfig) Lift(rb RustBufferI) Config {
	return LiftFromRustBuffer[Config](c, rb)
}

func (c FfiConverterTypeConfig) Read(reader io.Reader) Config {
	return Config{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterTypeNetworkINSTANCE.Read(reader),
		FfiConverterOptionalSequenceTypeSocketAddressINSTANCE.Read(reader),
		FfiConverterOptionalTypeNodeAliasINSTANCE.Read(reader),
		FfiConverterSequenceTypePublicKeyINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterTypeLogLevelINSTANCE.Read(reader),
		FfiConverterOptionalTypeAnchorChannelsConfigINSTANCE.Read(reader),
		FfiConverterOptionalTypeSendingParametersINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeConfig) Lower(value Config) RustBuffer {
	return LowerIntoRustBuffer[Config](c, value)
}

func (c FfiConverterTypeConfig) Write(writer io.Writer, value Config) {
	FfiConverterStringINSTANCE.Write(writer, value.StorageDirPath)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.LogDirPath)
	FfiConverterTypeNetworkINSTANCE.Write(writer, value.Network)
	FfiConverterOptionalSequenceTypeSocketAddressINSTANCE.Write(writer, value.ListeningAddresses)
	FfiConverterOptionalTypeNodeAliasINSTANCE.Write(writer, value.NodeAlias)
	FfiConverterSequenceTypePublicKeyINSTANCE.Write(writer, value.TrustedPeers0conf)
	FfiConverterUint64INSTANCE.Write(writer, value.ProbingLiquidityLimitMultiplier)
	FfiConverterTypeLogLevelINSTANCE.Write(writer, value.LogLevel)
	FfiConverterOptionalTypeAnchorChannelsConfigINSTANCE.Write(writer, value.AnchorChannelsConfig)
	FfiConverterOptionalTypeSendingParametersINSTANCE.Write(writer, value.SendingParameters)
	FfiConverterBoolINSTANCE.Write(writer, value.TransientNetworkGraph)
}

type FfiDestroyerTypeConfig struct{}

func (_ FfiDestroyerTypeConfig) Destroy(value Config) {
	value.Destroy()
}

type EsploraSyncConfig struct {
	OnchainWalletSyncIntervalSecs   uint64
	LightningWalletSyncIntervalSecs uint64
	FeeRateCacheUpdateIntervalSecs  uint64
}

func (r *EsploraSyncConfig) Destroy() {
	FfiDestroyerUint64{}.Destroy(r.OnchainWalletSyncIntervalSecs)
	FfiDestroyerUint64{}.Destroy(r.LightningWalletSyncIntervalSecs)
	FfiDestroyerUint64{}.Destroy(r.FeeRateCacheUpdateIntervalSecs)
}

type FfiConverterTypeEsploraSyncConfig struct{}

var FfiConverterTypeEsploraSyncConfigINSTANCE = FfiConverterTypeEsploraSyncConfig{}

func (c FfiConverterTypeEsploraSyncConfig) Lift(rb RustBufferI) EsploraSyncConfig {
	return LiftFromRustBuffer[EsploraSyncConfig](c, rb)
}

func (c FfiConverterTypeEsploraSyncConfig) Read(reader io.Reader) EsploraSyncConfig {
	return EsploraSyncConfig{
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeEsploraSyncConfig) Lower(value EsploraSyncConfig) RustBuffer {
	return LowerIntoRustBuffer[EsploraSyncConfig](c, value)
}

func (c FfiConverterTypeEsploraSyncConfig) Write(writer io.Writer, value EsploraSyncConfig) {
	FfiConverterUint64INSTANCE.Write(writer, value.OnchainWalletSyncIntervalSecs)
	FfiConverterUint64INSTANCE.Write(writer, value.LightningWalletSyncIntervalSecs)
	FfiConverterUint64INSTANCE.Write(writer, value.FeeRateCacheUpdateIntervalSecs)
}

type FfiDestroyerTypeEsploraSyncConfig struct{}

func (_ FfiDestroyerTypeEsploraSyncConfig) Destroy(value EsploraSyncConfig) {
	value.Destroy()
}

type KeyValue struct {
	Key   string
	Value []uint8
}

func (r *KeyValue) Destroy() {
	FfiDestroyerString{}.Destroy(r.Key)
	FfiDestroyerSequenceUint8{}.Destroy(r.Value)
}

type FfiConverterTypeKeyValue struct{}

var FfiConverterTypeKeyValueINSTANCE = FfiConverterTypeKeyValue{}

func (c FfiConverterTypeKeyValue) Lift(rb RustBufferI) KeyValue {
	return LiftFromRustBuffer[KeyValue](c, rb)
}

func (c FfiConverterTypeKeyValue) Read(reader io.Reader) KeyValue {
	return KeyValue{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterSequenceUint8INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeKeyValue) Lower(value KeyValue) RustBuffer {
	return LowerIntoRustBuffer[KeyValue](c, value)
}

func (c FfiConverterTypeKeyValue) Write(writer io.Writer, value KeyValue) {
	FfiConverterStringINSTANCE.Write(writer, value.Key)
	FfiConverterSequenceUint8INSTANCE.Write(writer, value.Value)
}

type FfiDestroyerTypeKeyValue struct{}

func (_ FfiDestroyerTypeKeyValue) Destroy(value KeyValue) {
	value.Destroy()
}

type LspFeeLimits struct {
	MaxTotalOpeningFeeMsat           *uint64
	MaxProportionalOpeningFeePpmMsat *uint64
}

func (r *LspFeeLimits) Destroy() {
	FfiDestroyerOptionalUint64{}.Destroy(r.MaxTotalOpeningFeeMsat)
	FfiDestroyerOptionalUint64{}.Destroy(r.MaxProportionalOpeningFeePpmMsat)
}

type FfiConverterTypeLSPFeeLimits struct{}

var FfiConverterTypeLSPFeeLimitsINSTANCE = FfiConverterTypeLSPFeeLimits{}

func (c FfiConverterTypeLSPFeeLimits) Lift(rb RustBufferI) LspFeeLimits {
	return LiftFromRustBuffer[LspFeeLimits](c, rb)
}

func (c FfiConverterTypeLSPFeeLimits) Read(reader io.Reader) LspFeeLimits {
	return LspFeeLimits{
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeLSPFeeLimits) Lower(value LspFeeLimits) RustBuffer {
	return LowerIntoRustBuffer[LspFeeLimits](c, value)
}

func (c FfiConverterTypeLSPFeeLimits) Write(writer io.Writer, value LspFeeLimits) {
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.MaxTotalOpeningFeeMsat)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.MaxProportionalOpeningFeePpmMsat)
}

type FfiDestroyerTypeLspFeeLimits struct{}

func (_ FfiDestroyerTypeLspFeeLimits) Destroy(value LspFeeLimits) {
	value.Destroy()
}

type NodeAnnouncementInfo struct {
	LastUpdate uint32
	Alias      string
	Addresses  []SocketAddress
}

func (r *NodeAnnouncementInfo) Destroy() {
	FfiDestroyerUint32{}.Destroy(r.LastUpdate)
	FfiDestroyerString{}.Destroy(r.Alias)
	FfiDestroyerSequenceTypeSocketAddress{}.Destroy(r.Addresses)
}

type FfiConverterTypeNodeAnnouncementInfo struct{}

var FfiConverterTypeNodeAnnouncementInfoINSTANCE = FfiConverterTypeNodeAnnouncementInfo{}

func (c FfiConverterTypeNodeAnnouncementInfo) Lift(rb RustBufferI) NodeAnnouncementInfo {
	return LiftFromRustBuffer[NodeAnnouncementInfo](c, rb)
}

func (c FfiConverterTypeNodeAnnouncementInfo) Read(reader io.Reader) NodeAnnouncementInfo {
	return NodeAnnouncementInfo{
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterSequenceTypeSocketAddressINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeNodeAnnouncementInfo) Lower(value NodeAnnouncementInfo) RustBuffer {
	return LowerIntoRustBuffer[NodeAnnouncementInfo](c, value)
}

func (c FfiConverterTypeNodeAnnouncementInfo) Write(writer io.Writer, value NodeAnnouncementInfo) {
	FfiConverterUint32INSTANCE.Write(writer, value.LastUpdate)
	FfiConverterStringINSTANCE.Write(writer, value.Alias)
	FfiConverterSequenceTypeSocketAddressINSTANCE.Write(writer, value.Addresses)
}

type FfiDestroyerTypeNodeAnnouncementInfo struct{}

func (_ FfiDestroyerTypeNodeAnnouncementInfo) Destroy(value NodeAnnouncementInfo) {
	value.Destroy()
}

type NodeInfo struct {
	Channels         []uint64
	AnnouncementInfo *NodeAnnouncementInfo
}

func (r *NodeInfo) Destroy() {
	FfiDestroyerSequenceUint64{}.Destroy(r.Channels)
	FfiDestroyerOptionalTypeNodeAnnouncementInfo{}.Destroy(r.AnnouncementInfo)
}

type FfiConverterTypeNodeInfo struct{}

var FfiConverterTypeNodeInfoINSTANCE = FfiConverterTypeNodeInfo{}

func (c FfiConverterTypeNodeInfo) Lift(rb RustBufferI) NodeInfo {
	return LiftFromRustBuffer[NodeInfo](c, rb)
}

func (c FfiConverterTypeNodeInfo) Read(reader io.Reader) NodeInfo {
	return NodeInfo{
		FfiConverterSequenceUint64INSTANCE.Read(reader),
		FfiConverterOptionalTypeNodeAnnouncementInfoINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeNodeInfo) Lower(value NodeInfo) RustBuffer {
	return LowerIntoRustBuffer[NodeInfo](c, value)
}

func (c FfiConverterTypeNodeInfo) Write(writer io.Writer, value NodeInfo) {
	FfiConverterSequenceUint64INSTANCE.Write(writer, value.Channels)
	FfiConverterOptionalTypeNodeAnnouncementInfoINSTANCE.Write(writer, value.AnnouncementInfo)
}

type FfiDestroyerTypeNodeInfo struct{}

func (_ FfiDestroyerTypeNodeInfo) Destroy(value NodeInfo) {
	value.Destroy()
}

type NodeStatus struct {
	IsRunning                                bool
	IsListening                              bool
	CurrentBestBlock                         BestBlock
	LatestLightningWalletSyncTimestamp       *uint64
	LatestOnchainWalletSyncTimestamp         *uint64
	LatestFeeRateCacheUpdateTimestamp        *uint64
	LatestRgsSnapshotTimestamp               *uint64
	LatestNodeAnnouncementBroadcastTimestamp *uint64
	LatestChannelMonitorArchivalHeight       *uint32
}

func (r *NodeStatus) Destroy() {
	FfiDestroyerBool{}.Destroy(r.IsRunning)
	FfiDestroyerBool{}.Destroy(r.IsListening)
	FfiDestroyerTypeBestBlock{}.Destroy(r.CurrentBestBlock)
	FfiDestroyerOptionalUint64{}.Destroy(r.LatestLightningWalletSyncTimestamp)
	FfiDestroyerOptionalUint64{}.Destroy(r.LatestOnchainWalletSyncTimestamp)
	FfiDestroyerOptionalUint64{}.Destroy(r.LatestFeeRateCacheUpdateTimestamp)
	FfiDestroyerOptionalUint64{}.Destroy(r.LatestRgsSnapshotTimestamp)
	FfiDestroyerOptionalUint64{}.Destroy(r.LatestNodeAnnouncementBroadcastTimestamp)
	FfiDestroyerOptionalUint32{}.Destroy(r.LatestChannelMonitorArchivalHeight)
}

type FfiConverterTypeNodeStatus struct{}

var FfiConverterTypeNodeStatusINSTANCE = FfiConverterTypeNodeStatus{}

func (c FfiConverterTypeNodeStatus) Lift(rb RustBufferI) NodeStatus {
	return LiftFromRustBuffer[NodeStatus](c, rb)
}

func (c FfiConverterTypeNodeStatus) Read(reader io.Reader) NodeStatus {
	return NodeStatus{
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterTypeBestBlockINSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeNodeStatus) Lower(value NodeStatus) RustBuffer {
	return LowerIntoRustBuffer[NodeStatus](c, value)
}

func (c FfiConverterTypeNodeStatus) Write(writer io.Writer, value NodeStatus) {
	FfiConverterBoolINSTANCE.Write(writer, value.IsRunning)
	FfiConverterBoolINSTANCE.Write(writer, value.IsListening)
	FfiConverterTypeBestBlockINSTANCE.Write(writer, value.CurrentBestBlock)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LatestLightningWalletSyncTimestamp)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LatestOnchainWalletSyncTimestamp)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LatestFeeRateCacheUpdateTimestamp)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LatestRgsSnapshotTimestamp)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.LatestNodeAnnouncementBroadcastTimestamp)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.LatestChannelMonitorArchivalHeight)
}

type FfiDestroyerTypeNodeStatus struct{}

func (_ FfiDestroyerTypeNodeStatus) Destroy(value NodeStatus) {
	value.Destroy()
}

type OutPoint struct {
	Txid Txid
	Vout uint32
}

func (r *OutPoint) Destroy() {
	FfiDestroyerTypeTxid{}.Destroy(r.Txid)
	FfiDestroyerUint32{}.Destroy(r.Vout)
}

type FfiConverterTypeOutPoint struct{}

var FfiConverterTypeOutPointINSTANCE = FfiConverterTypeOutPoint{}

func (c FfiConverterTypeOutPoint) Lift(rb RustBufferI) OutPoint {
	return LiftFromRustBuffer[OutPoint](c, rb)
}

func (c FfiConverterTypeOutPoint) Read(reader io.Reader) OutPoint {
	return OutPoint{
		FfiConverterTypeTxidINSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeOutPoint) Lower(value OutPoint) RustBuffer {
	return LowerIntoRustBuffer[OutPoint](c, value)
}

func (c FfiConverterTypeOutPoint) Write(writer io.Writer, value OutPoint) {
	FfiConverterTypeTxidINSTANCE.Write(writer, value.Txid)
	FfiConverterUint32INSTANCE.Write(writer, value.Vout)
}

type FfiDestroyerTypeOutPoint struct{}

func (_ FfiDestroyerTypeOutPoint) Destroy(value OutPoint) {
	value.Destroy()
}

type PaymentDetails struct {
	Id                    PaymentId
	Kind                  PaymentKind
	AmountMsat            *uint64
	Direction             PaymentDirection
	Status                PaymentStatus
	LastUpdate            uint64
	FeeMsat               *uint64
	CreatedAt             uint64
	LatestUpdateTimestamp uint64
}

func (r *PaymentDetails) Destroy() {
	FfiDestroyerTypePaymentId{}.Destroy(r.Id)
	FfiDestroyerTypePaymentKind{}.Destroy(r.Kind)
	FfiDestroyerOptionalUint64{}.Destroy(r.AmountMsat)
	FfiDestroyerTypePaymentDirection{}.Destroy(r.Direction)
	FfiDestroyerTypePaymentStatus{}.Destroy(r.Status)
	FfiDestroyerUint64{}.Destroy(r.LastUpdate)
	FfiDestroyerOptionalUint64{}.Destroy(r.FeeMsat)
	FfiDestroyerUint64{}.Destroy(r.CreatedAt)
	FfiDestroyerUint64{}.Destroy(r.LatestUpdateTimestamp)
}

type FfiConverterTypePaymentDetails struct{}

var FfiConverterTypePaymentDetailsINSTANCE = FfiConverterTypePaymentDetails{}

func (c FfiConverterTypePaymentDetails) Lift(rb RustBufferI) PaymentDetails {
	return LiftFromRustBuffer[PaymentDetails](c, rb)
}

func (c FfiConverterTypePaymentDetails) Read(reader io.Reader) PaymentDetails {
	return PaymentDetails{
		FfiConverterTypePaymentIdINSTANCE.Read(reader),
		FfiConverterTypePaymentKindINSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterTypePaymentDirectionINSTANCE.Read(reader),
		FfiConverterTypePaymentStatusINSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterOptionalUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterUint64INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePaymentDetails) Lower(value PaymentDetails) RustBuffer {
	return LowerIntoRustBuffer[PaymentDetails](c, value)
}

func (c FfiConverterTypePaymentDetails) Write(writer io.Writer, value PaymentDetails) {
	FfiConverterTypePaymentIdINSTANCE.Write(writer, value.Id)
	FfiConverterTypePaymentKindINSTANCE.Write(writer, value.Kind)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.AmountMsat)
	FfiConverterTypePaymentDirectionINSTANCE.Write(writer, value.Direction)
	FfiConverterTypePaymentStatusINSTANCE.Write(writer, value.Status)
	FfiConverterUint64INSTANCE.Write(writer, value.LastUpdate)
	FfiConverterOptionalUint64INSTANCE.Write(writer, value.FeeMsat)
	FfiConverterUint64INSTANCE.Write(writer, value.CreatedAt)
	FfiConverterUint64INSTANCE.Write(writer, value.LatestUpdateTimestamp)
}

type FfiDestroyerTypePaymentDetails struct{}

func (_ FfiDestroyerTypePaymentDetails) Destroy(value PaymentDetails) {
	value.Destroy()
}

type PeerDetails struct {
	NodeId      PublicKey
	Address     SocketAddress
	IsPersisted bool
	IsConnected bool
}

func (r *PeerDetails) Destroy() {
	FfiDestroyerTypePublicKey{}.Destroy(r.NodeId)
	FfiDestroyerTypeSocketAddress{}.Destroy(r.Address)
	FfiDestroyerBool{}.Destroy(r.IsPersisted)
	FfiDestroyerBool{}.Destroy(r.IsConnected)
}

type FfiConverterTypePeerDetails struct{}

var FfiConverterTypePeerDetailsINSTANCE = FfiConverterTypePeerDetails{}

func (c FfiConverterTypePeerDetails) Lift(rb RustBufferI) PeerDetails {
	return LiftFromRustBuffer[PeerDetails](c, rb)
}

func (c FfiConverterTypePeerDetails) Read(reader io.Reader) PeerDetails {
	return PeerDetails{
		FfiConverterTypePublicKeyINSTANCE.Read(reader),
		FfiConverterTypeSocketAddressINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
		FfiConverterBoolINSTANCE.Read(reader),
	}
}

func (c FfiConverterTypePeerDetails) Lower(value PeerDetails) RustBuffer {
	return LowerIntoRustBuffer[PeerDetails](c, value)
}

func (c FfiConverterTypePeerDetails) Write(writer io.Writer, value PeerDetails) {
	FfiConverterTypePublicKeyINSTANCE.Write(writer, value.NodeId)
	FfiConverterTypeSocketAddressINSTANCE.Write(writer, value.Address)
	FfiConverterBoolINSTANCE.Write(writer, value.IsPersisted)
	FfiConverterBoolINSTANCE.Write(writer, value.IsConnected)
}

type FfiDestroyerTypePeerDetails struct{}

func (_ FfiDestroyerTypePeerDetails) Destroy(value PeerDetails) {
	value.Destroy()
}

type RoutingFees struct {
	BaseMsat               uint32
	ProportionalMillionths uint32
}

func (r *RoutingFees) Destroy() {
	FfiDestroyerUint32{}.Destroy(r.BaseMsat)
	FfiDestroyerUint32{}.Destroy(r.ProportionalMillionths)
}

type FfiConverterTypeRoutingFees struct{}

var FfiConverterTypeRoutingFeesINSTANCE = FfiConverterTypeRoutingFees{}

func (c FfiConverterTypeRoutingFees) Lift(rb RustBufferI) RoutingFees {
	return LiftFromRustBuffer[RoutingFees](c, rb)
}

func (c FfiConverterTypeRoutingFees) Read(reader io.Reader) RoutingFees {
	return RoutingFees{
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeRoutingFees) Lower(value RoutingFees) RustBuffer {
	return LowerIntoRustBuffer[RoutingFees](c, value)
}

func (c FfiConverterTypeRoutingFees) Write(writer io.Writer, value RoutingFees) {
	FfiConverterUint32INSTANCE.Write(writer, value.BaseMsat)
	FfiConverterUint32INSTANCE.Write(writer, value.ProportionalMillionths)
}

type FfiDestroyerTypeRoutingFees struct{}

func (_ FfiDestroyerTypeRoutingFees) Destroy(value RoutingFees) {
	value.Destroy()
}

type SendingParameters struct {
	MaxTotalRoutingFeeMsat          *MaxTotalRoutingFeeLimit
	MaxTotalCltvExpiryDelta         *uint32
	MaxPathCount                    *uint8
	MaxChannelSaturationPowerOfHalf *uint8
}

func (r *SendingParameters) Destroy() {
	FfiDestroyerOptionalTypeMaxTotalRoutingFeeLimit{}.Destroy(r.MaxTotalRoutingFeeMsat)
	FfiDestroyerOptionalUint32{}.Destroy(r.MaxTotalCltvExpiryDelta)
	FfiDestroyerOptionalUint8{}.Destroy(r.MaxPathCount)
	FfiDestroyerOptionalUint8{}.Destroy(r.MaxChannelSaturationPowerOfHalf)
}

type FfiConverterTypeSendingParameters struct{}

var FfiConverterTypeSendingParametersINSTANCE = FfiConverterTypeSendingParameters{}

func (c FfiConverterTypeSendingParameters) Lift(rb RustBufferI) SendingParameters {
	return LiftFromRustBuffer[SendingParameters](c, rb)
}

func (c FfiConverterTypeSendingParameters) Read(reader io.Reader) SendingParameters {
	return SendingParameters{
		FfiConverterOptionalTypeMaxTotalRoutingFeeLimitINSTANCE.Read(reader),
		FfiConverterOptionalUint32INSTANCE.Read(reader),
		FfiConverterOptionalUint8INSTANCE.Read(reader),
		FfiConverterOptionalUint8INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeSendingParameters) Lower(value SendingParameters) RustBuffer {
	return LowerIntoRustBuffer[SendingParameters](c, value)
}

func (c FfiConverterTypeSendingParameters) Write(writer io.Writer, value SendingParameters) {
	FfiConverterOptionalTypeMaxTotalRoutingFeeLimitINSTANCE.Write(writer, value.MaxTotalRoutingFeeMsat)
	FfiConverterOptionalUint32INSTANCE.Write(writer, value.MaxTotalCltvExpiryDelta)
	FfiConverterOptionalUint8INSTANCE.Write(writer, value.MaxPathCount)
	FfiConverterOptionalUint8INSTANCE.Write(writer, value.MaxChannelSaturationPowerOfHalf)
}

type FfiDestroyerTypeSendingParameters struct{}

func (_ FfiDestroyerTypeSendingParameters) Destroy(value SendingParameters) {
	value.Destroy()
}

type TlvEntry struct {
	Type  uint64
	Value []uint8
}

func (r *TlvEntry) Destroy() {
	FfiDestroyerUint64{}.Destroy(r.Type)
	FfiDestroyerSequenceUint8{}.Destroy(r.Value)
}

type FfiConverterTypeTlvEntry struct{}

var FfiConverterTypeTlvEntryINSTANCE = FfiConverterTypeTlvEntry{}

func (c FfiConverterTypeTlvEntry) Lift(rb RustBufferI) TlvEntry {
	return LiftFromRustBuffer[TlvEntry](c, rb)
}

func (c FfiConverterTypeTlvEntry) Read(reader io.Reader) TlvEntry {
	return TlvEntry{
		FfiConverterUint64INSTANCE.Read(reader),
		FfiConverterSequenceUint8INSTANCE.Read(reader),
	}
}

func (c FfiConverterTypeTlvEntry) Lower(value TlvEntry) RustBuffer {
	return LowerIntoRustBuffer[TlvEntry](c, value)
}

func (c FfiConverterTypeTlvEntry) Write(writer io.Writer, value TlvEntry) {
	FfiConverterUint64INSTANCE.Write(writer, value.Type)
	FfiConverterSequenceUint8INSTANCE.Write(writer, value.Value)
}

type FfiDestroyerTypeTlvEntry struct{}

func (_ FfiDestroyerTypeTlvEntry) Destroy(value TlvEntry) {
	value.Destroy()
}

type BalanceSource uint

const (
	BalanceSourceHolderForceClosed       BalanceSource = 1
	BalanceSourceCounterpartyForceClosed BalanceSource = 2
	BalanceSourceCoopClose               BalanceSource = 3
	BalanceSourceHtlc                    BalanceSource = 4
)

type FfiConverterTypeBalanceSource struct{}

var FfiConverterTypeBalanceSourceINSTANCE = FfiConverterTypeBalanceSource{}

func (c FfiConverterTypeBalanceSource) Lift(rb RustBufferI) BalanceSource {
	return LiftFromRustBuffer[BalanceSource](c, rb)
}

func (c FfiConverterTypeBalanceSource) Lower(value BalanceSource) RustBuffer {
	return LowerIntoRustBuffer[BalanceSource](c, value)
}
func (FfiConverterTypeBalanceSource) Read(reader io.Reader) BalanceSource {
	id := readInt32(reader)
	return BalanceSource(id)
}

func (FfiConverterTypeBalanceSource) Write(writer io.Writer, value BalanceSource) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeBalanceSource struct{}

func (_ FfiDestroyerTypeBalanceSource) Destroy(value BalanceSource) {
}

type BuildError struct {
	err error
}

func (err BuildError) Error() string {
	return fmt.Sprintf("BuildError: %s", err.err.Error())
}

func (err BuildError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrBuildErrorInvalidSeedBytes = fmt.Errorf("BuildErrorInvalidSeedBytes")
var ErrBuildErrorInvalidSeedFile = fmt.Errorf("BuildErrorInvalidSeedFile")
var ErrBuildErrorInvalidSystemTime = fmt.Errorf("BuildErrorInvalidSystemTime")
var ErrBuildErrorInvalidChannelMonitor = fmt.Errorf("BuildErrorInvalidChannelMonitor")
var ErrBuildErrorInvalidListeningAddresses = fmt.Errorf("BuildErrorInvalidListeningAddresses")
var ErrBuildErrorInvalidNodeAlias = fmt.Errorf("BuildErrorInvalidNodeAlias")
var ErrBuildErrorReadFailed = fmt.Errorf("BuildErrorReadFailed")
var ErrBuildErrorWriteFailed = fmt.Errorf("BuildErrorWriteFailed")
var ErrBuildErrorStoragePathAccessFailed = fmt.Errorf("BuildErrorStoragePathAccessFailed")
var ErrBuildErrorKvStoreSetupFailed = fmt.Errorf("BuildErrorKvStoreSetupFailed")
var ErrBuildErrorWalletSetupFailed = fmt.Errorf("BuildErrorWalletSetupFailed")
var ErrBuildErrorLoggerSetupFailed = fmt.Errorf("BuildErrorLoggerSetupFailed")

// Variant structs
type BuildErrorInvalidSeedBytes struct {
	message string
}

func NewBuildErrorInvalidSeedBytes() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidSeedBytes{},
	}
}

func (err BuildErrorInvalidSeedBytes) Error() string {
	return fmt.Sprintf("InvalidSeedBytes: %s", err.message)
}

func (self BuildErrorInvalidSeedBytes) Is(target error) bool {
	return target == ErrBuildErrorInvalidSeedBytes
}

type BuildErrorInvalidSeedFile struct {
	message string
}

func NewBuildErrorInvalidSeedFile() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidSeedFile{},
	}
}

func (err BuildErrorInvalidSeedFile) Error() string {
	return fmt.Sprintf("InvalidSeedFile: %s", err.message)
}

func (self BuildErrorInvalidSeedFile) Is(target error) bool {
	return target == ErrBuildErrorInvalidSeedFile
}

type BuildErrorInvalidSystemTime struct {
	message string
}

func NewBuildErrorInvalidSystemTime() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidSystemTime{},
	}
}

func (err BuildErrorInvalidSystemTime) Error() string {
	return fmt.Sprintf("InvalidSystemTime: %s", err.message)
}

func (self BuildErrorInvalidSystemTime) Is(target error) bool {
	return target == ErrBuildErrorInvalidSystemTime
}

type BuildErrorInvalidChannelMonitor struct {
	message string
}

func NewBuildErrorInvalidChannelMonitor() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidChannelMonitor{},
	}
}

func (err BuildErrorInvalidChannelMonitor) Error() string {
	return fmt.Sprintf("InvalidChannelMonitor: %s", err.message)
}

func (self BuildErrorInvalidChannelMonitor) Is(target error) bool {
	return target == ErrBuildErrorInvalidChannelMonitor
}

type BuildErrorInvalidListeningAddresses struct {
	message string
}

func NewBuildErrorInvalidListeningAddresses() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidListeningAddresses{},
	}
}

func (err BuildErrorInvalidListeningAddresses) Error() string {
	return fmt.Sprintf("InvalidListeningAddresses: %s", err.message)
}

func (self BuildErrorInvalidListeningAddresses) Is(target error) bool {
	return target == ErrBuildErrorInvalidListeningAddresses
}

type BuildErrorInvalidNodeAlias struct {
	message string
}

func NewBuildErrorInvalidNodeAlias() *BuildError {
	return &BuildError{
		err: &BuildErrorInvalidNodeAlias{},
	}
}

func (err BuildErrorInvalidNodeAlias) Error() string {
	return fmt.Sprintf("InvalidNodeAlias: %s", err.message)
}

func (self BuildErrorInvalidNodeAlias) Is(target error) bool {
	return target == ErrBuildErrorInvalidNodeAlias
}

type BuildErrorReadFailed struct {
	message string
}

func NewBuildErrorReadFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorReadFailed{},
	}
}

func (err BuildErrorReadFailed) Error() string {
	return fmt.Sprintf("ReadFailed: %s", err.message)
}

func (self BuildErrorReadFailed) Is(target error) bool {
	return target == ErrBuildErrorReadFailed
}

type BuildErrorWriteFailed struct {
	message string
}

func NewBuildErrorWriteFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorWriteFailed{},
	}
}

func (err BuildErrorWriteFailed) Error() string {
	return fmt.Sprintf("WriteFailed: %s", err.message)
}

func (self BuildErrorWriteFailed) Is(target error) bool {
	return target == ErrBuildErrorWriteFailed
}

type BuildErrorStoragePathAccessFailed struct {
	message string
}

func NewBuildErrorStoragePathAccessFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorStoragePathAccessFailed{},
	}
}

func (err BuildErrorStoragePathAccessFailed) Error() string {
	return fmt.Sprintf("StoragePathAccessFailed: %s", err.message)
}

func (self BuildErrorStoragePathAccessFailed) Is(target error) bool {
	return target == ErrBuildErrorStoragePathAccessFailed
}

type BuildErrorKvStoreSetupFailed struct {
	message string
}

func NewBuildErrorKvStoreSetupFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorKvStoreSetupFailed{},
	}
}

func (err BuildErrorKvStoreSetupFailed) Error() string {
	return fmt.Sprintf("KvStoreSetupFailed: %s", err.message)
}

func (self BuildErrorKvStoreSetupFailed) Is(target error) bool {
	return target == ErrBuildErrorKvStoreSetupFailed
}

type BuildErrorWalletSetupFailed struct {
	message string
}

func NewBuildErrorWalletSetupFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorWalletSetupFailed{},
	}
}

func (err BuildErrorWalletSetupFailed) Error() string {
	return fmt.Sprintf("WalletSetupFailed: %s", err.message)
}

func (self BuildErrorWalletSetupFailed) Is(target error) bool {
	return target == ErrBuildErrorWalletSetupFailed
}

type BuildErrorLoggerSetupFailed struct {
	message string
}

func NewBuildErrorLoggerSetupFailed() *BuildError {
	return &BuildError{
		err: &BuildErrorLoggerSetupFailed{},
	}
}

func (err BuildErrorLoggerSetupFailed) Error() string {
	return fmt.Sprintf("LoggerSetupFailed: %s", err.message)
}

func (self BuildErrorLoggerSetupFailed) Is(target error) bool {
	return target == ErrBuildErrorLoggerSetupFailed
}

type FfiConverterTypeBuildError struct{}

var FfiConverterTypeBuildErrorINSTANCE = FfiConverterTypeBuildError{}

func (c FfiConverterTypeBuildError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeBuildError) Lower(value *BuildError) RustBuffer {
	return LowerIntoRustBuffer[*BuildError](c, value)
}

func (c FfiConverterTypeBuildError) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	message := FfiConverterStringINSTANCE.Read(reader)
	switch errorID {
	case 1:
		return &BuildError{&BuildErrorInvalidSeedBytes{message}}
	case 2:
		return &BuildError{&BuildErrorInvalidSeedFile{message}}
	case 3:
		return &BuildError{&BuildErrorInvalidSystemTime{message}}
	case 4:
		return &BuildError{&BuildErrorInvalidChannelMonitor{message}}
	case 5:
		return &BuildError{&BuildErrorInvalidListeningAddresses{message}}
	case 6:
		return &BuildError{&BuildErrorInvalidNodeAlias{message}}
	case 7:
		return &BuildError{&BuildErrorReadFailed{message}}
	case 8:
		return &BuildError{&BuildErrorWriteFailed{message}}
	case 9:
		return &BuildError{&BuildErrorStoragePathAccessFailed{message}}
	case 10:
		return &BuildError{&BuildErrorKvStoreSetupFailed{message}}
	case 11:
		return &BuildError{&BuildErrorWalletSetupFailed{message}}
	case 12:
		return &BuildError{&BuildErrorLoggerSetupFailed{message}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeBuildError.Read()", errorID))
	}

}

func (c FfiConverterTypeBuildError) Write(writer io.Writer, value *BuildError) {
	switch variantValue := value.err.(type) {
	case *BuildErrorInvalidSeedBytes:
		writeInt32(writer, 1)
	case *BuildErrorInvalidSeedFile:
		writeInt32(writer, 2)
	case *BuildErrorInvalidSystemTime:
		writeInt32(writer, 3)
	case *BuildErrorInvalidChannelMonitor:
		writeInt32(writer, 4)
	case *BuildErrorInvalidListeningAddresses:
		writeInt32(writer, 5)
	case *BuildErrorInvalidNodeAlias:
		writeInt32(writer, 6)
	case *BuildErrorReadFailed:
		writeInt32(writer, 7)
	case *BuildErrorWriteFailed:
		writeInt32(writer, 8)
	case *BuildErrorStoragePathAccessFailed:
		writeInt32(writer, 9)
	case *BuildErrorKvStoreSetupFailed:
		writeInt32(writer, 10)
	case *BuildErrorWalletSetupFailed:
		writeInt32(writer, 11)
	case *BuildErrorLoggerSetupFailed:
		writeInt32(writer, 12)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeBuildError.Write", value))
	}
}

type ChannelType uint

const (
	ChannelTypeStaticRemoteKey ChannelType = 1
	ChannelTypeAnchors         ChannelType = 2
)

type FfiConverterTypeChannelType struct{}

var FfiConverterTypeChannelTypeINSTANCE = FfiConverterTypeChannelType{}

func (c FfiConverterTypeChannelType) Lift(rb RustBufferI) ChannelType {
	return LiftFromRustBuffer[ChannelType](c, rb)
}

func (c FfiConverterTypeChannelType) Lower(value ChannelType) RustBuffer {
	return LowerIntoRustBuffer[ChannelType](c, value)
}
func (FfiConverterTypeChannelType) Read(reader io.Reader) ChannelType {
	id := readInt32(reader)
	return ChannelType(id)
}

func (FfiConverterTypeChannelType) Write(writer io.Writer, value ChannelType) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeChannelType struct{}

func (_ FfiDestroyerTypeChannelType) Destroy(value ChannelType) {
}

type ClosureReason interface {
	Destroy()
}
type ClosureReasonCounterpartyForceClosed struct {
	PeerMsg UntrustedString
}

func (e ClosureReasonCounterpartyForceClosed) Destroy() {
	FfiDestroyerTypeUntrustedString{}.Destroy(e.PeerMsg)
}

type ClosureReasonHolderForceClosed struct {
	BroadcastedLatestTxn *bool
}

func (e ClosureReasonHolderForceClosed) Destroy() {
	FfiDestroyerOptionalBool{}.Destroy(e.BroadcastedLatestTxn)
}

type ClosureReasonLegacyCooperativeClosure struct {
}

func (e ClosureReasonLegacyCooperativeClosure) Destroy() {
}

type ClosureReasonCounterpartyInitiatedCooperativeClosure struct {
}

func (e ClosureReasonCounterpartyInitiatedCooperativeClosure) Destroy() {
}

type ClosureReasonLocallyInitiatedCooperativeClosure struct {
}

func (e ClosureReasonLocallyInitiatedCooperativeClosure) Destroy() {
}

type ClosureReasonCommitmentTxConfirmed struct {
}

func (e ClosureReasonCommitmentTxConfirmed) Destroy() {
}

type ClosureReasonFundingTimedOut struct {
}

func (e ClosureReasonFundingTimedOut) Destroy() {
}

type ClosureReasonProcessingError struct {
	Err string
}

func (e ClosureReasonProcessingError) Destroy() {
	FfiDestroyerString{}.Destroy(e.Err)
}

type ClosureReasonDisconnectedPeer struct {
}

func (e ClosureReasonDisconnectedPeer) Destroy() {
}

type ClosureReasonOutdatedChannelManager struct {
}

func (e ClosureReasonOutdatedChannelManager) Destroy() {
}

type ClosureReasonCounterpartyCoopClosedUnfundedChannel struct {
}

func (e ClosureReasonCounterpartyCoopClosedUnfundedChannel) Destroy() {
}

type ClosureReasonFundingBatchClosure struct {
}

func (e ClosureReasonFundingBatchClosure) Destroy() {
}

type ClosureReasonHtlCsTimedOut struct {
}

func (e ClosureReasonHtlCsTimedOut) Destroy() {
}

type ClosureReasonPeerFeerateTooLow struct {
	PeerFeerateSatPerKw     uint32
	RequiredFeerateSatPerKw uint32
}

func (e ClosureReasonPeerFeerateTooLow) Destroy() {
	FfiDestroyerUint32{}.Destroy(e.PeerFeerateSatPerKw)
	FfiDestroyerUint32{}.Destroy(e.RequiredFeerateSatPerKw)
}

type FfiConverterTypeClosureReason struct{}

var FfiConverterTypeClosureReasonINSTANCE = FfiConverterTypeClosureReason{}

func (c FfiConverterTypeClosureReason) Lift(rb RustBufferI) ClosureReason {
	return LiftFromRustBuffer[ClosureReason](c, rb)
}

func (c FfiConverterTypeClosureReason) Lower(value ClosureReason) RustBuffer {
	return LowerIntoRustBuffer[ClosureReason](c, value)
}
func (FfiConverterTypeClosureReason) Read(reader io.Reader) ClosureReason {
	id := readInt32(reader)
	switch id {
	case 1:
		return ClosureReasonCounterpartyForceClosed{
			FfiConverterTypeUntrustedStringINSTANCE.Read(reader),
		}
	case 2:
		return ClosureReasonHolderForceClosed{
			FfiConverterOptionalBoolINSTANCE.Read(reader),
		}
	case 3:
		return ClosureReasonLegacyCooperativeClosure{}
	case 4:
		return ClosureReasonCounterpartyInitiatedCooperativeClosure{}
	case 5:
		return ClosureReasonLocallyInitiatedCooperativeClosure{}
	case 6:
		return ClosureReasonCommitmentTxConfirmed{}
	case 7:
		return ClosureReasonFundingTimedOut{}
	case 8:
		return ClosureReasonProcessingError{
			FfiConverterStringINSTANCE.Read(reader),
		}
	case 9:
		return ClosureReasonDisconnectedPeer{}
	case 10:
		return ClosureReasonOutdatedChannelManager{}
	case 11:
		return ClosureReasonCounterpartyCoopClosedUnfundedChannel{}
	case 12:
		return ClosureReasonFundingBatchClosure{}
	case 13:
		return ClosureReasonHtlCsTimedOut{}
	case 14:
		return ClosureReasonPeerFeerateTooLow{
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeClosureReason.Read()", id))
	}
}

func (FfiConverterTypeClosureReason) Write(writer io.Writer, value ClosureReason) {
	switch variant_value := value.(type) {
	case ClosureReasonCounterpartyForceClosed:
		writeInt32(writer, 1)
		FfiConverterTypeUntrustedStringINSTANCE.Write(writer, variant_value.PeerMsg)
	case ClosureReasonHolderForceClosed:
		writeInt32(writer, 2)
		FfiConverterOptionalBoolINSTANCE.Write(writer, variant_value.BroadcastedLatestTxn)
	case ClosureReasonLegacyCooperativeClosure:
		writeInt32(writer, 3)
	case ClosureReasonCounterpartyInitiatedCooperativeClosure:
		writeInt32(writer, 4)
	case ClosureReasonLocallyInitiatedCooperativeClosure:
		writeInt32(writer, 5)
	case ClosureReasonCommitmentTxConfirmed:
		writeInt32(writer, 6)
	case ClosureReasonFundingTimedOut:
		writeInt32(writer, 7)
	case ClosureReasonProcessingError:
		writeInt32(writer, 8)
		FfiConverterStringINSTANCE.Write(writer, variant_value.Err)
	case ClosureReasonDisconnectedPeer:
		writeInt32(writer, 9)
	case ClosureReasonOutdatedChannelManager:
		writeInt32(writer, 10)
	case ClosureReasonCounterpartyCoopClosedUnfundedChannel:
		writeInt32(writer, 11)
	case ClosureReasonFundingBatchClosure:
		writeInt32(writer, 12)
	case ClosureReasonHtlCsTimedOut:
		writeInt32(writer, 13)
	case ClosureReasonPeerFeerateTooLow:
		writeInt32(writer, 14)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.PeerFeerateSatPerKw)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.RequiredFeerateSatPerKw)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeClosureReason.Write", value))
	}
}

type FfiDestroyerTypeClosureReason struct{}

func (_ FfiDestroyerTypeClosureReason) Destroy(value ClosureReason) {
	value.Destroy()
}

type Event interface {
	Destroy()
}
type EventPaymentSuccessful struct {
	PaymentId   *PaymentId
	PaymentHash PaymentHash
	FeePaidMsat *uint64
}

func (e EventPaymentSuccessful) Destroy() {
	FfiDestroyerOptionalTypePaymentId{}.Destroy(e.PaymentId)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
	FfiDestroyerOptionalUint64{}.Destroy(e.FeePaidMsat)
}

type EventPaymentFailed struct {
	PaymentId   *PaymentId
	PaymentHash *PaymentHash
	Reason      *PaymentFailureReason
}

func (e EventPaymentFailed) Destroy() {
	FfiDestroyerOptionalTypePaymentId{}.Destroy(e.PaymentId)
	FfiDestroyerOptionalTypePaymentHash{}.Destroy(e.PaymentHash)
	FfiDestroyerOptionalTypePaymentFailureReason{}.Destroy(e.Reason)
}

type EventPaymentReceived struct {
	PaymentId   *PaymentId
	PaymentHash PaymentHash
	AmountMsat  uint64
}

func (e EventPaymentReceived) Destroy() {
	FfiDestroyerOptionalTypePaymentId{}.Destroy(e.PaymentId)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
	FfiDestroyerUint64{}.Destroy(e.AmountMsat)
}

type EventPaymentClaimable struct {
	PaymentId           PaymentId
	PaymentHash         PaymentHash
	ClaimableAmountMsat uint64
	ClaimDeadline       *uint32
}

func (e EventPaymentClaimable) Destroy() {
	FfiDestroyerTypePaymentId{}.Destroy(e.PaymentId)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
	FfiDestroyerUint64{}.Destroy(e.ClaimableAmountMsat)
	FfiDestroyerOptionalUint32{}.Destroy(e.ClaimDeadline)
}

type EventChannelPending struct {
	ChannelId                ChannelId
	UserChannelId            UserChannelId
	FormerTemporaryChannelId ChannelId
	CounterpartyNodeId       PublicKey
	FundingTxo               OutPoint
}

func (e EventChannelPending) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypeUserChannelId{}.Destroy(e.UserChannelId)
	FfiDestroyerTypeChannelId{}.Destroy(e.FormerTemporaryChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerTypeOutPoint{}.Destroy(e.FundingTxo)
}

type EventChannelReady struct {
	ChannelId          ChannelId
	UserChannelId      UserChannelId
	CounterpartyNodeId *PublicKey
}

func (e EventChannelReady) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypeUserChannelId{}.Destroy(e.UserChannelId)
	FfiDestroyerOptionalTypePublicKey{}.Destroy(e.CounterpartyNodeId)
}

type EventChannelClosed struct {
	ChannelId          ChannelId
	UserChannelId      UserChannelId
	CounterpartyNodeId *PublicKey
	Reason             *ClosureReason
}

func (e EventChannelClosed) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypeUserChannelId{}.Destroy(e.UserChannelId)
	FfiDestroyerOptionalTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerOptionalTypeClosureReason{}.Destroy(e.Reason)
}

type FfiConverterTypeEvent struct{}

var FfiConverterTypeEventINSTANCE = FfiConverterTypeEvent{}

func (c FfiConverterTypeEvent) Lift(rb RustBufferI) Event {
	return LiftFromRustBuffer[Event](c, rb)
}

func (c FfiConverterTypeEvent) Lower(value Event) RustBuffer {
	return LowerIntoRustBuffer[Event](c, value)
}
func (FfiConverterTypeEvent) Read(reader io.Reader) Event {
	id := readInt32(reader)
	switch id {
	case 1:
		return EventPaymentSuccessful{
			FfiConverterOptionalTypePaymentIdINSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalUint64INSTANCE.Read(reader),
		}
	case 2:
		return EventPaymentFailed{
			FfiConverterOptionalTypePaymentIdINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentFailureReasonINSTANCE.Read(reader),
		}
	case 3:
		return EventPaymentReceived{
			FfiConverterOptionalTypePaymentIdINSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
		}
	case 4:
		return EventPaymentClaimable{
			FfiConverterTypePaymentIdINSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterOptionalUint32INSTANCE.Read(reader),
		}
	case 5:
		return EventChannelPending{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypeUserChannelIdINSTANCE.Read(reader),
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterTypeOutPointINSTANCE.Read(reader),
		}
	case 6:
		return EventChannelReady{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypeUserChannelIdINSTANCE.Read(reader),
			FfiConverterOptionalTypePublicKeyINSTANCE.Read(reader),
		}
	case 7:
		return EventChannelClosed{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypeUserChannelIdINSTANCE.Read(reader),
			FfiConverterOptionalTypePublicKeyINSTANCE.Read(reader),
			FfiConverterOptionalTypeClosureReasonINSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeEvent.Read()", id))
	}
}

func (FfiConverterTypeEvent) Write(writer io.Writer, value Event) {
	switch variant_value := value.(type) {
	case EventPaymentSuccessful:
		writeInt32(writer, 1)
		FfiConverterOptionalTypePaymentIdINSTANCE.Write(writer, variant_value.PaymentId)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
		FfiConverterOptionalUint64INSTANCE.Write(writer, variant_value.FeePaidMsat)
	case EventPaymentFailed:
		writeInt32(writer, 2)
		FfiConverterOptionalTypePaymentIdINSTANCE.Write(writer, variant_value.PaymentId)
		FfiConverterOptionalTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
		FfiConverterOptionalTypePaymentFailureReasonINSTANCE.Write(writer, variant_value.Reason)
	case EventPaymentReceived:
		writeInt32(writer, 3)
		FfiConverterOptionalTypePaymentIdINSTANCE.Write(writer, variant_value.PaymentId)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountMsat)
	case EventPaymentClaimable:
		writeInt32(writer, 4)
		FfiConverterTypePaymentIdINSTANCE.Write(writer, variant_value.PaymentId)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.ClaimableAmountMsat)
		FfiConverterOptionalUint32INSTANCE.Write(writer, variant_value.ClaimDeadline)
	case EventChannelPending:
		writeInt32(writer, 5)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypeUserChannelIdINSTANCE.Write(writer, variant_value.UserChannelId)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.FormerTemporaryChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterTypeOutPointINSTANCE.Write(writer, variant_value.FundingTxo)
	case EventChannelReady:
		writeInt32(writer, 6)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypeUserChannelIdINSTANCE.Write(writer, variant_value.UserChannelId)
		FfiConverterOptionalTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
	case EventChannelClosed:
		writeInt32(writer, 7)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypeUserChannelIdINSTANCE.Write(writer, variant_value.UserChannelId)
		FfiConverterOptionalTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterOptionalTypeClosureReasonINSTANCE.Write(writer, variant_value.Reason)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeEvent.Write", value))
	}
}

type FfiDestroyerTypeEvent struct{}

func (_ FfiDestroyerTypeEvent) Destroy(value Event) {
	value.Destroy()
}

type LightningBalance interface {
	Destroy()
}
type LightningBalanceClaimableOnChannelClose struct {
	ChannelId                        ChannelId
	CounterpartyNodeId               PublicKey
	FundingTxId                      Txid
	FundingTxIndex                   uint16
	AmountSatoshis                   uint64
	TransactionFeeSatoshis           uint64
	OutboundPaymentHtlcRoundedMsat   uint64
	OutboundForwardedHtlcRoundedMsat uint64
	InboundClaimingHtlcRoundedMsat   uint64
	InboundHtlcRoundedMsat           uint64
}

func (e LightningBalanceClaimableOnChannelClose) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerTypeTxid{}.Destroy(e.FundingTxId)
	FfiDestroyerUint16{}.Destroy(e.FundingTxIndex)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerUint64{}.Destroy(e.TransactionFeeSatoshis)
	FfiDestroyerUint64{}.Destroy(e.OutboundPaymentHtlcRoundedMsat)
	FfiDestroyerUint64{}.Destroy(e.OutboundForwardedHtlcRoundedMsat)
	FfiDestroyerUint64{}.Destroy(e.InboundClaimingHtlcRoundedMsat)
	FfiDestroyerUint64{}.Destroy(e.InboundHtlcRoundedMsat)
}

type LightningBalanceClaimableAwaitingConfirmations struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	FundingTxId        Txid
	FundingTxIndex     uint16
	AmountSatoshis     uint64
	ConfirmationHeight uint32
	Source             BalanceSource
}

func (e LightningBalanceClaimableAwaitingConfirmations) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerTypeTxid{}.Destroy(e.FundingTxId)
	FfiDestroyerUint16{}.Destroy(e.FundingTxIndex)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerUint32{}.Destroy(e.ConfirmationHeight)
	FfiDestroyerTypeBalanceSource{}.Destroy(e.Source)
}

type LightningBalanceContentiousClaimable struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	FundingTxId        Txid
	FundingTxIndex     uint16
	AmountSatoshis     uint64
	TimeoutHeight      uint32
	PaymentHash        PaymentHash
	PaymentPreimage    PaymentPreimage
}

func (e LightningBalanceContentiousClaimable) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerTypeTxid{}.Destroy(e.FundingTxId)
	FfiDestroyerUint16{}.Destroy(e.FundingTxIndex)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerUint32{}.Destroy(e.TimeoutHeight)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
	FfiDestroyerTypePaymentPreimage{}.Destroy(e.PaymentPreimage)
}

type LightningBalanceMaybeTimeoutClaimableHtlc struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	FundingTxId        Txid
	FundingTxIndex     uint16
	AmountSatoshis     uint64
	ClaimableHeight    uint32
	PaymentHash        PaymentHash
	OutboundPayment    bool
}

func (e LightningBalanceMaybeTimeoutClaimableHtlc) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerTypeTxid{}.Destroy(e.FundingTxId)
	FfiDestroyerUint16{}.Destroy(e.FundingTxIndex)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerUint32{}.Destroy(e.ClaimableHeight)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
	FfiDestroyerBool{}.Destroy(e.OutboundPayment)
}

type LightningBalanceMaybePreimageClaimableHtlc struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	FundingTxId        Txid
	FundingTxIndex     uint16
	AmountSatoshis     uint64
	ExpiryHeight       uint32
	PaymentHash        PaymentHash
}

func (e LightningBalanceMaybePreimageClaimableHtlc) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerTypeTxid{}.Destroy(e.FundingTxId)
	FfiDestroyerUint16{}.Destroy(e.FundingTxIndex)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerUint32{}.Destroy(e.ExpiryHeight)
	FfiDestroyerTypePaymentHash{}.Destroy(e.PaymentHash)
}

type LightningBalanceCounterpartyRevokedOutputClaimable struct {
	ChannelId          ChannelId
	CounterpartyNodeId PublicKey
	FundingTxId        Txid
	FundingTxIndex     uint16
	AmountSatoshis     uint64
}

func (e LightningBalanceCounterpartyRevokedOutputClaimable) Destroy() {
	FfiDestroyerTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerTypeTxid{}.Destroy(e.FundingTxId)
	FfiDestroyerUint16{}.Destroy(e.FundingTxIndex)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
}

type FfiConverterTypeLightningBalance struct{}

var FfiConverterTypeLightningBalanceINSTANCE = FfiConverterTypeLightningBalance{}

func (c FfiConverterTypeLightningBalance) Lift(rb RustBufferI) LightningBalance {
	return LiftFromRustBuffer[LightningBalance](c, rb)
}

func (c FfiConverterTypeLightningBalance) Lower(value LightningBalance) RustBuffer {
	return LowerIntoRustBuffer[LightningBalance](c, value)
}
func (FfiConverterTypeLightningBalance) Read(reader io.Reader) LightningBalance {
	id := readInt32(reader)
	switch id {
	case 1:
		return LightningBalanceClaimableOnChannelClose{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterTypeTxidINSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
		}
	case 2:
		return LightningBalanceClaimableAwaitingConfirmations{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterTypeTxidINSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterTypeBalanceSourceINSTANCE.Read(reader),
		}
	case 3:
		return LightningBalanceContentiousClaimable{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterTypeTxidINSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterTypePaymentPreimageINSTANCE.Read(reader),
		}
	case 4:
		return LightningBalanceMaybeTimeoutClaimableHtlc{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterTypeTxidINSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterBoolINSTANCE.Read(reader),
		}
	case 5:
		return LightningBalanceMaybePreimageClaimableHtlc{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterTypeTxidINSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
		}
	case 6:
		return LightningBalanceCounterpartyRevokedOutputClaimable{
			FfiConverterTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypePublicKeyINSTANCE.Read(reader),
			FfiConverterTypeTxidINSTANCE.Read(reader),
			FfiConverterUint16INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeLightningBalance.Read()", id))
	}
}

func (FfiConverterTypeLightningBalance) Write(writer io.Writer, value LightningBalance) {
	switch variant_value := value.(type) {
	case LightningBalanceClaimableOnChannelClose:
		writeInt32(writer, 1)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.FundingTxId)
		FfiConverterUint16INSTANCE.Write(writer, variant_value.FundingTxIndex)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.TransactionFeeSatoshis)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.OutboundPaymentHtlcRoundedMsat)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.OutboundForwardedHtlcRoundedMsat)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.InboundClaimingHtlcRoundedMsat)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.InboundHtlcRoundedMsat)
	case LightningBalanceClaimableAwaitingConfirmations:
		writeInt32(writer, 2)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.FundingTxId)
		FfiConverterUint16INSTANCE.Write(writer, variant_value.FundingTxIndex)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.ConfirmationHeight)
		FfiConverterTypeBalanceSourceINSTANCE.Write(writer, variant_value.Source)
	case LightningBalanceContentiousClaimable:
		writeInt32(writer, 3)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.FundingTxId)
		FfiConverterUint16INSTANCE.Write(writer, variant_value.FundingTxIndex)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.TimeoutHeight)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
		FfiConverterTypePaymentPreimageINSTANCE.Write(writer, variant_value.PaymentPreimage)
	case LightningBalanceMaybeTimeoutClaimableHtlc:
		writeInt32(writer, 4)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.FundingTxId)
		FfiConverterUint16INSTANCE.Write(writer, variant_value.FundingTxIndex)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.ClaimableHeight)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
		FfiConverterBoolINSTANCE.Write(writer, variant_value.OutboundPayment)
	case LightningBalanceMaybePreimageClaimableHtlc:
		writeInt32(writer, 5)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.FundingTxId)
		FfiConverterUint16INSTANCE.Write(writer, variant_value.FundingTxIndex)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.ExpiryHeight)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.PaymentHash)
	case LightningBalanceCounterpartyRevokedOutputClaimable:
		writeInt32(writer, 6)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.FundingTxId)
		FfiConverterUint16INSTANCE.Write(writer, variant_value.FundingTxIndex)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeLightningBalance.Write", value))
	}
}

type FfiDestroyerTypeLightningBalance struct{}

func (_ FfiDestroyerTypeLightningBalance) Destroy(value LightningBalance) {
	value.Destroy()
}

type LogLevel uint

const (
	LogLevelGossip LogLevel = 1
	LogLevelTrace  LogLevel = 2
	LogLevelDebug  LogLevel = 3
	LogLevelInfo   LogLevel = 4
	LogLevelWarn   LogLevel = 5
	LogLevelError  LogLevel = 6
)

type FfiConverterTypeLogLevel struct{}

var FfiConverterTypeLogLevelINSTANCE = FfiConverterTypeLogLevel{}

func (c FfiConverterTypeLogLevel) Lift(rb RustBufferI) LogLevel {
	return LiftFromRustBuffer[LogLevel](c, rb)
}

func (c FfiConverterTypeLogLevel) Lower(value LogLevel) RustBuffer {
	return LowerIntoRustBuffer[LogLevel](c, value)
}
func (FfiConverterTypeLogLevel) Read(reader io.Reader) LogLevel {
	id := readInt32(reader)
	return LogLevel(id)
}

func (FfiConverterTypeLogLevel) Write(writer io.Writer, value LogLevel) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeLogLevel struct{}

func (_ FfiDestroyerTypeLogLevel) Destroy(value LogLevel) {
}

type MaxDustHtlcExposure interface {
	Destroy()
}
type MaxDustHtlcExposureFixedLimit struct {
	LimitMsat uint64
}

func (e MaxDustHtlcExposureFixedLimit) Destroy() {
	FfiDestroyerUint64{}.Destroy(e.LimitMsat)
}

type MaxDustHtlcExposureFeeRateMultiplier struct {
	Multiplier uint64
}

func (e MaxDustHtlcExposureFeeRateMultiplier) Destroy() {
	FfiDestroyerUint64{}.Destroy(e.Multiplier)
}

type FfiConverterTypeMaxDustHTLCExposure struct{}

var FfiConverterTypeMaxDustHTLCExposureINSTANCE = FfiConverterTypeMaxDustHTLCExposure{}

func (c FfiConverterTypeMaxDustHTLCExposure) Lift(rb RustBufferI) MaxDustHtlcExposure {
	return LiftFromRustBuffer[MaxDustHtlcExposure](c, rb)
}

func (c FfiConverterTypeMaxDustHTLCExposure) Lower(value MaxDustHtlcExposure) RustBuffer {
	return LowerIntoRustBuffer[MaxDustHtlcExposure](c, value)
}
func (FfiConverterTypeMaxDustHTLCExposure) Read(reader io.Reader) MaxDustHtlcExposure {
	id := readInt32(reader)
	switch id {
	case 1:
		return MaxDustHtlcExposureFixedLimit{
			FfiConverterUint64INSTANCE.Read(reader),
		}
	case 2:
		return MaxDustHtlcExposureFeeRateMultiplier{
			FfiConverterUint64INSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeMaxDustHTLCExposure.Read()", id))
	}
}

func (FfiConverterTypeMaxDustHTLCExposure) Write(writer io.Writer, value MaxDustHtlcExposure) {
	switch variant_value := value.(type) {
	case MaxDustHtlcExposureFixedLimit:
		writeInt32(writer, 1)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.LimitMsat)
	case MaxDustHtlcExposureFeeRateMultiplier:
		writeInt32(writer, 2)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.Multiplier)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeMaxDustHTLCExposure.Write", value))
	}
}

type FfiDestroyerTypeMaxDustHtlcExposure struct{}

func (_ FfiDestroyerTypeMaxDustHtlcExposure) Destroy(value MaxDustHtlcExposure) {
	value.Destroy()
}

type MaxTotalRoutingFeeLimit interface {
	Destroy()
}
type MaxTotalRoutingFeeLimitNone struct {
}

func (e MaxTotalRoutingFeeLimitNone) Destroy() {
}

type MaxTotalRoutingFeeLimitSome struct {
	AmountMsat uint64
}

func (e MaxTotalRoutingFeeLimitSome) Destroy() {
	FfiDestroyerUint64{}.Destroy(e.AmountMsat)
}

type FfiConverterTypeMaxTotalRoutingFeeLimit struct{}

var FfiConverterTypeMaxTotalRoutingFeeLimitINSTANCE = FfiConverterTypeMaxTotalRoutingFeeLimit{}

func (c FfiConverterTypeMaxTotalRoutingFeeLimit) Lift(rb RustBufferI) MaxTotalRoutingFeeLimit {
	return LiftFromRustBuffer[MaxTotalRoutingFeeLimit](c, rb)
}

func (c FfiConverterTypeMaxTotalRoutingFeeLimit) Lower(value MaxTotalRoutingFeeLimit) RustBuffer {
	return LowerIntoRustBuffer[MaxTotalRoutingFeeLimit](c, value)
}
func (FfiConverterTypeMaxTotalRoutingFeeLimit) Read(reader io.Reader) MaxTotalRoutingFeeLimit {
	id := readInt32(reader)
	switch id {
	case 1:
		return MaxTotalRoutingFeeLimitNone{}
	case 2:
		return MaxTotalRoutingFeeLimitSome{
			FfiConverterUint64INSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeMaxTotalRoutingFeeLimit.Read()", id))
	}
}

func (FfiConverterTypeMaxTotalRoutingFeeLimit) Write(writer io.Writer, value MaxTotalRoutingFeeLimit) {
	switch variant_value := value.(type) {
	case MaxTotalRoutingFeeLimitNone:
		writeInt32(writer, 1)
	case MaxTotalRoutingFeeLimitSome:
		writeInt32(writer, 2)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountMsat)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeMaxTotalRoutingFeeLimit.Write", value))
	}
}

type FfiDestroyerTypeMaxTotalRoutingFeeLimit struct{}

func (_ FfiDestroyerTypeMaxTotalRoutingFeeLimit) Destroy(value MaxTotalRoutingFeeLimit) {
	value.Destroy()
}

type MigrateStorage uint

const (
	MigrateStorageVss MigrateStorage = 1
)

type FfiConverterTypeMigrateStorage struct{}

var FfiConverterTypeMigrateStorageINSTANCE = FfiConverterTypeMigrateStorage{}

func (c FfiConverterTypeMigrateStorage) Lift(rb RustBufferI) MigrateStorage {
	return LiftFromRustBuffer[MigrateStorage](c, rb)
}

func (c FfiConverterTypeMigrateStorage) Lower(value MigrateStorage) RustBuffer {
	return LowerIntoRustBuffer[MigrateStorage](c, value)
}
func (FfiConverterTypeMigrateStorage) Read(reader io.Reader) MigrateStorage {
	id := readInt32(reader)
	return MigrateStorage(id)
}

func (FfiConverterTypeMigrateStorage) Write(writer io.Writer, value MigrateStorage) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeMigrateStorage struct{}

func (_ FfiDestroyerTypeMigrateStorage) Destroy(value MigrateStorage) {
}

type NodeError struct {
	err error
}

func (err NodeError) Error() string {
	return fmt.Sprintf("NodeError: %s", err.err.Error())
}

func (err NodeError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrNodeErrorAlreadyRunning = fmt.Errorf("NodeErrorAlreadyRunning")
var ErrNodeErrorNotRunning = fmt.Errorf("NodeErrorNotRunning")
var ErrNodeErrorOnchainTxCreationFailed = fmt.Errorf("NodeErrorOnchainTxCreationFailed")
var ErrNodeErrorConnectionFailed = fmt.Errorf("NodeErrorConnectionFailed")
var ErrNodeErrorInvoiceCreationFailed = fmt.Errorf("NodeErrorInvoiceCreationFailed")
var ErrNodeErrorInvoiceRequestCreationFailed = fmt.Errorf("NodeErrorInvoiceRequestCreationFailed")
var ErrNodeErrorOfferCreationFailed = fmt.Errorf("NodeErrorOfferCreationFailed")
var ErrNodeErrorRefundCreationFailed = fmt.Errorf("NodeErrorRefundCreationFailed")
var ErrNodeErrorPaymentSendingFailed = fmt.Errorf("NodeErrorPaymentSendingFailed")
var ErrNodeErrorProbeSendingFailed = fmt.Errorf("NodeErrorProbeSendingFailed")
var ErrNodeErrorChannelCreationFailed = fmt.Errorf("NodeErrorChannelCreationFailed")
var ErrNodeErrorChannelClosingFailed = fmt.Errorf("NodeErrorChannelClosingFailed")
var ErrNodeErrorChannelConfigUpdateFailed = fmt.Errorf("NodeErrorChannelConfigUpdateFailed")
var ErrNodeErrorPersistenceFailed = fmt.Errorf("NodeErrorPersistenceFailed")
var ErrNodeErrorFeerateEstimationUpdateFailed = fmt.Errorf("NodeErrorFeerateEstimationUpdateFailed")
var ErrNodeErrorFeerateEstimationUpdateTimeout = fmt.Errorf("NodeErrorFeerateEstimationUpdateTimeout")
var ErrNodeErrorWalletOperationFailed = fmt.Errorf("NodeErrorWalletOperationFailed")
var ErrNodeErrorWalletOperationTimeout = fmt.Errorf("NodeErrorWalletOperationTimeout")
var ErrNodeErrorOnchainTxSigningFailed = fmt.Errorf("NodeErrorOnchainTxSigningFailed")
var ErrNodeErrorTxSyncFailed = fmt.Errorf("NodeErrorTxSyncFailed")
var ErrNodeErrorTxSyncTimeout = fmt.Errorf("NodeErrorTxSyncTimeout")
var ErrNodeErrorGossipUpdateFailed = fmt.Errorf("NodeErrorGossipUpdateFailed")
var ErrNodeErrorGossipUpdateTimeout = fmt.Errorf("NodeErrorGossipUpdateTimeout")
var ErrNodeErrorLiquidityRequestFailed = fmt.Errorf("NodeErrorLiquidityRequestFailed")
var ErrNodeErrorUriParameterParsingFailed = fmt.Errorf("NodeErrorUriParameterParsingFailed")
var ErrNodeErrorInvalidAddress = fmt.Errorf("NodeErrorInvalidAddress")
var ErrNodeErrorInvalidSocketAddress = fmt.Errorf("NodeErrorInvalidSocketAddress")
var ErrNodeErrorInvalidPublicKey = fmt.Errorf("NodeErrorInvalidPublicKey")
var ErrNodeErrorInvalidSecretKey = fmt.Errorf("NodeErrorInvalidSecretKey")
var ErrNodeErrorInvalidOfferId = fmt.Errorf("NodeErrorInvalidOfferId")
var ErrNodeErrorInvalidNodeId = fmt.Errorf("NodeErrorInvalidNodeId")
var ErrNodeErrorInvalidPaymentId = fmt.Errorf("NodeErrorInvalidPaymentId")
var ErrNodeErrorInvalidPaymentHash = fmt.Errorf("NodeErrorInvalidPaymentHash")
var ErrNodeErrorInvalidPaymentPreimage = fmt.Errorf("NodeErrorInvalidPaymentPreimage")
var ErrNodeErrorInvalidPaymentSecret = fmt.Errorf("NodeErrorInvalidPaymentSecret")
var ErrNodeErrorInvalidAmount = fmt.Errorf("NodeErrorInvalidAmount")
var ErrNodeErrorInvalidInvoice = fmt.Errorf("NodeErrorInvalidInvoice")
var ErrNodeErrorInvalidOffer = fmt.Errorf("NodeErrorInvalidOffer")
var ErrNodeErrorInvalidRefund = fmt.Errorf("NodeErrorInvalidRefund")
var ErrNodeErrorInvalidChannelId = fmt.Errorf("NodeErrorInvalidChannelId")
var ErrNodeErrorInvalidNetwork = fmt.Errorf("NodeErrorInvalidNetwork")
var ErrNodeErrorInvalidCustomTlv = fmt.Errorf("NodeErrorInvalidCustomTlv")
var ErrNodeErrorInvalidUri = fmt.Errorf("NodeErrorInvalidUri")
var ErrNodeErrorInvalidQuantity = fmt.Errorf("NodeErrorInvalidQuantity")
var ErrNodeErrorInvalidNodeAlias = fmt.Errorf("NodeErrorInvalidNodeAlias")
var ErrNodeErrorDuplicatePayment = fmt.Errorf("NodeErrorDuplicatePayment")
var ErrNodeErrorUnsupportedCurrency = fmt.Errorf("NodeErrorUnsupportedCurrency")
var ErrNodeErrorInsufficientFunds = fmt.Errorf("NodeErrorInsufficientFunds")
var ErrNodeErrorLiquiditySourceUnavailable = fmt.Errorf("NodeErrorLiquiditySourceUnavailable")
var ErrNodeErrorLiquidityFeeTooHigh = fmt.Errorf("NodeErrorLiquidityFeeTooHigh")

// Variant structs
type NodeErrorAlreadyRunning struct {
	message string
}

func NewNodeErrorAlreadyRunning() *NodeError {
	return &NodeError{
		err: &NodeErrorAlreadyRunning{},
	}
}

func (err NodeErrorAlreadyRunning) Error() string {
	return fmt.Sprintf("AlreadyRunning: %s", err.message)
}

func (self NodeErrorAlreadyRunning) Is(target error) bool {
	return target == ErrNodeErrorAlreadyRunning
}

type NodeErrorNotRunning struct {
	message string
}

func NewNodeErrorNotRunning() *NodeError {
	return &NodeError{
		err: &NodeErrorNotRunning{},
	}
}

func (err NodeErrorNotRunning) Error() string {
	return fmt.Sprintf("NotRunning: %s", err.message)
}

func (self NodeErrorNotRunning) Is(target error) bool {
	return target == ErrNodeErrorNotRunning
}

type NodeErrorOnchainTxCreationFailed struct {
	message string
}

func NewNodeErrorOnchainTxCreationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorOnchainTxCreationFailed{},
	}
}

func (err NodeErrorOnchainTxCreationFailed) Error() string {
	return fmt.Sprintf("OnchainTxCreationFailed: %s", err.message)
}

func (self NodeErrorOnchainTxCreationFailed) Is(target error) bool {
	return target == ErrNodeErrorOnchainTxCreationFailed
}

type NodeErrorConnectionFailed struct {
	message string
}

func NewNodeErrorConnectionFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorConnectionFailed{},
	}
}

func (err NodeErrorConnectionFailed) Error() string {
	return fmt.Sprintf("ConnectionFailed: %s", err.message)
}

func (self NodeErrorConnectionFailed) Is(target error) bool {
	return target == ErrNodeErrorConnectionFailed
}

type NodeErrorInvoiceCreationFailed struct {
	message string
}

func NewNodeErrorInvoiceCreationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorInvoiceCreationFailed{},
	}
}

func (err NodeErrorInvoiceCreationFailed) Error() string {
	return fmt.Sprintf("InvoiceCreationFailed: %s", err.message)
}

func (self NodeErrorInvoiceCreationFailed) Is(target error) bool {
	return target == ErrNodeErrorInvoiceCreationFailed
}

type NodeErrorInvoiceRequestCreationFailed struct {
	message string
}

func NewNodeErrorInvoiceRequestCreationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorInvoiceRequestCreationFailed{},
	}
}

func (err NodeErrorInvoiceRequestCreationFailed) Error() string {
	return fmt.Sprintf("InvoiceRequestCreationFailed: %s", err.message)
}

func (self NodeErrorInvoiceRequestCreationFailed) Is(target error) bool {
	return target == ErrNodeErrorInvoiceRequestCreationFailed
}

type NodeErrorOfferCreationFailed struct {
	message string
}

func NewNodeErrorOfferCreationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorOfferCreationFailed{},
	}
}

func (err NodeErrorOfferCreationFailed) Error() string {
	return fmt.Sprintf("OfferCreationFailed: %s", err.message)
}

func (self NodeErrorOfferCreationFailed) Is(target error) bool {
	return target == ErrNodeErrorOfferCreationFailed
}

type NodeErrorRefundCreationFailed struct {
	message string
}

func NewNodeErrorRefundCreationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorRefundCreationFailed{},
	}
}

func (err NodeErrorRefundCreationFailed) Error() string {
	return fmt.Sprintf("RefundCreationFailed: %s", err.message)
}

func (self NodeErrorRefundCreationFailed) Is(target error) bool {
	return target == ErrNodeErrorRefundCreationFailed
}

type NodeErrorPaymentSendingFailed struct {
	message string
}

func NewNodeErrorPaymentSendingFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorPaymentSendingFailed{},
	}
}

func (err NodeErrorPaymentSendingFailed) Error() string {
	return fmt.Sprintf("PaymentSendingFailed: %s", err.message)
}

func (self NodeErrorPaymentSendingFailed) Is(target error) bool {
	return target == ErrNodeErrorPaymentSendingFailed
}

type NodeErrorProbeSendingFailed struct {
	message string
}

func NewNodeErrorProbeSendingFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorProbeSendingFailed{},
	}
}

func (err NodeErrorProbeSendingFailed) Error() string {
	return fmt.Sprintf("ProbeSendingFailed: %s", err.message)
}

func (self NodeErrorProbeSendingFailed) Is(target error) bool {
	return target == ErrNodeErrorProbeSendingFailed
}

type NodeErrorChannelCreationFailed struct {
	message string
}

func NewNodeErrorChannelCreationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorChannelCreationFailed{},
	}
}

func (err NodeErrorChannelCreationFailed) Error() string {
	return fmt.Sprintf("ChannelCreationFailed: %s", err.message)
}

func (self NodeErrorChannelCreationFailed) Is(target error) bool {
	return target == ErrNodeErrorChannelCreationFailed
}

type NodeErrorChannelClosingFailed struct {
	message string
}

func NewNodeErrorChannelClosingFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorChannelClosingFailed{},
	}
}

func (err NodeErrorChannelClosingFailed) Error() string {
	return fmt.Sprintf("ChannelClosingFailed: %s", err.message)
}

func (self NodeErrorChannelClosingFailed) Is(target error) bool {
	return target == ErrNodeErrorChannelClosingFailed
}

type NodeErrorChannelConfigUpdateFailed struct {
	message string
}

func NewNodeErrorChannelConfigUpdateFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorChannelConfigUpdateFailed{},
	}
}

func (err NodeErrorChannelConfigUpdateFailed) Error() string {
	return fmt.Sprintf("ChannelConfigUpdateFailed: %s", err.message)
}

func (self NodeErrorChannelConfigUpdateFailed) Is(target error) bool {
	return target == ErrNodeErrorChannelConfigUpdateFailed
}

type NodeErrorPersistenceFailed struct {
	message string
}

func NewNodeErrorPersistenceFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorPersistenceFailed{},
	}
}

func (err NodeErrorPersistenceFailed) Error() string {
	return fmt.Sprintf("PersistenceFailed: %s", err.message)
}

func (self NodeErrorPersistenceFailed) Is(target error) bool {
	return target == ErrNodeErrorPersistenceFailed
}

type NodeErrorFeerateEstimationUpdateFailed struct {
	message string
}

func NewNodeErrorFeerateEstimationUpdateFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorFeerateEstimationUpdateFailed{},
	}
}

func (err NodeErrorFeerateEstimationUpdateFailed) Error() string {
	return fmt.Sprintf("FeerateEstimationUpdateFailed: %s", err.message)
}

func (self NodeErrorFeerateEstimationUpdateFailed) Is(target error) bool {
	return target == ErrNodeErrorFeerateEstimationUpdateFailed
}

type NodeErrorFeerateEstimationUpdateTimeout struct {
	message string
}

func NewNodeErrorFeerateEstimationUpdateTimeout() *NodeError {
	return &NodeError{
		err: &NodeErrorFeerateEstimationUpdateTimeout{},
	}
}

func (err NodeErrorFeerateEstimationUpdateTimeout) Error() string {
	return fmt.Sprintf("FeerateEstimationUpdateTimeout: %s", err.message)
}

func (self NodeErrorFeerateEstimationUpdateTimeout) Is(target error) bool {
	return target == ErrNodeErrorFeerateEstimationUpdateTimeout
}

type NodeErrorWalletOperationFailed struct {
	message string
}

func NewNodeErrorWalletOperationFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorWalletOperationFailed{},
	}
}

func (err NodeErrorWalletOperationFailed) Error() string {
	return fmt.Sprintf("WalletOperationFailed: %s", err.message)
}

func (self NodeErrorWalletOperationFailed) Is(target error) bool {
	return target == ErrNodeErrorWalletOperationFailed
}

type NodeErrorWalletOperationTimeout struct {
	message string
}

func NewNodeErrorWalletOperationTimeout() *NodeError {
	return &NodeError{
		err: &NodeErrorWalletOperationTimeout{},
	}
}

func (err NodeErrorWalletOperationTimeout) Error() string {
	return fmt.Sprintf("WalletOperationTimeout: %s", err.message)
}

func (self NodeErrorWalletOperationTimeout) Is(target error) bool {
	return target == ErrNodeErrorWalletOperationTimeout
}

type NodeErrorOnchainTxSigningFailed struct {
	message string
}

func NewNodeErrorOnchainTxSigningFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorOnchainTxSigningFailed{},
	}
}

func (err NodeErrorOnchainTxSigningFailed) Error() string {
	return fmt.Sprintf("OnchainTxSigningFailed: %s", err.message)
}

func (self NodeErrorOnchainTxSigningFailed) Is(target error) bool {
	return target == ErrNodeErrorOnchainTxSigningFailed
}

type NodeErrorTxSyncFailed struct {
	message string
}

func NewNodeErrorTxSyncFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorTxSyncFailed{},
	}
}

func (err NodeErrorTxSyncFailed) Error() string {
	return fmt.Sprintf("TxSyncFailed: %s", err.message)
}

func (self NodeErrorTxSyncFailed) Is(target error) bool {
	return target == ErrNodeErrorTxSyncFailed
}

type NodeErrorTxSyncTimeout struct {
	message string
}

func NewNodeErrorTxSyncTimeout() *NodeError {
	return &NodeError{
		err: &NodeErrorTxSyncTimeout{},
	}
}

func (err NodeErrorTxSyncTimeout) Error() string {
	return fmt.Sprintf("TxSyncTimeout: %s", err.message)
}

func (self NodeErrorTxSyncTimeout) Is(target error) bool {
	return target == ErrNodeErrorTxSyncTimeout
}

type NodeErrorGossipUpdateFailed struct {
	message string
}

func NewNodeErrorGossipUpdateFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorGossipUpdateFailed{},
	}
}

func (err NodeErrorGossipUpdateFailed) Error() string {
	return fmt.Sprintf("GossipUpdateFailed: %s", err.message)
}

func (self NodeErrorGossipUpdateFailed) Is(target error) bool {
	return target == ErrNodeErrorGossipUpdateFailed
}

type NodeErrorGossipUpdateTimeout struct {
	message string
}

func NewNodeErrorGossipUpdateTimeout() *NodeError {
	return &NodeError{
		err: &NodeErrorGossipUpdateTimeout{},
	}
}

func (err NodeErrorGossipUpdateTimeout) Error() string {
	return fmt.Sprintf("GossipUpdateTimeout: %s", err.message)
}

func (self NodeErrorGossipUpdateTimeout) Is(target error) bool {
	return target == ErrNodeErrorGossipUpdateTimeout
}

type NodeErrorLiquidityRequestFailed struct {
	message string
}

func NewNodeErrorLiquidityRequestFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorLiquidityRequestFailed{},
	}
}

func (err NodeErrorLiquidityRequestFailed) Error() string {
	return fmt.Sprintf("LiquidityRequestFailed: %s", err.message)
}

func (self NodeErrorLiquidityRequestFailed) Is(target error) bool {
	return target == ErrNodeErrorLiquidityRequestFailed
}

type NodeErrorUriParameterParsingFailed struct {
	message string
}

func NewNodeErrorUriParameterParsingFailed() *NodeError {
	return &NodeError{
		err: &NodeErrorUriParameterParsingFailed{},
	}
}

func (err NodeErrorUriParameterParsingFailed) Error() string {
	return fmt.Sprintf("UriParameterParsingFailed: %s", err.message)
}

func (self NodeErrorUriParameterParsingFailed) Is(target error) bool {
	return target == ErrNodeErrorUriParameterParsingFailed
}

type NodeErrorInvalidAddress struct {
	message string
}

func NewNodeErrorInvalidAddress() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidAddress{},
	}
}

func (err NodeErrorInvalidAddress) Error() string {
	return fmt.Sprintf("InvalidAddress: %s", err.message)
}

func (self NodeErrorInvalidAddress) Is(target error) bool {
	return target == ErrNodeErrorInvalidAddress
}

type NodeErrorInvalidSocketAddress struct {
	message string
}

func NewNodeErrorInvalidSocketAddress() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidSocketAddress{},
	}
}

func (err NodeErrorInvalidSocketAddress) Error() string {
	return fmt.Sprintf("InvalidSocketAddress: %s", err.message)
}

func (self NodeErrorInvalidSocketAddress) Is(target error) bool {
	return target == ErrNodeErrorInvalidSocketAddress
}

type NodeErrorInvalidPublicKey struct {
	message string
}

func NewNodeErrorInvalidPublicKey() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidPublicKey{},
	}
}

func (err NodeErrorInvalidPublicKey) Error() string {
	return fmt.Sprintf("InvalidPublicKey: %s", err.message)
}

func (self NodeErrorInvalidPublicKey) Is(target error) bool {
	return target == ErrNodeErrorInvalidPublicKey
}

type NodeErrorInvalidSecretKey struct {
	message string
}

func NewNodeErrorInvalidSecretKey() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidSecretKey{},
	}
}

func (err NodeErrorInvalidSecretKey) Error() string {
	return fmt.Sprintf("InvalidSecretKey: %s", err.message)
}

func (self NodeErrorInvalidSecretKey) Is(target error) bool {
	return target == ErrNodeErrorInvalidSecretKey
}

type NodeErrorInvalidOfferId struct {
	message string
}

func NewNodeErrorInvalidOfferId() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidOfferId{},
	}
}

func (err NodeErrorInvalidOfferId) Error() string {
	return fmt.Sprintf("InvalidOfferId: %s", err.message)
}

func (self NodeErrorInvalidOfferId) Is(target error) bool {
	return target == ErrNodeErrorInvalidOfferId
}

type NodeErrorInvalidNodeId struct {
	message string
}

func NewNodeErrorInvalidNodeId() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidNodeId{},
	}
}

func (err NodeErrorInvalidNodeId) Error() string {
	return fmt.Sprintf("InvalidNodeId: %s", err.message)
}

func (self NodeErrorInvalidNodeId) Is(target error) bool {
	return target == ErrNodeErrorInvalidNodeId
}

type NodeErrorInvalidPaymentId struct {
	message string
}

func NewNodeErrorInvalidPaymentId() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidPaymentId{},
	}
}

func (err NodeErrorInvalidPaymentId) Error() string {
	return fmt.Sprintf("InvalidPaymentId: %s", err.message)
}

func (self NodeErrorInvalidPaymentId) Is(target error) bool {
	return target == ErrNodeErrorInvalidPaymentId
}

type NodeErrorInvalidPaymentHash struct {
	message string
}

func NewNodeErrorInvalidPaymentHash() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidPaymentHash{},
	}
}

func (err NodeErrorInvalidPaymentHash) Error() string {
	return fmt.Sprintf("InvalidPaymentHash: %s", err.message)
}

func (self NodeErrorInvalidPaymentHash) Is(target error) bool {
	return target == ErrNodeErrorInvalidPaymentHash
}

type NodeErrorInvalidPaymentPreimage struct {
	message string
}

func NewNodeErrorInvalidPaymentPreimage() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidPaymentPreimage{},
	}
}

func (err NodeErrorInvalidPaymentPreimage) Error() string {
	return fmt.Sprintf("InvalidPaymentPreimage: %s", err.message)
}

func (self NodeErrorInvalidPaymentPreimage) Is(target error) bool {
	return target == ErrNodeErrorInvalidPaymentPreimage
}

type NodeErrorInvalidPaymentSecret struct {
	message string
}

func NewNodeErrorInvalidPaymentSecret() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidPaymentSecret{},
	}
}

func (err NodeErrorInvalidPaymentSecret) Error() string {
	return fmt.Sprintf("InvalidPaymentSecret: %s", err.message)
}

func (self NodeErrorInvalidPaymentSecret) Is(target error) bool {
	return target == ErrNodeErrorInvalidPaymentSecret
}

type NodeErrorInvalidAmount struct {
	message string
}

func NewNodeErrorInvalidAmount() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidAmount{},
	}
}

func (err NodeErrorInvalidAmount) Error() string {
	return fmt.Sprintf("InvalidAmount: %s", err.message)
}

func (self NodeErrorInvalidAmount) Is(target error) bool {
	return target == ErrNodeErrorInvalidAmount
}

type NodeErrorInvalidInvoice struct {
	message string
}

func NewNodeErrorInvalidInvoice() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidInvoice{},
	}
}

func (err NodeErrorInvalidInvoice) Error() string {
	return fmt.Sprintf("InvalidInvoice: %s", err.message)
}

func (self NodeErrorInvalidInvoice) Is(target error) bool {
	return target == ErrNodeErrorInvalidInvoice
}

type NodeErrorInvalidOffer struct {
	message string
}

func NewNodeErrorInvalidOffer() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidOffer{},
	}
}

func (err NodeErrorInvalidOffer) Error() string {
	return fmt.Sprintf("InvalidOffer: %s", err.message)
}

func (self NodeErrorInvalidOffer) Is(target error) bool {
	return target == ErrNodeErrorInvalidOffer
}

type NodeErrorInvalidRefund struct {
	message string
}

func NewNodeErrorInvalidRefund() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidRefund{},
	}
}

func (err NodeErrorInvalidRefund) Error() string {
	return fmt.Sprintf("InvalidRefund: %s", err.message)
}

func (self NodeErrorInvalidRefund) Is(target error) bool {
	return target == ErrNodeErrorInvalidRefund
}

type NodeErrorInvalidChannelId struct {
	message string
}

func NewNodeErrorInvalidChannelId() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidChannelId{},
	}
}

func (err NodeErrorInvalidChannelId) Error() string {
	return fmt.Sprintf("InvalidChannelId: %s", err.message)
}

func (self NodeErrorInvalidChannelId) Is(target error) bool {
	return target == ErrNodeErrorInvalidChannelId
}

type NodeErrorInvalidNetwork struct {
	message string
}

func NewNodeErrorInvalidNetwork() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidNetwork{},
	}
}

func (err NodeErrorInvalidNetwork) Error() string {
	return fmt.Sprintf("InvalidNetwork: %s", err.message)
}

func (self NodeErrorInvalidNetwork) Is(target error) bool {
	return target == ErrNodeErrorInvalidNetwork
}

type NodeErrorInvalidCustomTlv struct {
	message string
}

func NewNodeErrorInvalidCustomTlv() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidCustomTlv{},
	}
}

func (err NodeErrorInvalidCustomTlv) Error() string {
	return fmt.Sprintf("InvalidCustomTlv: %s", err.message)
}

func (self NodeErrorInvalidCustomTlv) Is(target error) bool {
	return target == ErrNodeErrorInvalidCustomTlv
}

type NodeErrorInvalidUri struct {
	message string
}

func NewNodeErrorInvalidUri() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidUri{},
	}
}

func (err NodeErrorInvalidUri) Error() string {
	return fmt.Sprintf("InvalidUri: %s", err.message)
}

func (self NodeErrorInvalidUri) Is(target error) bool {
	return target == ErrNodeErrorInvalidUri
}

type NodeErrorInvalidQuantity struct {
	message string
}

func NewNodeErrorInvalidQuantity() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidQuantity{},
	}
}

func (err NodeErrorInvalidQuantity) Error() string {
	return fmt.Sprintf("InvalidQuantity: %s", err.message)
}

func (self NodeErrorInvalidQuantity) Is(target error) bool {
	return target == ErrNodeErrorInvalidQuantity
}

type NodeErrorInvalidNodeAlias struct {
	message string
}

func NewNodeErrorInvalidNodeAlias() *NodeError {
	return &NodeError{
		err: &NodeErrorInvalidNodeAlias{},
	}
}

func (err NodeErrorInvalidNodeAlias) Error() string {
	return fmt.Sprintf("InvalidNodeAlias: %s", err.message)
}

func (self NodeErrorInvalidNodeAlias) Is(target error) bool {
	return target == ErrNodeErrorInvalidNodeAlias
}

type NodeErrorDuplicatePayment struct {
	message string
}

func NewNodeErrorDuplicatePayment() *NodeError {
	return &NodeError{
		err: &NodeErrorDuplicatePayment{},
	}
}

func (err NodeErrorDuplicatePayment) Error() string {
	return fmt.Sprintf("DuplicatePayment: %s", err.message)
}

func (self NodeErrorDuplicatePayment) Is(target error) bool {
	return target == ErrNodeErrorDuplicatePayment
}

type NodeErrorUnsupportedCurrency struct {
	message string
}

func NewNodeErrorUnsupportedCurrency() *NodeError {
	return &NodeError{
		err: &NodeErrorUnsupportedCurrency{},
	}
}

func (err NodeErrorUnsupportedCurrency) Error() string {
	return fmt.Sprintf("UnsupportedCurrency: %s", err.message)
}

func (self NodeErrorUnsupportedCurrency) Is(target error) bool {
	return target == ErrNodeErrorUnsupportedCurrency
}

type NodeErrorInsufficientFunds struct {
	message string
}

func NewNodeErrorInsufficientFunds() *NodeError {
	return &NodeError{
		err: &NodeErrorInsufficientFunds{},
	}
}

func (err NodeErrorInsufficientFunds) Error() string {
	return fmt.Sprintf("InsufficientFunds: %s", err.message)
}

func (self NodeErrorInsufficientFunds) Is(target error) bool {
	return target == ErrNodeErrorInsufficientFunds
}

type NodeErrorLiquiditySourceUnavailable struct {
	message string
}

func NewNodeErrorLiquiditySourceUnavailable() *NodeError {
	return &NodeError{
		err: &NodeErrorLiquiditySourceUnavailable{},
	}
}

func (err NodeErrorLiquiditySourceUnavailable) Error() string {
	return fmt.Sprintf("LiquiditySourceUnavailable: %s", err.message)
}

func (self NodeErrorLiquiditySourceUnavailable) Is(target error) bool {
	return target == ErrNodeErrorLiquiditySourceUnavailable
}

type NodeErrorLiquidityFeeTooHigh struct {
	message string
}

func NewNodeErrorLiquidityFeeTooHigh() *NodeError {
	return &NodeError{
		err: &NodeErrorLiquidityFeeTooHigh{},
	}
}

func (err NodeErrorLiquidityFeeTooHigh) Error() string {
	return fmt.Sprintf("LiquidityFeeTooHigh: %s", err.message)
}

func (self NodeErrorLiquidityFeeTooHigh) Is(target error) bool {
	return target == ErrNodeErrorLiquidityFeeTooHigh
}

type FfiConverterTypeNodeError struct{}

var FfiConverterTypeNodeErrorINSTANCE = FfiConverterTypeNodeError{}

func (c FfiConverterTypeNodeError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeNodeError) Lower(value *NodeError) RustBuffer {
	return LowerIntoRustBuffer[*NodeError](c, value)
}

func (c FfiConverterTypeNodeError) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	message := FfiConverterStringINSTANCE.Read(reader)
	switch errorID {
	case 1:
		return &NodeError{&NodeErrorAlreadyRunning{message}}
	case 2:
		return &NodeError{&NodeErrorNotRunning{message}}
	case 3:
		return &NodeError{&NodeErrorOnchainTxCreationFailed{message}}
	case 4:
		return &NodeError{&NodeErrorConnectionFailed{message}}
	case 5:
		return &NodeError{&NodeErrorInvoiceCreationFailed{message}}
	case 6:
		return &NodeError{&NodeErrorInvoiceRequestCreationFailed{message}}
	case 7:
		return &NodeError{&NodeErrorOfferCreationFailed{message}}
	case 8:
		return &NodeError{&NodeErrorRefundCreationFailed{message}}
	case 9:
		return &NodeError{&NodeErrorPaymentSendingFailed{message}}
	case 10:
		return &NodeError{&NodeErrorProbeSendingFailed{message}}
	case 11:
		return &NodeError{&NodeErrorChannelCreationFailed{message}}
	case 12:
		return &NodeError{&NodeErrorChannelClosingFailed{message}}
	case 13:
		return &NodeError{&NodeErrorChannelConfigUpdateFailed{message}}
	case 14:
		return &NodeError{&NodeErrorPersistenceFailed{message}}
	case 15:
		return &NodeError{&NodeErrorFeerateEstimationUpdateFailed{message}}
	case 16:
		return &NodeError{&NodeErrorFeerateEstimationUpdateTimeout{message}}
	case 17:
		return &NodeError{&NodeErrorWalletOperationFailed{message}}
	case 18:
		return &NodeError{&NodeErrorWalletOperationTimeout{message}}
	case 19:
		return &NodeError{&NodeErrorOnchainTxSigningFailed{message}}
	case 20:
		return &NodeError{&NodeErrorTxSyncFailed{message}}
	case 21:
		return &NodeError{&NodeErrorTxSyncTimeout{message}}
	case 22:
		return &NodeError{&NodeErrorGossipUpdateFailed{message}}
	case 23:
		return &NodeError{&NodeErrorGossipUpdateTimeout{message}}
	case 24:
		return &NodeError{&NodeErrorLiquidityRequestFailed{message}}
	case 25:
		return &NodeError{&NodeErrorUriParameterParsingFailed{message}}
	case 26:
		return &NodeError{&NodeErrorInvalidAddress{message}}
	case 27:
		return &NodeError{&NodeErrorInvalidSocketAddress{message}}
	case 28:
		return &NodeError{&NodeErrorInvalidPublicKey{message}}
	case 29:
		return &NodeError{&NodeErrorInvalidSecretKey{message}}
	case 30:
		return &NodeError{&NodeErrorInvalidOfferId{message}}
	case 31:
		return &NodeError{&NodeErrorInvalidNodeId{message}}
	case 32:
		return &NodeError{&NodeErrorInvalidPaymentId{message}}
	case 33:
		return &NodeError{&NodeErrorInvalidPaymentHash{message}}
	case 34:
		return &NodeError{&NodeErrorInvalidPaymentPreimage{message}}
	case 35:
		return &NodeError{&NodeErrorInvalidPaymentSecret{message}}
	case 36:
		return &NodeError{&NodeErrorInvalidAmount{message}}
	case 37:
		return &NodeError{&NodeErrorInvalidInvoice{message}}
	case 38:
		return &NodeError{&NodeErrorInvalidOffer{message}}
	case 39:
		return &NodeError{&NodeErrorInvalidRefund{message}}
	case 40:
		return &NodeError{&NodeErrorInvalidChannelId{message}}
	case 41:
		return &NodeError{&NodeErrorInvalidNetwork{message}}
	case 42:
		return &NodeError{&NodeErrorInvalidCustomTlv{message}}
	case 43:
		return &NodeError{&NodeErrorInvalidUri{message}}
	case 44:
		return &NodeError{&NodeErrorInvalidQuantity{message}}
	case 45:
		return &NodeError{&NodeErrorInvalidNodeAlias{message}}
	case 46:
		return &NodeError{&NodeErrorDuplicatePayment{message}}
	case 47:
		return &NodeError{&NodeErrorUnsupportedCurrency{message}}
	case 48:
		return &NodeError{&NodeErrorInsufficientFunds{message}}
	case 49:
		return &NodeError{&NodeErrorLiquiditySourceUnavailable{message}}
	case 50:
		return &NodeError{&NodeErrorLiquidityFeeTooHigh{message}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeNodeError.Read()", errorID))
	}

}

func (c FfiConverterTypeNodeError) Write(writer io.Writer, value *NodeError) {
	switch variantValue := value.err.(type) {
	case *NodeErrorAlreadyRunning:
		writeInt32(writer, 1)
	case *NodeErrorNotRunning:
		writeInt32(writer, 2)
	case *NodeErrorOnchainTxCreationFailed:
		writeInt32(writer, 3)
	case *NodeErrorConnectionFailed:
		writeInt32(writer, 4)
	case *NodeErrorInvoiceCreationFailed:
		writeInt32(writer, 5)
	case *NodeErrorInvoiceRequestCreationFailed:
		writeInt32(writer, 6)
	case *NodeErrorOfferCreationFailed:
		writeInt32(writer, 7)
	case *NodeErrorRefundCreationFailed:
		writeInt32(writer, 8)
	case *NodeErrorPaymentSendingFailed:
		writeInt32(writer, 9)
	case *NodeErrorProbeSendingFailed:
		writeInt32(writer, 10)
	case *NodeErrorChannelCreationFailed:
		writeInt32(writer, 11)
	case *NodeErrorChannelClosingFailed:
		writeInt32(writer, 12)
	case *NodeErrorChannelConfigUpdateFailed:
		writeInt32(writer, 13)
	case *NodeErrorPersistenceFailed:
		writeInt32(writer, 14)
	case *NodeErrorFeerateEstimationUpdateFailed:
		writeInt32(writer, 15)
	case *NodeErrorFeerateEstimationUpdateTimeout:
		writeInt32(writer, 16)
	case *NodeErrorWalletOperationFailed:
		writeInt32(writer, 17)
	case *NodeErrorWalletOperationTimeout:
		writeInt32(writer, 18)
	case *NodeErrorOnchainTxSigningFailed:
		writeInt32(writer, 19)
	case *NodeErrorTxSyncFailed:
		writeInt32(writer, 20)
	case *NodeErrorTxSyncTimeout:
		writeInt32(writer, 21)
	case *NodeErrorGossipUpdateFailed:
		writeInt32(writer, 22)
	case *NodeErrorGossipUpdateTimeout:
		writeInt32(writer, 23)
	case *NodeErrorLiquidityRequestFailed:
		writeInt32(writer, 24)
	case *NodeErrorUriParameterParsingFailed:
		writeInt32(writer, 25)
	case *NodeErrorInvalidAddress:
		writeInt32(writer, 26)
	case *NodeErrorInvalidSocketAddress:
		writeInt32(writer, 27)
	case *NodeErrorInvalidPublicKey:
		writeInt32(writer, 28)
	case *NodeErrorInvalidSecretKey:
		writeInt32(writer, 29)
	case *NodeErrorInvalidOfferId:
		writeInt32(writer, 30)
	case *NodeErrorInvalidNodeId:
		writeInt32(writer, 31)
	case *NodeErrorInvalidPaymentId:
		writeInt32(writer, 32)
	case *NodeErrorInvalidPaymentHash:
		writeInt32(writer, 33)
	case *NodeErrorInvalidPaymentPreimage:
		writeInt32(writer, 34)
	case *NodeErrorInvalidPaymentSecret:
		writeInt32(writer, 35)
	case *NodeErrorInvalidAmount:
		writeInt32(writer, 36)
	case *NodeErrorInvalidInvoice:
		writeInt32(writer, 37)
	case *NodeErrorInvalidOffer:
		writeInt32(writer, 38)
	case *NodeErrorInvalidRefund:
		writeInt32(writer, 39)
	case *NodeErrorInvalidChannelId:
		writeInt32(writer, 40)
	case *NodeErrorInvalidNetwork:
		writeInt32(writer, 41)
	case *NodeErrorInvalidCustomTlv:
		writeInt32(writer, 42)
	case *NodeErrorInvalidUri:
		writeInt32(writer, 43)
	case *NodeErrorInvalidQuantity:
		writeInt32(writer, 44)
	case *NodeErrorInvalidNodeAlias:
		writeInt32(writer, 45)
	case *NodeErrorDuplicatePayment:
		writeInt32(writer, 46)
	case *NodeErrorUnsupportedCurrency:
		writeInt32(writer, 47)
	case *NodeErrorInsufficientFunds:
		writeInt32(writer, 48)
	case *NodeErrorLiquiditySourceUnavailable:
		writeInt32(writer, 49)
	case *NodeErrorLiquidityFeeTooHigh:
		writeInt32(writer, 50)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeNodeError.Write", value))
	}
}

type PaymentDirection uint

const (
	PaymentDirectionInbound  PaymentDirection = 1
	PaymentDirectionOutbound PaymentDirection = 2
)

type FfiConverterTypePaymentDirection struct{}

var FfiConverterTypePaymentDirectionINSTANCE = FfiConverterTypePaymentDirection{}

func (c FfiConverterTypePaymentDirection) Lift(rb RustBufferI) PaymentDirection {
	return LiftFromRustBuffer[PaymentDirection](c, rb)
}

func (c FfiConverterTypePaymentDirection) Lower(value PaymentDirection) RustBuffer {
	return LowerIntoRustBuffer[PaymentDirection](c, value)
}
func (FfiConverterTypePaymentDirection) Read(reader io.Reader) PaymentDirection {
	id := readInt32(reader)
	return PaymentDirection(id)
}

func (FfiConverterTypePaymentDirection) Write(writer io.Writer, value PaymentDirection) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypePaymentDirection struct{}

func (_ FfiDestroyerTypePaymentDirection) Destroy(value PaymentDirection) {
}

type PaymentFailureReason uint

const (
	PaymentFailureReasonRecipientRejected       PaymentFailureReason = 1
	PaymentFailureReasonUserAbandoned           PaymentFailureReason = 2
	PaymentFailureReasonRetriesExhausted        PaymentFailureReason = 3
	PaymentFailureReasonPaymentExpired          PaymentFailureReason = 4
	PaymentFailureReasonRouteNotFound           PaymentFailureReason = 5
	PaymentFailureReasonUnexpectedError         PaymentFailureReason = 6
	PaymentFailureReasonUnknownRequiredFeatures PaymentFailureReason = 7
	PaymentFailureReasonInvoiceRequestExpired   PaymentFailureReason = 8
	PaymentFailureReasonInvoiceRequestRejected  PaymentFailureReason = 9
)

type FfiConverterTypePaymentFailureReason struct{}

var FfiConverterTypePaymentFailureReasonINSTANCE = FfiConverterTypePaymentFailureReason{}

func (c FfiConverterTypePaymentFailureReason) Lift(rb RustBufferI) PaymentFailureReason {
	return LiftFromRustBuffer[PaymentFailureReason](c, rb)
}

func (c FfiConverterTypePaymentFailureReason) Lower(value PaymentFailureReason) RustBuffer {
	return LowerIntoRustBuffer[PaymentFailureReason](c, value)
}
func (FfiConverterTypePaymentFailureReason) Read(reader io.Reader) PaymentFailureReason {
	id := readInt32(reader)
	return PaymentFailureReason(id)
}

func (FfiConverterTypePaymentFailureReason) Write(writer io.Writer, value PaymentFailureReason) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypePaymentFailureReason struct{}

func (_ FfiDestroyerTypePaymentFailureReason) Destroy(value PaymentFailureReason) {
}

type PaymentKind interface {
	Destroy()
}
type PaymentKindOnchain struct {
}

func (e PaymentKindOnchain) Destroy() {
}

type PaymentKindBolt11 struct {
	Hash          PaymentHash
	Preimage      *PaymentPreimage
	Secret        *PaymentSecret
	Bolt11Invoice *string
}

func (e PaymentKindBolt11) Destroy() {
	FfiDestroyerTypePaymentHash{}.Destroy(e.Hash)
	FfiDestroyerOptionalTypePaymentPreimage{}.Destroy(e.Preimage)
	FfiDestroyerOptionalTypePaymentSecret{}.Destroy(e.Secret)
	FfiDestroyerOptionalString{}.Destroy(e.Bolt11Invoice)
}

type PaymentKindBolt11Jit struct {
	Hash         PaymentHash
	Preimage     *PaymentPreimage
	Secret       *PaymentSecret
	LspFeeLimits LspFeeLimits
}

func (e PaymentKindBolt11Jit) Destroy() {
	FfiDestroyerTypePaymentHash{}.Destroy(e.Hash)
	FfiDestroyerOptionalTypePaymentPreimage{}.Destroy(e.Preimage)
	FfiDestroyerOptionalTypePaymentSecret{}.Destroy(e.Secret)
	FfiDestroyerTypeLspFeeLimits{}.Destroy(e.LspFeeLimits)
}

type PaymentKindBolt12Offer struct {
	Hash      *PaymentHash
	Preimage  *PaymentPreimage
	Secret    *PaymentSecret
	OfferId   OfferId
	PayerNote *UntrustedString
	Quantity  *uint64
}

func (e PaymentKindBolt12Offer) Destroy() {
	FfiDestroyerOptionalTypePaymentHash{}.Destroy(e.Hash)
	FfiDestroyerOptionalTypePaymentPreimage{}.Destroy(e.Preimage)
	FfiDestroyerOptionalTypePaymentSecret{}.Destroy(e.Secret)
	FfiDestroyerTypeOfferId{}.Destroy(e.OfferId)
	FfiDestroyerOptionalTypeUntrustedString{}.Destroy(e.PayerNote)
	FfiDestroyerOptionalUint64{}.Destroy(e.Quantity)
}

type PaymentKindBolt12Refund struct {
	Hash      *PaymentHash
	Preimage  *PaymentPreimage
	Secret    *PaymentSecret
	PayerNote *UntrustedString
	Quantity  *uint64
}

func (e PaymentKindBolt12Refund) Destroy() {
	FfiDestroyerOptionalTypePaymentHash{}.Destroy(e.Hash)
	FfiDestroyerOptionalTypePaymentPreimage{}.Destroy(e.Preimage)
	FfiDestroyerOptionalTypePaymentSecret{}.Destroy(e.Secret)
	FfiDestroyerOptionalTypeUntrustedString{}.Destroy(e.PayerNote)
	FfiDestroyerOptionalUint64{}.Destroy(e.Quantity)
}

type PaymentKindSpontaneous struct {
	Hash       PaymentHash
	Preimage   *PaymentPreimage
	CustomTlvs []TlvEntry
}

func (e PaymentKindSpontaneous) Destroy() {
	FfiDestroyerTypePaymentHash{}.Destroy(e.Hash)
	FfiDestroyerOptionalTypePaymentPreimage{}.Destroy(e.Preimage)
	FfiDestroyerSequenceTypeTlvEntry{}.Destroy(e.CustomTlvs)
}

type FfiConverterTypePaymentKind struct{}

var FfiConverterTypePaymentKindINSTANCE = FfiConverterTypePaymentKind{}

func (c FfiConverterTypePaymentKind) Lift(rb RustBufferI) PaymentKind {
	return LiftFromRustBuffer[PaymentKind](c, rb)
}

func (c FfiConverterTypePaymentKind) Lower(value PaymentKind) RustBuffer {
	return LowerIntoRustBuffer[PaymentKind](c, value)
}
func (FfiConverterTypePaymentKind) Read(reader io.Reader) PaymentKind {
	id := readInt32(reader)
	switch id {
	case 1:
		return PaymentKindOnchain{}
	case 2:
		return PaymentKindBolt11{
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentPreimageINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentSecretINSTANCE.Read(reader),
			FfiConverterOptionalStringINSTANCE.Read(reader),
		}
	case 3:
		return PaymentKindBolt11Jit{
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentPreimageINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentSecretINSTANCE.Read(reader),
			FfiConverterTypeLSPFeeLimitsINSTANCE.Read(reader),
		}
	case 4:
		return PaymentKindBolt12Offer{
			FfiConverterOptionalTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentPreimageINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentSecretINSTANCE.Read(reader),
			FfiConverterTypeOfferIdINSTANCE.Read(reader),
			FfiConverterOptionalTypeUntrustedStringINSTANCE.Read(reader),
			FfiConverterOptionalUint64INSTANCE.Read(reader),
		}
	case 5:
		return PaymentKindBolt12Refund{
			FfiConverterOptionalTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentPreimageINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentSecretINSTANCE.Read(reader),
			FfiConverterOptionalTypeUntrustedStringINSTANCE.Read(reader),
			FfiConverterOptionalUint64INSTANCE.Read(reader),
		}
	case 6:
		return PaymentKindSpontaneous{
			FfiConverterTypePaymentHashINSTANCE.Read(reader),
			FfiConverterOptionalTypePaymentPreimageINSTANCE.Read(reader),
			FfiConverterSequenceTypeTlvEntryINSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypePaymentKind.Read()", id))
	}
}

func (FfiConverterTypePaymentKind) Write(writer io.Writer, value PaymentKind) {
	switch variant_value := value.(type) {
	case PaymentKindOnchain:
		writeInt32(writer, 1)
	case PaymentKindBolt11:
		writeInt32(writer, 2)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.Hash)
		FfiConverterOptionalTypePaymentPreimageINSTANCE.Write(writer, variant_value.Preimage)
		FfiConverterOptionalTypePaymentSecretINSTANCE.Write(writer, variant_value.Secret)
		FfiConverterOptionalStringINSTANCE.Write(writer, variant_value.Bolt11Invoice)
	case PaymentKindBolt11Jit:
		writeInt32(writer, 3)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.Hash)
		FfiConverterOptionalTypePaymentPreimageINSTANCE.Write(writer, variant_value.Preimage)
		FfiConverterOptionalTypePaymentSecretINSTANCE.Write(writer, variant_value.Secret)
		FfiConverterTypeLSPFeeLimitsINSTANCE.Write(writer, variant_value.LspFeeLimits)
	case PaymentKindBolt12Offer:
		writeInt32(writer, 4)
		FfiConverterOptionalTypePaymentHashINSTANCE.Write(writer, variant_value.Hash)
		FfiConverterOptionalTypePaymentPreimageINSTANCE.Write(writer, variant_value.Preimage)
		FfiConverterOptionalTypePaymentSecretINSTANCE.Write(writer, variant_value.Secret)
		FfiConverterTypeOfferIdINSTANCE.Write(writer, variant_value.OfferId)
		FfiConverterOptionalTypeUntrustedStringINSTANCE.Write(writer, variant_value.PayerNote)
		FfiConverterOptionalUint64INSTANCE.Write(writer, variant_value.Quantity)
	case PaymentKindBolt12Refund:
		writeInt32(writer, 5)
		FfiConverterOptionalTypePaymentHashINSTANCE.Write(writer, variant_value.Hash)
		FfiConverterOptionalTypePaymentPreimageINSTANCE.Write(writer, variant_value.Preimage)
		FfiConverterOptionalTypePaymentSecretINSTANCE.Write(writer, variant_value.Secret)
		FfiConverterOptionalTypeUntrustedStringINSTANCE.Write(writer, variant_value.PayerNote)
		FfiConverterOptionalUint64INSTANCE.Write(writer, variant_value.Quantity)
	case PaymentKindSpontaneous:
		writeInt32(writer, 6)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, variant_value.Hash)
		FfiConverterOptionalTypePaymentPreimageINSTANCE.Write(writer, variant_value.Preimage)
		FfiConverterSequenceTypeTlvEntryINSTANCE.Write(writer, variant_value.CustomTlvs)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypePaymentKind.Write", value))
	}
}

type FfiDestroyerTypePaymentKind struct{}

func (_ FfiDestroyerTypePaymentKind) Destroy(value PaymentKind) {
	value.Destroy()
}

type PaymentStatus uint

const (
	PaymentStatusPending   PaymentStatus = 1
	PaymentStatusSucceeded PaymentStatus = 2
	PaymentStatusFailed    PaymentStatus = 3
)

type FfiConverterTypePaymentStatus struct{}

var FfiConverterTypePaymentStatusINSTANCE = FfiConverterTypePaymentStatus{}

func (c FfiConverterTypePaymentStatus) Lift(rb RustBufferI) PaymentStatus {
	return LiftFromRustBuffer[PaymentStatus](c, rb)
}

func (c FfiConverterTypePaymentStatus) Lower(value PaymentStatus) RustBuffer {
	return LowerIntoRustBuffer[PaymentStatus](c, value)
}
func (FfiConverterTypePaymentStatus) Read(reader io.Reader) PaymentStatus {
	id := readInt32(reader)
	return PaymentStatus(id)
}

func (FfiConverterTypePaymentStatus) Write(writer io.Writer, value PaymentStatus) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypePaymentStatus struct{}

func (_ FfiDestroyerTypePaymentStatus) Destroy(value PaymentStatus) {
}

type PendingSweepBalance interface {
	Destroy()
}
type PendingSweepBalancePendingBroadcast struct {
	ChannelId          *ChannelId
	AmountSatoshis     uint64
	CounterpartyNodeId *PublicKey
	FundingTxId        *Txid
	FundingTxIndex     *uint16
}

func (e PendingSweepBalancePendingBroadcast) Destroy() {
	FfiDestroyerOptionalTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerOptionalTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerOptionalTypeTxid{}.Destroy(e.FundingTxId)
	FfiDestroyerOptionalUint16{}.Destroy(e.FundingTxIndex)
}

type PendingSweepBalanceBroadcastAwaitingConfirmation struct {
	ChannelId             *ChannelId
	LatestBroadcastHeight uint32
	LatestSpendingTxid    Txid
	AmountSatoshis        uint64
	CounterpartyNodeId    *PublicKey
	FundingTxId           *Txid
	FundingTxIndex        *uint16
}

func (e PendingSweepBalanceBroadcastAwaitingConfirmation) Destroy() {
	FfiDestroyerOptionalTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerUint32{}.Destroy(e.LatestBroadcastHeight)
	FfiDestroyerTypeTxid{}.Destroy(e.LatestSpendingTxid)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerOptionalTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerOptionalTypeTxid{}.Destroy(e.FundingTxId)
	FfiDestroyerOptionalUint16{}.Destroy(e.FundingTxIndex)
}

type PendingSweepBalanceAwaitingThresholdConfirmations struct {
	ChannelId          *ChannelId
	LatestSpendingTxid Txid
	ConfirmationHash   BlockHash
	ConfirmationHeight uint32
	AmountSatoshis     uint64
	CounterpartyNodeId *PublicKey
	FundingTxId        *Txid
	FundingTxIndex     *uint16
}

func (e PendingSweepBalanceAwaitingThresholdConfirmations) Destroy() {
	FfiDestroyerOptionalTypeChannelId{}.Destroy(e.ChannelId)
	FfiDestroyerTypeTxid{}.Destroy(e.LatestSpendingTxid)
	FfiDestroyerTypeBlockHash{}.Destroy(e.ConfirmationHash)
	FfiDestroyerUint32{}.Destroy(e.ConfirmationHeight)
	FfiDestroyerUint64{}.Destroy(e.AmountSatoshis)
	FfiDestroyerOptionalTypePublicKey{}.Destroy(e.CounterpartyNodeId)
	FfiDestroyerOptionalTypeTxid{}.Destroy(e.FundingTxId)
	FfiDestroyerOptionalUint16{}.Destroy(e.FundingTxIndex)
}

type FfiConverterTypePendingSweepBalance struct{}

var FfiConverterTypePendingSweepBalanceINSTANCE = FfiConverterTypePendingSweepBalance{}

func (c FfiConverterTypePendingSweepBalance) Lift(rb RustBufferI) PendingSweepBalance {
	return LiftFromRustBuffer[PendingSweepBalance](c, rb)
}

func (c FfiConverterTypePendingSweepBalance) Lower(value PendingSweepBalance) RustBuffer {
	return LowerIntoRustBuffer[PendingSweepBalance](c, value)
}
func (FfiConverterTypePendingSweepBalance) Read(reader io.Reader) PendingSweepBalance {
	id := readInt32(reader)
	switch id {
	case 1:
		return PendingSweepBalancePendingBroadcast{
			FfiConverterOptionalTypeChannelIdINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterOptionalTypePublicKeyINSTANCE.Read(reader),
			FfiConverterOptionalTypeTxidINSTANCE.Read(reader),
			FfiConverterOptionalUint16INSTANCE.Read(reader),
		}
	case 2:
		return PendingSweepBalanceBroadcastAwaitingConfirmation{
			FfiConverterOptionalTypeChannelIdINSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterTypeTxidINSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterOptionalTypePublicKeyINSTANCE.Read(reader),
			FfiConverterOptionalTypeTxidINSTANCE.Read(reader),
			FfiConverterOptionalUint16INSTANCE.Read(reader),
		}
	case 3:
		return PendingSweepBalanceAwaitingThresholdConfirmations{
			FfiConverterOptionalTypeChannelIdINSTANCE.Read(reader),
			FfiConverterTypeTxidINSTANCE.Read(reader),
			FfiConverterTypeBlockHashINSTANCE.Read(reader),
			FfiConverterUint32INSTANCE.Read(reader),
			FfiConverterUint64INSTANCE.Read(reader),
			FfiConverterOptionalTypePublicKeyINSTANCE.Read(reader),
			FfiConverterOptionalTypeTxidINSTANCE.Read(reader),
			FfiConverterOptionalUint16INSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypePendingSweepBalance.Read()", id))
	}
}

func (FfiConverterTypePendingSweepBalance) Write(writer io.Writer, value PendingSweepBalance) {
	switch variant_value := value.(type) {
	case PendingSweepBalancePendingBroadcast:
		writeInt32(writer, 1)
		FfiConverterOptionalTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterOptionalTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterOptionalTypeTxidINSTANCE.Write(writer, variant_value.FundingTxId)
		FfiConverterOptionalUint16INSTANCE.Write(writer, variant_value.FundingTxIndex)
	case PendingSweepBalanceBroadcastAwaitingConfirmation:
		writeInt32(writer, 2)
		FfiConverterOptionalTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.LatestBroadcastHeight)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.LatestSpendingTxid)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterOptionalTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterOptionalTypeTxidINSTANCE.Write(writer, variant_value.FundingTxId)
		FfiConverterOptionalUint16INSTANCE.Write(writer, variant_value.FundingTxIndex)
	case PendingSweepBalanceAwaitingThresholdConfirmations:
		writeInt32(writer, 3)
		FfiConverterOptionalTypeChannelIdINSTANCE.Write(writer, variant_value.ChannelId)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.LatestSpendingTxid)
		FfiConverterTypeBlockHashINSTANCE.Write(writer, variant_value.ConfirmationHash)
		FfiConverterUint32INSTANCE.Write(writer, variant_value.ConfirmationHeight)
		FfiConverterUint64INSTANCE.Write(writer, variant_value.AmountSatoshis)
		FfiConverterOptionalTypePublicKeyINSTANCE.Write(writer, variant_value.CounterpartyNodeId)
		FfiConverterOptionalTypeTxidINSTANCE.Write(writer, variant_value.FundingTxId)
		FfiConverterOptionalUint16INSTANCE.Write(writer, variant_value.FundingTxIndex)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypePendingSweepBalance.Write", value))
	}
}

type FfiDestroyerTypePendingSweepBalance struct{}

func (_ FfiDestroyerTypePendingSweepBalance) Destroy(value PendingSweepBalance) {
	value.Destroy()
}

type QrPaymentResult interface {
	Destroy()
}
type QrPaymentResultOnchain struct {
	Txid Txid
}

func (e QrPaymentResultOnchain) Destroy() {
	FfiDestroyerTypeTxid{}.Destroy(e.Txid)
}

type QrPaymentResultBolt11 struct {
	PaymentId PaymentId
}

func (e QrPaymentResultBolt11) Destroy() {
	FfiDestroyerTypePaymentId{}.Destroy(e.PaymentId)
}

type QrPaymentResultBolt12 struct {
	PaymentId PaymentId
}

func (e QrPaymentResultBolt12) Destroy() {
	FfiDestroyerTypePaymentId{}.Destroy(e.PaymentId)
}

type FfiConverterTypeQrPaymentResult struct{}

var FfiConverterTypeQrPaymentResultINSTANCE = FfiConverterTypeQrPaymentResult{}

func (c FfiConverterTypeQrPaymentResult) Lift(rb RustBufferI) QrPaymentResult {
	return LiftFromRustBuffer[QrPaymentResult](c, rb)
}

func (c FfiConverterTypeQrPaymentResult) Lower(value QrPaymentResult) RustBuffer {
	return LowerIntoRustBuffer[QrPaymentResult](c, value)
}
func (FfiConverterTypeQrPaymentResult) Read(reader io.Reader) QrPaymentResult {
	id := readInt32(reader)
	switch id {
	case 1:
		return QrPaymentResultOnchain{
			FfiConverterTypeTxidINSTANCE.Read(reader),
		}
	case 2:
		return QrPaymentResultBolt11{
			FfiConverterTypePaymentIdINSTANCE.Read(reader),
		}
	case 3:
		return QrPaymentResultBolt12{
			FfiConverterTypePaymentIdINSTANCE.Read(reader),
		}
	default:
		panic(fmt.Sprintf("invalid enum value %v in FfiConverterTypeQrPaymentResult.Read()", id))
	}
}

func (FfiConverterTypeQrPaymentResult) Write(writer io.Writer, value QrPaymentResult) {
	switch variant_value := value.(type) {
	case QrPaymentResultOnchain:
		writeInt32(writer, 1)
		FfiConverterTypeTxidINSTANCE.Write(writer, variant_value.Txid)
	case QrPaymentResultBolt11:
		writeInt32(writer, 2)
		FfiConverterTypePaymentIdINSTANCE.Write(writer, variant_value.PaymentId)
	case QrPaymentResultBolt12:
		writeInt32(writer, 3)
		FfiConverterTypePaymentIdINSTANCE.Write(writer, variant_value.PaymentId)
	default:
		_ = variant_value
		panic(fmt.Sprintf("invalid enum value `%v` in FfiConverterTypeQrPaymentResult.Write", value))
	}
}

type FfiDestroyerTypeQrPaymentResult struct{}

func (_ FfiDestroyerTypeQrPaymentResult) Destroy(value QrPaymentResult) {
	value.Destroy()
}

type ResetState uint

const (
	ResetStateNodeMetrics  ResetState = 1
	ResetStateScorer       ResetState = 2
	ResetStateNetworkGraph ResetState = 3
	ResetStateAll          ResetState = 4
)

type FfiConverterTypeResetState struct{}

var FfiConverterTypeResetStateINSTANCE = FfiConverterTypeResetState{}

func (c FfiConverterTypeResetState) Lift(rb RustBufferI) ResetState {
	return LiftFromRustBuffer[ResetState](c, rb)
}

func (c FfiConverterTypeResetState) Lower(value ResetState) RustBuffer {
	return LowerIntoRustBuffer[ResetState](c, value)
}
func (FfiConverterTypeResetState) Read(reader io.Reader) ResetState {
	id := readInt32(reader)
	return ResetState(id)
}

func (FfiConverterTypeResetState) Write(writer io.Writer, value ResetState) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerTypeResetState struct{}

func (_ FfiDestroyerTypeResetState) Destroy(value ResetState) {
}

type VssHeaderProviderError struct {
	err error
}

func (err VssHeaderProviderError) Error() string {
	return fmt.Sprintf("VssHeaderProviderError: %s", err.err.Error())
}

func (err VssHeaderProviderError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrVssHeaderProviderErrorInvalidData = fmt.Errorf("VssHeaderProviderErrorInvalidData")
var ErrVssHeaderProviderErrorRequestError = fmt.Errorf("VssHeaderProviderErrorRequestError")
var ErrVssHeaderProviderErrorAuthorizationError = fmt.Errorf("VssHeaderProviderErrorAuthorizationError")
var ErrVssHeaderProviderErrorInternalError = fmt.Errorf("VssHeaderProviderErrorInternalError")

// Variant structs
type VssHeaderProviderErrorInvalidData struct {
	message string
}

func NewVssHeaderProviderErrorInvalidData() *VssHeaderProviderError {
	return &VssHeaderProviderError{
		err: &VssHeaderProviderErrorInvalidData{},
	}
}

func (err VssHeaderProviderErrorInvalidData) Error() string {
	return fmt.Sprintf("InvalidData: %s", err.message)
}

func (self VssHeaderProviderErrorInvalidData) Is(target error) bool {
	return target == ErrVssHeaderProviderErrorInvalidData
}

type VssHeaderProviderErrorRequestError struct {
	message string
}

func NewVssHeaderProviderErrorRequestError() *VssHeaderProviderError {
	return &VssHeaderProviderError{
		err: &VssHeaderProviderErrorRequestError{},
	}
}

func (err VssHeaderProviderErrorRequestError) Error() string {
	return fmt.Sprintf("RequestError: %s", err.message)
}

func (self VssHeaderProviderErrorRequestError) Is(target error) bool {
	return target == ErrVssHeaderProviderErrorRequestError
}

type VssHeaderProviderErrorAuthorizationError struct {
	message string
}

func NewVssHeaderProviderErrorAuthorizationError() *VssHeaderProviderError {
	return &VssHeaderProviderError{
		err: &VssHeaderProviderErrorAuthorizationError{},
	}
}

func (err VssHeaderProviderErrorAuthorizationError) Error() string {
	return fmt.Sprintf("AuthorizationError: %s", err.message)
}

func (self VssHeaderProviderErrorAuthorizationError) Is(target error) bool {
	return target == ErrVssHeaderProviderErrorAuthorizationError
}

type VssHeaderProviderErrorInternalError struct {
	message string
}

func NewVssHeaderProviderErrorInternalError() *VssHeaderProviderError {
	return &VssHeaderProviderError{
		err: &VssHeaderProviderErrorInternalError{},
	}
}

func (err VssHeaderProviderErrorInternalError) Error() string {
	return fmt.Sprintf("InternalError: %s", err.message)
}

func (self VssHeaderProviderErrorInternalError) Is(target error) bool {
	return target == ErrVssHeaderProviderErrorInternalError
}

type FfiConverterTypeVssHeaderProviderError struct{}

var FfiConverterTypeVssHeaderProviderErrorINSTANCE = FfiConverterTypeVssHeaderProviderError{}

func (c FfiConverterTypeVssHeaderProviderError) Lift(eb RustBufferI) error {
	return LiftFromRustBuffer[error](c, eb)
}

func (c FfiConverterTypeVssHeaderProviderError) Lower(value *VssHeaderProviderError) RustBuffer {
	return LowerIntoRustBuffer[*VssHeaderProviderError](c, value)
}

func (c FfiConverterTypeVssHeaderProviderError) Read(reader io.Reader) error {
	errorID := readUint32(reader)

	message := FfiConverterStringINSTANCE.Read(reader)
	switch errorID {
	case 1:
		return &VssHeaderProviderError{&VssHeaderProviderErrorInvalidData{message}}
	case 2:
		return &VssHeaderProviderError{&VssHeaderProviderErrorRequestError{message}}
	case 3:
		return &VssHeaderProviderError{&VssHeaderProviderErrorAuthorizationError{message}}
	case 4:
		return &VssHeaderProviderError{&VssHeaderProviderErrorInternalError{message}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterTypeVssHeaderProviderError.Read()", errorID))
	}

}

func (c FfiConverterTypeVssHeaderProviderError) Write(writer io.Writer, value *VssHeaderProviderError) {
	switch variantValue := value.err.(type) {
	case *VssHeaderProviderErrorInvalidData:
		writeInt32(writer, 1)
	case *VssHeaderProviderErrorRequestError:
		writeInt32(writer, 2)
	case *VssHeaderProviderErrorAuthorizationError:
		writeInt32(writer, 3)
	case *VssHeaderProviderErrorInternalError:
		writeInt32(writer, 4)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterTypeVssHeaderProviderError.Write", value))
	}
}

type FfiConverterOptionalUint8 struct{}

var FfiConverterOptionalUint8INSTANCE = FfiConverterOptionalUint8{}

func (c FfiConverterOptionalUint8) Lift(rb RustBufferI) *uint8 {
	return LiftFromRustBuffer[*uint8](c, rb)
}

func (_ FfiConverterOptionalUint8) Read(reader io.Reader) *uint8 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint8INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint8) Lower(value *uint8) RustBuffer {
	return LowerIntoRustBuffer[*uint8](c, value)
}

func (_ FfiConverterOptionalUint8) Write(writer io.Writer, value *uint8) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint8INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint8 struct{}

func (_ FfiDestroyerOptionalUint8) Destroy(value *uint8) {
	if value != nil {
		FfiDestroyerUint8{}.Destroy(*value)
	}
}

type FfiConverterOptionalUint16 struct{}

var FfiConverterOptionalUint16INSTANCE = FfiConverterOptionalUint16{}

func (c FfiConverterOptionalUint16) Lift(rb RustBufferI) *uint16 {
	return LiftFromRustBuffer[*uint16](c, rb)
}

func (_ FfiConverterOptionalUint16) Read(reader io.Reader) *uint16 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint16INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint16) Lower(value *uint16) RustBuffer {
	return LowerIntoRustBuffer[*uint16](c, value)
}

func (_ FfiConverterOptionalUint16) Write(writer io.Writer, value *uint16) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint16INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint16 struct{}

func (_ FfiDestroyerOptionalUint16) Destroy(value *uint16) {
	if value != nil {
		FfiDestroyerUint16{}.Destroy(*value)
	}
}

type FfiConverterOptionalUint32 struct{}

var FfiConverterOptionalUint32INSTANCE = FfiConverterOptionalUint32{}

func (c FfiConverterOptionalUint32) Lift(rb RustBufferI) *uint32 {
	return LiftFromRustBuffer[*uint32](c, rb)
}

func (_ FfiConverterOptionalUint32) Read(reader io.Reader) *uint32 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint32INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint32) Lower(value *uint32) RustBuffer {
	return LowerIntoRustBuffer[*uint32](c, value)
}

func (_ FfiConverterOptionalUint32) Write(writer io.Writer, value *uint32) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint32INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint32 struct{}

func (_ FfiDestroyerOptionalUint32) Destroy(value *uint32) {
	if value != nil {
		FfiDestroyerUint32{}.Destroy(*value)
	}
}

type FfiConverterOptionalUint64 struct{}

var FfiConverterOptionalUint64INSTANCE = FfiConverterOptionalUint64{}

func (c FfiConverterOptionalUint64) Lift(rb RustBufferI) *uint64 {
	return LiftFromRustBuffer[*uint64](c, rb)
}

func (_ FfiConverterOptionalUint64) Read(reader io.Reader) *uint64 {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterUint64INSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalUint64) Lower(value *uint64) RustBuffer {
	return LowerIntoRustBuffer[*uint64](c, value)
}

func (_ FfiConverterOptionalUint64) Write(writer io.Writer, value *uint64) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterUint64INSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalUint64 struct{}

func (_ FfiDestroyerOptionalUint64) Destroy(value *uint64) {
	if value != nil {
		FfiDestroyerUint64{}.Destroy(*value)
	}
}

type FfiConverterOptionalBool struct{}

var FfiConverterOptionalBoolINSTANCE = FfiConverterOptionalBool{}

func (c FfiConverterOptionalBool) Lift(rb RustBufferI) *bool {
	return LiftFromRustBuffer[*bool](c, rb)
}

func (_ FfiConverterOptionalBool) Read(reader io.Reader) *bool {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterBoolINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalBool) Lower(value *bool) RustBuffer {
	return LowerIntoRustBuffer[*bool](c, value)
}

func (_ FfiConverterOptionalBool) Write(writer io.Writer, value *bool) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterBoolINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalBool struct{}

func (_ FfiDestroyerOptionalBool) Destroy(value *bool) {
	if value != nil {
		FfiDestroyerBool{}.Destroy(*value)
	}
}

type FfiConverterOptionalString struct{}

var FfiConverterOptionalStringINSTANCE = FfiConverterOptionalString{}

func (c FfiConverterOptionalString) Lift(rb RustBufferI) *string {
	return LiftFromRustBuffer[*string](c, rb)
}

func (_ FfiConverterOptionalString) Read(reader io.Reader) *string {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterStringINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalString) Lower(value *string) RustBuffer {
	return LowerIntoRustBuffer[*string](c, value)
}

func (_ FfiConverterOptionalString) Write(writer io.Writer, value *string) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterStringINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalString struct{}

func (_ FfiDestroyerOptionalString) Destroy(value *string) {
	if value != nil {
		FfiDestroyerString{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeAnchorChannelsConfig struct{}

var FfiConverterOptionalTypeAnchorChannelsConfigINSTANCE = FfiConverterOptionalTypeAnchorChannelsConfig{}

func (c FfiConverterOptionalTypeAnchorChannelsConfig) Lift(rb RustBufferI) *AnchorChannelsConfig {
	return LiftFromRustBuffer[*AnchorChannelsConfig](c, rb)
}

func (_ FfiConverterOptionalTypeAnchorChannelsConfig) Read(reader io.Reader) *AnchorChannelsConfig {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeAnchorChannelsConfigINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeAnchorChannelsConfig) Lower(value *AnchorChannelsConfig) RustBuffer {
	return LowerIntoRustBuffer[*AnchorChannelsConfig](c, value)
}

func (_ FfiConverterOptionalTypeAnchorChannelsConfig) Write(writer io.Writer, value *AnchorChannelsConfig) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeAnchorChannelsConfigINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeAnchorChannelsConfig struct{}

func (_ FfiDestroyerOptionalTypeAnchorChannelsConfig) Destroy(value *AnchorChannelsConfig) {
	if value != nil {
		FfiDestroyerTypeAnchorChannelsConfig{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeChannelConfig struct{}

var FfiConverterOptionalTypeChannelConfigINSTANCE = FfiConverterOptionalTypeChannelConfig{}

func (c FfiConverterOptionalTypeChannelConfig) Lift(rb RustBufferI) *ChannelConfig {
	return LiftFromRustBuffer[*ChannelConfig](c, rb)
}

func (_ FfiConverterOptionalTypeChannelConfig) Read(reader io.Reader) *ChannelConfig {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeChannelConfigINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeChannelConfig) Lower(value *ChannelConfig) RustBuffer {
	return LowerIntoRustBuffer[*ChannelConfig](c, value)
}

func (_ FfiConverterOptionalTypeChannelConfig) Write(writer io.Writer, value *ChannelConfig) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeChannelConfigINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeChannelConfig struct{}

func (_ FfiDestroyerOptionalTypeChannelConfig) Destroy(value *ChannelConfig) {
	if value != nil {
		FfiDestroyerTypeChannelConfig{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeChannelInfo struct{}

var FfiConverterOptionalTypeChannelInfoINSTANCE = FfiConverterOptionalTypeChannelInfo{}

func (c FfiConverterOptionalTypeChannelInfo) Lift(rb RustBufferI) *ChannelInfo {
	return LiftFromRustBuffer[*ChannelInfo](c, rb)
}

func (_ FfiConverterOptionalTypeChannelInfo) Read(reader io.Reader) *ChannelInfo {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeChannelInfoINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeChannelInfo) Lower(value *ChannelInfo) RustBuffer {
	return LowerIntoRustBuffer[*ChannelInfo](c, value)
}

func (_ FfiConverterOptionalTypeChannelInfo) Write(writer io.Writer, value *ChannelInfo) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeChannelInfoINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeChannelInfo struct{}

func (_ FfiDestroyerOptionalTypeChannelInfo) Destroy(value *ChannelInfo) {
	if value != nil {
		FfiDestroyerTypeChannelInfo{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeChannelUpdateInfo struct{}

var FfiConverterOptionalTypeChannelUpdateInfoINSTANCE = FfiConverterOptionalTypeChannelUpdateInfo{}

func (c FfiConverterOptionalTypeChannelUpdateInfo) Lift(rb RustBufferI) *ChannelUpdateInfo {
	return LiftFromRustBuffer[*ChannelUpdateInfo](c, rb)
}

func (_ FfiConverterOptionalTypeChannelUpdateInfo) Read(reader io.Reader) *ChannelUpdateInfo {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeChannelUpdateInfoINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeChannelUpdateInfo) Lower(value *ChannelUpdateInfo) RustBuffer {
	return LowerIntoRustBuffer[*ChannelUpdateInfo](c, value)
}

func (_ FfiConverterOptionalTypeChannelUpdateInfo) Write(writer io.Writer, value *ChannelUpdateInfo) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeChannelUpdateInfoINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeChannelUpdateInfo struct{}

func (_ FfiDestroyerOptionalTypeChannelUpdateInfo) Destroy(value *ChannelUpdateInfo) {
	if value != nil {
		FfiDestroyerTypeChannelUpdateInfo{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeEsploraSyncConfig struct{}

var FfiConverterOptionalTypeEsploraSyncConfigINSTANCE = FfiConverterOptionalTypeEsploraSyncConfig{}

func (c FfiConverterOptionalTypeEsploraSyncConfig) Lift(rb RustBufferI) *EsploraSyncConfig {
	return LiftFromRustBuffer[*EsploraSyncConfig](c, rb)
}

func (_ FfiConverterOptionalTypeEsploraSyncConfig) Read(reader io.Reader) *EsploraSyncConfig {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeEsploraSyncConfigINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeEsploraSyncConfig) Lower(value *EsploraSyncConfig) RustBuffer {
	return LowerIntoRustBuffer[*EsploraSyncConfig](c, value)
}

func (_ FfiConverterOptionalTypeEsploraSyncConfig) Write(writer io.Writer, value *EsploraSyncConfig) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeEsploraSyncConfigINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeEsploraSyncConfig struct{}

func (_ FfiDestroyerOptionalTypeEsploraSyncConfig) Destroy(value *EsploraSyncConfig) {
	if value != nil {
		FfiDestroyerTypeEsploraSyncConfig{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeNodeAnnouncementInfo struct{}

var FfiConverterOptionalTypeNodeAnnouncementInfoINSTANCE = FfiConverterOptionalTypeNodeAnnouncementInfo{}

func (c FfiConverterOptionalTypeNodeAnnouncementInfo) Lift(rb RustBufferI) *NodeAnnouncementInfo {
	return LiftFromRustBuffer[*NodeAnnouncementInfo](c, rb)
}

func (_ FfiConverterOptionalTypeNodeAnnouncementInfo) Read(reader io.Reader) *NodeAnnouncementInfo {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeNodeAnnouncementInfoINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeNodeAnnouncementInfo) Lower(value *NodeAnnouncementInfo) RustBuffer {
	return LowerIntoRustBuffer[*NodeAnnouncementInfo](c, value)
}

func (_ FfiConverterOptionalTypeNodeAnnouncementInfo) Write(writer io.Writer, value *NodeAnnouncementInfo) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeNodeAnnouncementInfoINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeNodeAnnouncementInfo struct{}

func (_ FfiDestroyerOptionalTypeNodeAnnouncementInfo) Destroy(value *NodeAnnouncementInfo) {
	if value != nil {
		FfiDestroyerTypeNodeAnnouncementInfo{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeNodeInfo struct{}

var FfiConverterOptionalTypeNodeInfoINSTANCE = FfiConverterOptionalTypeNodeInfo{}

func (c FfiConverterOptionalTypeNodeInfo) Lift(rb RustBufferI) *NodeInfo {
	return LiftFromRustBuffer[*NodeInfo](c, rb)
}

func (_ FfiConverterOptionalTypeNodeInfo) Read(reader io.Reader) *NodeInfo {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeNodeInfoINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeNodeInfo) Lower(value *NodeInfo) RustBuffer {
	return LowerIntoRustBuffer[*NodeInfo](c, value)
}

func (_ FfiConverterOptionalTypeNodeInfo) Write(writer io.Writer, value *NodeInfo) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeNodeInfoINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeNodeInfo struct{}

func (_ FfiDestroyerOptionalTypeNodeInfo) Destroy(value *NodeInfo) {
	if value != nil {
		FfiDestroyerTypeNodeInfo{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeOutPoint struct{}

var FfiConverterOptionalTypeOutPointINSTANCE = FfiConverterOptionalTypeOutPoint{}

func (c FfiConverterOptionalTypeOutPoint) Lift(rb RustBufferI) *OutPoint {
	return LiftFromRustBuffer[*OutPoint](c, rb)
}

func (_ FfiConverterOptionalTypeOutPoint) Read(reader io.Reader) *OutPoint {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeOutPointINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeOutPoint) Lower(value *OutPoint) RustBuffer {
	return LowerIntoRustBuffer[*OutPoint](c, value)
}

func (_ FfiConverterOptionalTypeOutPoint) Write(writer io.Writer, value *OutPoint) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeOutPointINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeOutPoint struct{}

func (_ FfiDestroyerOptionalTypeOutPoint) Destroy(value *OutPoint) {
	if value != nil {
		FfiDestroyerTypeOutPoint{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentDetails struct{}

var FfiConverterOptionalTypePaymentDetailsINSTANCE = FfiConverterOptionalTypePaymentDetails{}

func (c FfiConverterOptionalTypePaymentDetails) Lift(rb RustBufferI) *PaymentDetails {
	return LiftFromRustBuffer[*PaymentDetails](c, rb)
}

func (_ FfiConverterOptionalTypePaymentDetails) Read(reader io.Reader) *PaymentDetails {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentDetailsINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentDetails) Lower(value *PaymentDetails) RustBuffer {
	return LowerIntoRustBuffer[*PaymentDetails](c, value)
}

func (_ FfiConverterOptionalTypePaymentDetails) Write(writer io.Writer, value *PaymentDetails) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentDetailsINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentDetails struct{}

func (_ FfiDestroyerOptionalTypePaymentDetails) Destroy(value *PaymentDetails) {
	if value != nil {
		FfiDestroyerTypePaymentDetails{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeSendingParameters struct{}

var FfiConverterOptionalTypeSendingParametersINSTANCE = FfiConverterOptionalTypeSendingParameters{}

func (c FfiConverterOptionalTypeSendingParameters) Lift(rb RustBufferI) *SendingParameters {
	return LiftFromRustBuffer[*SendingParameters](c, rb)
}

func (_ FfiConverterOptionalTypeSendingParameters) Read(reader io.Reader) *SendingParameters {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeSendingParametersINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeSendingParameters) Lower(value *SendingParameters) RustBuffer {
	return LowerIntoRustBuffer[*SendingParameters](c, value)
}

func (_ FfiConverterOptionalTypeSendingParameters) Write(writer io.Writer, value *SendingParameters) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeSendingParametersINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeSendingParameters struct{}

func (_ FfiDestroyerOptionalTypeSendingParameters) Destroy(value *SendingParameters) {
	if value != nil {
		FfiDestroyerTypeSendingParameters{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeChannelType struct{}

var FfiConverterOptionalTypeChannelTypeINSTANCE = FfiConverterOptionalTypeChannelType{}

func (c FfiConverterOptionalTypeChannelType) Lift(rb RustBufferI) *ChannelType {
	return LiftFromRustBuffer[*ChannelType](c, rb)
}

func (_ FfiConverterOptionalTypeChannelType) Read(reader io.Reader) *ChannelType {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeChannelTypeINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeChannelType) Lower(value *ChannelType) RustBuffer {
	return LowerIntoRustBuffer[*ChannelType](c, value)
}

func (_ FfiConverterOptionalTypeChannelType) Write(writer io.Writer, value *ChannelType) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeChannelTypeINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeChannelType struct{}

func (_ FfiDestroyerOptionalTypeChannelType) Destroy(value *ChannelType) {
	if value != nil {
		FfiDestroyerTypeChannelType{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeClosureReason struct{}

var FfiConverterOptionalTypeClosureReasonINSTANCE = FfiConverterOptionalTypeClosureReason{}

func (c FfiConverterOptionalTypeClosureReason) Lift(rb RustBufferI) *ClosureReason {
	return LiftFromRustBuffer[*ClosureReason](c, rb)
}

func (_ FfiConverterOptionalTypeClosureReason) Read(reader io.Reader) *ClosureReason {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeClosureReasonINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeClosureReason) Lower(value *ClosureReason) RustBuffer {
	return LowerIntoRustBuffer[*ClosureReason](c, value)
}

func (_ FfiConverterOptionalTypeClosureReason) Write(writer io.Writer, value *ClosureReason) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeClosureReasonINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeClosureReason struct{}

func (_ FfiDestroyerOptionalTypeClosureReason) Destroy(value *ClosureReason) {
	if value != nil {
		FfiDestroyerTypeClosureReason{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeEvent struct{}

var FfiConverterOptionalTypeEventINSTANCE = FfiConverterOptionalTypeEvent{}

func (c FfiConverterOptionalTypeEvent) Lift(rb RustBufferI) *Event {
	return LiftFromRustBuffer[*Event](c, rb)
}

func (_ FfiConverterOptionalTypeEvent) Read(reader io.Reader) *Event {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeEventINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeEvent) Lower(value *Event) RustBuffer {
	return LowerIntoRustBuffer[*Event](c, value)
}

func (_ FfiConverterOptionalTypeEvent) Write(writer io.Writer, value *Event) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeEventINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeEvent struct{}

func (_ FfiDestroyerOptionalTypeEvent) Destroy(value *Event) {
	if value != nil {
		FfiDestroyerTypeEvent{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeMaxTotalRoutingFeeLimit struct{}

var FfiConverterOptionalTypeMaxTotalRoutingFeeLimitINSTANCE = FfiConverterOptionalTypeMaxTotalRoutingFeeLimit{}

func (c FfiConverterOptionalTypeMaxTotalRoutingFeeLimit) Lift(rb RustBufferI) *MaxTotalRoutingFeeLimit {
	return LiftFromRustBuffer[*MaxTotalRoutingFeeLimit](c, rb)
}

func (_ FfiConverterOptionalTypeMaxTotalRoutingFeeLimit) Read(reader io.Reader) *MaxTotalRoutingFeeLimit {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeMaxTotalRoutingFeeLimitINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeMaxTotalRoutingFeeLimit) Lower(value *MaxTotalRoutingFeeLimit) RustBuffer {
	return LowerIntoRustBuffer[*MaxTotalRoutingFeeLimit](c, value)
}

func (_ FfiConverterOptionalTypeMaxTotalRoutingFeeLimit) Write(writer io.Writer, value *MaxTotalRoutingFeeLimit) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeMaxTotalRoutingFeeLimitINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeMaxTotalRoutingFeeLimit struct{}

func (_ FfiDestroyerOptionalTypeMaxTotalRoutingFeeLimit) Destroy(value *MaxTotalRoutingFeeLimit) {
	if value != nil {
		FfiDestroyerTypeMaxTotalRoutingFeeLimit{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentFailureReason struct{}

var FfiConverterOptionalTypePaymentFailureReasonINSTANCE = FfiConverterOptionalTypePaymentFailureReason{}

func (c FfiConverterOptionalTypePaymentFailureReason) Lift(rb RustBufferI) *PaymentFailureReason {
	return LiftFromRustBuffer[*PaymentFailureReason](c, rb)
}

func (_ FfiConverterOptionalTypePaymentFailureReason) Read(reader io.Reader) *PaymentFailureReason {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentFailureReasonINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentFailureReason) Lower(value *PaymentFailureReason) RustBuffer {
	return LowerIntoRustBuffer[*PaymentFailureReason](c, value)
}

func (_ FfiConverterOptionalTypePaymentFailureReason) Write(writer io.Writer, value *PaymentFailureReason) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentFailureReasonINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentFailureReason struct{}

func (_ FfiDestroyerOptionalTypePaymentFailureReason) Destroy(value *PaymentFailureReason) {
	if value != nil {
		FfiDestroyerTypePaymentFailureReason{}.Destroy(*value)
	}
}

type FfiConverterOptionalSequenceTypeSocketAddress struct{}

var FfiConverterOptionalSequenceTypeSocketAddressINSTANCE = FfiConverterOptionalSequenceTypeSocketAddress{}

func (c FfiConverterOptionalSequenceTypeSocketAddress) Lift(rb RustBufferI) *[]SocketAddress {
	return LiftFromRustBuffer[*[]SocketAddress](c, rb)
}

func (_ FfiConverterOptionalSequenceTypeSocketAddress) Read(reader io.Reader) *[]SocketAddress {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceTypeSocketAddressINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceTypeSocketAddress) Lower(value *[]SocketAddress) RustBuffer {
	return LowerIntoRustBuffer[*[]SocketAddress](c, value)
}

func (_ FfiConverterOptionalSequenceTypeSocketAddress) Write(writer io.Writer, value *[]SocketAddress) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceTypeSocketAddressINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceTypeSocketAddress struct{}

func (_ FfiDestroyerOptionalSequenceTypeSocketAddress) Destroy(value *[]SocketAddress) {
	if value != nil {
		FfiDestroyerSequenceTypeSocketAddress{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeChannelId struct{}

var FfiConverterOptionalTypeChannelIdINSTANCE = FfiConverterOptionalTypeChannelId{}

func (c FfiConverterOptionalTypeChannelId) Lift(rb RustBufferI) *ChannelId {
	return LiftFromRustBuffer[*ChannelId](c, rb)
}

func (_ FfiConverterOptionalTypeChannelId) Read(reader io.Reader) *ChannelId {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeChannelIdINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeChannelId) Lower(value *ChannelId) RustBuffer {
	return LowerIntoRustBuffer[*ChannelId](c, value)
}

func (_ FfiConverterOptionalTypeChannelId) Write(writer io.Writer, value *ChannelId) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeChannelIdINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeChannelId struct{}

func (_ FfiDestroyerOptionalTypeChannelId) Destroy(value *ChannelId) {
	if value != nil {
		FfiDestroyerTypeChannelId{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeNodeAlias struct{}

var FfiConverterOptionalTypeNodeAliasINSTANCE = FfiConverterOptionalTypeNodeAlias{}

func (c FfiConverterOptionalTypeNodeAlias) Lift(rb RustBufferI) *NodeAlias {
	return LiftFromRustBuffer[*NodeAlias](c, rb)
}

func (_ FfiConverterOptionalTypeNodeAlias) Read(reader io.Reader) *NodeAlias {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeNodeAliasINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeNodeAlias) Lower(value *NodeAlias) RustBuffer {
	return LowerIntoRustBuffer[*NodeAlias](c, value)
}

func (_ FfiConverterOptionalTypeNodeAlias) Write(writer io.Writer, value *NodeAlias) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeNodeAliasINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeNodeAlias struct{}

func (_ FfiDestroyerOptionalTypeNodeAlias) Destroy(value *NodeAlias) {
	if value != nil {
		FfiDestroyerTypeNodeAlias{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentHash struct{}

var FfiConverterOptionalTypePaymentHashINSTANCE = FfiConverterOptionalTypePaymentHash{}

func (c FfiConverterOptionalTypePaymentHash) Lift(rb RustBufferI) *PaymentHash {
	return LiftFromRustBuffer[*PaymentHash](c, rb)
}

func (_ FfiConverterOptionalTypePaymentHash) Read(reader io.Reader) *PaymentHash {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentHashINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentHash) Lower(value *PaymentHash) RustBuffer {
	return LowerIntoRustBuffer[*PaymentHash](c, value)
}

func (_ FfiConverterOptionalTypePaymentHash) Write(writer io.Writer, value *PaymentHash) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentHashINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentHash struct{}

func (_ FfiDestroyerOptionalTypePaymentHash) Destroy(value *PaymentHash) {
	if value != nil {
		FfiDestroyerTypePaymentHash{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentId struct{}

var FfiConverterOptionalTypePaymentIdINSTANCE = FfiConverterOptionalTypePaymentId{}

func (c FfiConverterOptionalTypePaymentId) Lift(rb RustBufferI) *PaymentId {
	return LiftFromRustBuffer[*PaymentId](c, rb)
}

func (_ FfiConverterOptionalTypePaymentId) Read(reader io.Reader) *PaymentId {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentIdINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentId) Lower(value *PaymentId) RustBuffer {
	return LowerIntoRustBuffer[*PaymentId](c, value)
}

func (_ FfiConverterOptionalTypePaymentId) Write(writer io.Writer, value *PaymentId) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentIdINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentId struct{}

func (_ FfiDestroyerOptionalTypePaymentId) Destroy(value *PaymentId) {
	if value != nil {
		FfiDestroyerTypePaymentId{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentPreimage struct{}

var FfiConverterOptionalTypePaymentPreimageINSTANCE = FfiConverterOptionalTypePaymentPreimage{}

func (c FfiConverterOptionalTypePaymentPreimage) Lift(rb RustBufferI) *PaymentPreimage {
	return LiftFromRustBuffer[*PaymentPreimage](c, rb)
}

func (_ FfiConverterOptionalTypePaymentPreimage) Read(reader io.Reader) *PaymentPreimage {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentPreimageINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentPreimage) Lower(value *PaymentPreimage) RustBuffer {
	return LowerIntoRustBuffer[*PaymentPreimage](c, value)
}

func (_ FfiConverterOptionalTypePaymentPreimage) Write(writer io.Writer, value *PaymentPreimage) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentPreimageINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentPreimage struct{}

func (_ FfiDestroyerOptionalTypePaymentPreimage) Destroy(value *PaymentPreimage) {
	if value != nil {
		FfiDestroyerTypePaymentPreimage{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePaymentSecret struct{}

var FfiConverterOptionalTypePaymentSecretINSTANCE = FfiConverterOptionalTypePaymentSecret{}

func (c FfiConverterOptionalTypePaymentSecret) Lift(rb RustBufferI) *PaymentSecret {
	return LiftFromRustBuffer[*PaymentSecret](c, rb)
}

func (_ FfiConverterOptionalTypePaymentSecret) Read(reader io.Reader) *PaymentSecret {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePaymentSecretINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePaymentSecret) Lower(value *PaymentSecret) RustBuffer {
	return LowerIntoRustBuffer[*PaymentSecret](c, value)
}

func (_ FfiConverterOptionalTypePaymentSecret) Write(writer io.Writer, value *PaymentSecret) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePaymentSecretINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePaymentSecret struct{}

func (_ FfiDestroyerOptionalTypePaymentSecret) Destroy(value *PaymentSecret) {
	if value != nil {
		FfiDestroyerTypePaymentSecret{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypePublicKey struct{}

var FfiConverterOptionalTypePublicKeyINSTANCE = FfiConverterOptionalTypePublicKey{}

func (c FfiConverterOptionalTypePublicKey) Lift(rb RustBufferI) *PublicKey {
	return LiftFromRustBuffer[*PublicKey](c, rb)
}

func (_ FfiConverterOptionalTypePublicKey) Read(reader io.Reader) *PublicKey {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypePublicKeyINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypePublicKey) Lower(value *PublicKey) RustBuffer {
	return LowerIntoRustBuffer[*PublicKey](c, value)
}

func (_ FfiConverterOptionalTypePublicKey) Write(writer io.Writer, value *PublicKey) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypePublicKeyINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypePublicKey struct{}

func (_ FfiDestroyerOptionalTypePublicKey) Destroy(value *PublicKey) {
	if value != nil {
		FfiDestroyerTypePublicKey{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeTxid struct{}

var FfiConverterOptionalTypeTxidINSTANCE = FfiConverterOptionalTypeTxid{}

func (c FfiConverterOptionalTypeTxid) Lift(rb RustBufferI) *Txid {
	return LiftFromRustBuffer[*Txid](c, rb)
}

func (_ FfiConverterOptionalTypeTxid) Read(reader io.Reader) *Txid {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeTxidINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeTxid) Lower(value *Txid) RustBuffer {
	return LowerIntoRustBuffer[*Txid](c, value)
}

func (_ FfiConverterOptionalTypeTxid) Write(writer io.Writer, value *Txid) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeTxidINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeTxid struct{}

func (_ FfiDestroyerOptionalTypeTxid) Destroy(value *Txid) {
	if value != nil {
		FfiDestroyerTypeTxid{}.Destroy(*value)
	}
}

type FfiConverterOptionalTypeUntrustedString struct{}

var FfiConverterOptionalTypeUntrustedStringINSTANCE = FfiConverterOptionalTypeUntrustedString{}

func (c FfiConverterOptionalTypeUntrustedString) Lift(rb RustBufferI) *UntrustedString {
	return LiftFromRustBuffer[*UntrustedString](c, rb)
}

func (_ FfiConverterOptionalTypeUntrustedString) Read(reader io.Reader) *UntrustedString {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterTypeUntrustedStringINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalTypeUntrustedString) Lower(value *UntrustedString) RustBuffer {
	return LowerIntoRustBuffer[*UntrustedString](c, value)
}

func (_ FfiConverterOptionalTypeUntrustedString) Write(writer io.Writer, value *UntrustedString) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterTypeUntrustedStringINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalTypeUntrustedString struct{}

func (_ FfiDestroyerOptionalTypeUntrustedString) Destroy(value *UntrustedString) {
	if value != nil {
		FfiDestroyerTypeUntrustedString{}.Destroy(*value)
	}
}

type FfiConverterSequenceUint8 struct{}

var FfiConverterSequenceUint8INSTANCE = FfiConverterSequenceUint8{}

func (c FfiConverterSequenceUint8) Lift(rb RustBufferI) []uint8 {
	return LiftFromRustBuffer[[]uint8](c, rb)
}

func (c FfiConverterSequenceUint8) Read(reader io.Reader) []uint8 {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]uint8, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterUint8INSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceUint8) Lower(value []uint8) RustBuffer {
	return LowerIntoRustBuffer[[]uint8](c, value)
}

func (c FfiConverterSequenceUint8) Write(writer io.Writer, value []uint8) {
	if len(value) > math.MaxInt32 {
		panic("[]uint8 is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterUint8INSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceUint8 struct{}

func (FfiDestroyerSequenceUint8) Destroy(sequence []uint8) {
	for _, value := range sequence {
		FfiDestroyerUint8{}.Destroy(value)
	}
}

type FfiConverterSequenceUint64 struct{}

var FfiConverterSequenceUint64INSTANCE = FfiConverterSequenceUint64{}

func (c FfiConverterSequenceUint64) Lift(rb RustBufferI) []uint64 {
	return LiftFromRustBuffer[[]uint64](c, rb)
}

func (c FfiConverterSequenceUint64) Read(reader io.Reader) []uint64 {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]uint64, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterUint64INSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceUint64) Lower(value []uint64) RustBuffer {
	return LowerIntoRustBuffer[[]uint64](c, value)
}

func (c FfiConverterSequenceUint64) Write(writer io.Writer, value []uint64) {
	if len(value) > math.MaxInt32 {
		panic("[]uint64 is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterUint64INSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceUint64 struct{}

func (FfiDestroyerSequenceUint64) Destroy(sequence []uint64) {
	for _, value := range sequence {
		FfiDestroyerUint64{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeChannelDetails struct{}

var FfiConverterSequenceTypeChannelDetailsINSTANCE = FfiConverterSequenceTypeChannelDetails{}

func (c FfiConverterSequenceTypeChannelDetails) Lift(rb RustBufferI) []ChannelDetails {
	return LiftFromRustBuffer[[]ChannelDetails](c, rb)
}

func (c FfiConverterSequenceTypeChannelDetails) Read(reader io.Reader) []ChannelDetails {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]ChannelDetails, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeChannelDetailsINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeChannelDetails) Lower(value []ChannelDetails) RustBuffer {
	return LowerIntoRustBuffer[[]ChannelDetails](c, value)
}

func (c FfiConverterSequenceTypeChannelDetails) Write(writer io.Writer, value []ChannelDetails) {
	if len(value) > math.MaxInt32 {
		panic("[]ChannelDetails is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeChannelDetailsINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeChannelDetails struct{}

func (FfiDestroyerSequenceTypeChannelDetails) Destroy(sequence []ChannelDetails) {
	for _, value := range sequence {
		FfiDestroyerTypeChannelDetails{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeKeyValue struct{}

var FfiConverterSequenceTypeKeyValueINSTANCE = FfiConverterSequenceTypeKeyValue{}

func (c FfiConverterSequenceTypeKeyValue) Lift(rb RustBufferI) []KeyValue {
	return LiftFromRustBuffer[[]KeyValue](c, rb)
}

func (c FfiConverterSequenceTypeKeyValue) Read(reader io.Reader) []KeyValue {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]KeyValue, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeKeyValueINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeKeyValue) Lower(value []KeyValue) RustBuffer {
	return LowerIntoRustBuffer[[]KeyValue](c, value)
}

func (c FfiConverterSequenceTypeKeyValue) Write(writer io.Writer, value []KeyValue) {
	if len(value) > math.MaxInt32 {
		panic("[]KeyValue is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeKeyValueINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeKeyValue struct{}

func (FfiDestroyerSequenceTypeKeyValue) Destroy(sequence []KeyValue) {
	for _, value := range sequence {
		FfiDestroyerTypeKeyValue{}.Destroy(value)
	}
}

type FfiConverterSequenceTypePaymentDetails struct{}

var FfiConverterSequenceTypePaymentDetailsINSTANCE = FfiConverterSequenceTypePaymentDetails{}

func (c FfiConverterSequenceTypePaymentDetails) Lift(rb RustBufferI) []PaymentDetails {
	return LiftFromRustBuffer[[]PaymentDetails](c, rb)
}

func (c FfiConverterSequenceTypePaymentDetails) Read(reader io.Reader) []PaymentDetails {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PaymentDetails, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePaymentDetailsINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePaymentDetails) Lower(value []PaymentDetails) RustBuffer {
	return LowerIntoRustBuffer[[]PaymentDetails](c, value)
}

func (c FfiConverterSequenceTypePaymentDetails) Write(writer io.Writer, value []PaymentDetails) {
	if len(value) > math.MaxInt32 {
		panic("[]PaymentDetails is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePaymentDetailsINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePaymentDetails struct{}

func (FfiDestroyerSequenceTypePaymentDetails) Destroy(sequence []PaymentDetails) {
	for _, value := range sequence {
		FfiDestroyerTypePaymentDetails{}.Destroy(value)
	}
}

type FfiConverterSequenceTypePeerDetails struct{}

var FfiConverterSequenceTypePeerDetailsINSTANCE = FfiConverterSequenceTypePeerDetails{}

func (c FfiConverterSequenceTypePeerDetails) Lift(rb RustBufferI) []PeerDetails {
	return LiftFromRustBuffer[[]PeerDetails](c, rb)
}

func (c FfiConverterSequenceTypePeerDetails) Read(reader io.Reader) []PeerDetails {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PeerDetails, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePeerDetailsINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePeerDetails) Lower(value []PeerDetails) RustBuffer {
	return LowerIntoRustBuffer[[]PeerDetails](c, value)
}

func (c FfiConverterSequenceTypePeerDetails) Write(writer io.Writer, value []PeerDetails) {
	if len(value) > math.MaxInt32 {
		panic("[]PeerDetails is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePeerDetailsINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePeerDetails struct{}

func (FfiDestroyerSequenceTypePeerDetails) Destroy(sequence []PeerDetails) {
	for _, value := range sequence {
		FfiDestroyerTypePeerDetails{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeTlvEntry struct{}

var FfiConverterSequenceTypeTlvEntryINSTANCE = FfiConverterSequenceTypeTlvEntry{}

func (c FfiConverterSequenceTypeTlvEntry) Lift(rb RustBufferI) []TlvEntry {
	return LiftFromRustBuffer[[]TlvEntry](c, rb)
}

func (c FfiConverterSequenceTypeTlvEntry) Read(reader io.Reader) []TlvEntry {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]TlvEntry, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeTlvEntryINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeTlvEntry) Lower(value []TlvEntry) RustBuffer {
	return LowerIntoRustBuffer[[]TlvEntry](c, value)
}

func (c FfiConverterSequenceTypeTlvEntry) Write(writer io.Writer, value []TlvEntry) {
	if len(value) > math.MaxInt32 {
		panic("[]TlvEntry is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeTlvEntryINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeTlvEntry struct{}

func (FfiDestroyerSequenceTypeTlvEntry) Destroy(sequence []TlvEntry) {
	for _, value := range sequence {
		FfiDestroyerTypeTlvEntry{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeLightningBalance struct{}

var FfiConverterSequenceTypeLightningBalanceINSTANCE = FfiConverterSequenceTypeLightningBalance{}

func (c FfiConverterSequenceTypeLightningBalance) Lift(rb RustBufferI) []LightningBalance {
	return LiftFromRustBuffer[[]LightningBalance](c, rb)
}

func (c FfiConverterSequenceTypeLightningBalance) Read(reader io.Reader) []LightningBalance {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]LightningBalance, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeLightningBalanceINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeLightningBalance) Lower(value []LightningBalance) RustBuffer {
	return LowerIntoRustBuffer[[]LightningBalance](c, value)
}

func (c FfiConverterSequenceTypeLightningBalance) Write(writer io.Writer, value []LightningBalance) {
	if len(value) > math.MaxInt32 {
		panic("[]LightningBalance is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeLightningBalanceINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeLightningBalance struct{}

func (FfiDestroyerSequenceTypeLightningBalance) Destroy(sequence []LightningBalance) {
	for _, value := range sequence {
		FfiDestroyerTypeLightningBalance{}.Destroy(value)
	}
}

type FfiConverterSequenceTypePendingSweepBalance struct{}

var FfiConverterSequenceTypePendingSweepBalanceINSTANCE = FfiConverterSequenceTypePendingSweepBalance{}

func (c FfiConverterSequenceTypePendingSweepBalance) Lift(rb RustBufferI) []PendingSweepBalance {
	return LiftFromRustBuffer[[]PendingSweepBalance](c, rb)
}

func (c FfiConverterSequenceTypePendingSweepBalance) Read(reader io.Reader) []PendingSweepBalance {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PendingSweepBalance, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePendingSweepBalanceINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePendingSweepBalance) Lower(value []PendingSweepBalance) RustBuffer {
	return LowerIntoRustBuffer[[]PendingSweepBalance](c, value)
}

func (c FfiConverterSequenceTypePendingSweepBalance) Write(writer io.Writer, value []PendingSweepBalance) {
	if len(value) > math.MaxInt32 {
		panic("[]PendingSweepBalance is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePendingSweepBalanceINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePendingSweepBalance struct{}

func (FfiDestroyerSequenceTypePendingSweepBalance) Destroy(sequence []PendingSweepBalance) {
	for _, value := range sequence {
		FfiDestroyerTypePendingSweepBalance{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeNodeId struct{}

var FfiConverterSequenceTypeNodeIdINSTANCE = FfiConverterSequenceTypeNodeId{}

func (c FfiConverterSequenceTypeNodeId) Lift(rb RustBufferI) []NodeId {
	return LiftFromRustBuffer[[]NodeId](c, rb)
}

func (c FfiConverterSequenceTypeNodeId) Read(reader io.Reader) []NodeId {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]NodeId, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeNodeIdINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeNodeId) Lower(value []NodeId) RustBuffer {
	return LowerIntoRustBuffer[[]NodeId](c, value)
}

func (c FfiConverterSequenceTypeNodeId) Write(writer io.Writer, value []NodeId) {
	if len(value) > math.MaxInt32 {
		panic("[]NodeId is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeNodeIdINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeNodeId struct{}

func (FfiDestroyerSequenceTypeNodeId) Destroy(sequence []NodeId) {
	for _, value := range sequence {
		FfiDestroyerTypeNodeId{}.Destroy(value)
	}
}

type FfiConverterSequenceTypePublicKey struct{}

var FfiConverterSequenceTypePublicKeyINSTANCE = FfiConverterSequenceTypePublicKey{}

func (c FfiConverterSequenceTypePublicKey) Lift(rb RustBufferI) []PublicKey {
	return LiftFromRustBuffer[[]PublicKey](c, rb)
}

func (c FfiConverterSequenceTypePublicKey) Read(reader io.Reader) []PublicKey {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]PublicKey, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypePublicKeyINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypePublicKey) Lower(value []PublicKey) RustBuffer {
	return LowerIntoRustBuffer[[]PublicKey](c, value)
}

func (c FfiConverterSequenceTypePublicKey) Write(writer io.Writer, value []PublicKey) {
	if len(value) > math.MaxInt32 {
		panic("[]PublicKey is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypePublicKeyINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypePublicKey struct{}

func (FfiDestroyerSequenceTypePublicKey) Destroy(sequence []PublicKey) {
	for _, value := range sequence {
		FfiDestroyerTypePublicKey{}.Destroy(value)
	}
}

type FfiConverterSequenceTypeSocketAddress struct{}

var FfiConverterSequenceTypeSocketAddressINSTANCE = FfiConverterSequenceTypeSocketAddress{}

func (c FfiConverterSequenceTypeSocketAddress) Lift(rb RustBufferI) []SocketAddress {
	return LiftFromRustBuffer[[]SocketAddress](c, rb)
}

func (c FfiConverterSequenceTypeSocketAddress) Read(reader io.Reader) []SocketAddress {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]SocketAddress, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterTypeSocketAddressINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceTypeSocketAddress) Lower(value []SocketAddress) RustBuffer {
	return LowerIntoRustBuffer[[]SocketAddress](c, value)
}

func (c FfiConverterSequenceTypeSocketAddress) Write(writer io.Writer, value []SocketAddress) {
	if len(value) > math.MaxInt32 {
		panic("[]SocketAddress is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterTypeSocketAddressINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceTypeSocketAddress struct{}

func (FfiDestroyerSequenceTypeSocketAddress) Destroy(sequence []SocketAddress) {
	for _, value := range sequence {
		FfiDestroyerTypeSocketAddress{}.Destroy(value)
	}
}

type FfiConverterMapStringString struct{}

var FfiConverterMapStringStringINSTANCE = FfiConverterMapStringString{}

func (c FfiConverterMapStringString) Lift(rb RustBufferI) map[string]string {
	return LiftFromRustBuffer[map[string]string](c, rb)
}

func (_ FfiConverterMapStringString) Read(reader io.Reader) map[string]string {
	result := make(map[string]string)
	length := readInt32(reader)
	for i := int32(0); i < length; i++ {
		key := FfiConverterStringINSTANCE.Read(reader)
		value := FfiConverterStringINSTANCE.Read(reader)
		result[key] = value
	}
	return result
}

func (c FfiConverterMapStringString) Lower(value map[string]string) RustBuffer {
	return LowerIntoRustBuffer[map[string]string](c, value)
}

func (_ FfiConverterMapStringString) Write(writer io.Writer, mapValue map[string]string) {
	if len(mapValue) > math.MaxInt32 {
		panic("map[string]string is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(mapValue)))
	for key, value := range mapValue {
		FfiConverterStringINSTANCE.Write(writer, key)
		FfiConverterStringINSTANCE.Write(writer, value)
	}
}

type FfiDestroyerMapStringString struct{}

func (_ FfiDestroyerMapStringString) Destroy(mapValue map[string]string) {
	for key, value := range mapValue {
		FfiDestroyerString{}.Destroy(key)
		FfiDestroyerString{}.Destroy(value)
	}
}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Address = string
type FfiConverterTypeAddress = FfiConverterString
type FfiDestroyerTypeAddress = FfiDestroyerString

var FfiConverterTypeAddressINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type BlockHash = string
type FfiConverterTypeBlockHash = FfiConverterString
type FfiDestroyerTypeBlockHash = FfiDestroyerString

var FfiConverterTypeBlockHashINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Bolt11Invoice = string
type FfiConverterTypeBolt11Invoice = FfiConverterString
type FfiDestroyerTypeBolt11Invoice = FfiDestroyerString

var FfiConverterTypeBolt11InvoiceINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Bolt12Invoice = string
type FfiConverterTypeBolt12Invoice = FfiConverterString
type FfiDestroyerTypeBolt12Invoice = FfiDestroyerString

var FfiConverterTypeBolt12InvoiceINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type ChannelId = string
type FfiConverterTypeChannelId = FfiConverterString
type FfiDestroyerTypeChannelId = FfiDestroyerString

var FfiConverterTypeChannelIdINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Mnemonic = string
type FfiConverterTypeMnemonic = FfiConverterString
type FfiDestroyerTypeMnemonic = FfiDestroyerString

var FfiConverterTypeMnemonicINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Network = string
type FfiConverterTypeNetwork = FfiConverterString
type FfiDestroyerTypeNetwork = FfiDestroyerString

var FfiConverterTypeNetworkINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type NodeAlias = string
type FfiConverterTypeNodeAlias = FfiConverterString
type FfiDestroyerTypeNodeAlias = FfiDestroyerString

var FfiConverterTypeNodeAliasINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type NodeId = string
type FfiConverterTypeNodeId = FfiConverterString
type FfiDestroyerTypeNodeId = FfiDestroyerString

var FfiConverterTypeNodeIdINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Offer = string
type FfiConverterTypeOffer = FfiConverterString
type FfiDestroyerTypeOffer = FfiDestroyerString

var FfiConverterTypeOfferINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type OfferId = string
type FfiConverterTypeOfferId = FfiConverterString
type FfiDestroyerTypeOfferId = FfiDestroyerString

var FfiConverterTypeOfferIdINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PaymentHash = string
type FfiConverterTypePaymentHash = FfiConverterString
type FfiDestroyerTypePaymentHash = FfiDestroyerString

var FfiConverterTypePaymentHashINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PaymentId = string
type FfiConverterTypePaymentId = FfiConverterString
type FfiDestroyerTypePaymentId = FfiDestroyerString

var FfiConverterTypePaymentIdINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PaymentPreimage = string
type FfiConverterTypePaymentPreimage = FfiConverterString
type FfiDestroyerTypePaymentPreimage = FfiDestroyerString

var FfiConverterTypePaymentPreimageINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PaymentSecret = string
type FfiConverterTypePaymentSecret = FfiConverterString
type FfiDestroyerTypePaymentSecret = FfiDestroyerString

var FfiConverterTypePaymentSecretINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type PublicKey = string
type FfiConverterTypePublicKey = FfiConverterString
type FfiDestroyerTypePublicKey = FfiDestroyerString

var FfiConverterTypePublicKeyINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Refund = string
type FfiConverterTypeRefund = FfiConverterString
type FfiDestroyerTypeRefund = FfiDestroyerString

var FfiConverterTypeRefundINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type SocketAddress = string
type FfiConverterTypeSocketAddress = FfiConverterString
type FfiDestroyerTypeSocketAddress = FfiDestroyerString

var FfiConverterTypeSocketAddressINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type Txid = string
type FfiConverterTypeTxid = FfiConverterString
type FfiDestroyerTypeTxid = FfiDestroyerString

var FfiConverterTypeTxidINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type UntrustedString = string
type FfiConverterTypeUntrustedString = FfiConverterString
type FfiDestroyerTypeUntrustedString = FfiDestroyerString

var FfiConverterTypeUntrustedStringINSTANCE = FfiConverterString{}

/**
 * Typealias from the type name used in the UDL file to the builtin type.  This
 * is needed because the UDL type name is used in function/method signatures.
 * It's also what we have an external type that references a custom type.
 */
type UserChannelId = string
type FfiConverterTypeUserChannelId = FfiConverterString
type FfiDestroyerTypeUserChannelId = FfiDestroyerString

var FfiConverterTypeUserChannelIdINSTANCE = FfiConverterString{}

func DefaultConfig() Config {
	return FfiConverterTypeConfigINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_func_default_config(_uniffiStatus)
	}))
}

func GenerateEntropyMnemonic() Mnemonic {
	return FfiConverterTypeMnemonicINSTANCE.Lift(rustCall(func(_uniffiStatus *C.RustCallStatus) RustBufferI {
		return C.uniffi_ldk_node_fn_func_generate_entropy_mnemonic(_uniffiStatus)
	}))
}
