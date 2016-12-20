// Copyright 2013-2016 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	ParticleType "github.com/verticalmass/aerospike-client-go/types/particle_type"
	Buffer "github.com/verticalmass/aerospike-client-go/utils/buffer"
)

var __packObjectReflect func(aerospikeBuffer, interface{}, bool) (int, error)

func __PackIfcList(cmd aerospikeBuffer, list []interface{}) (int, error) {
	size := 0
	n, err := __PackArrayBegin(cmd, len(list))
	if err != nil {
		return n, err
	}
	size += n

	for i := range list {
		n, err := __PackObject(cmd, list[i], false)
		if err != nil {
			return 0, err
		}
		size += n
	}

	return size, err
}

func __PackList(cmd aerospikeBuffer, list ListIter) (int, error) {
	size := 0
	n, err := __PackArrayBegin(cmd, list.Len())
	if err != nil {
		return n, err
	}
	size += n

	f := func(v interface{}) error {
		n, err := __PackObject(cmd, v, false)
		if err != nil {
			return err
		}
		size += n

		return nil
	}

	err = list.Range(f)
	return size, err
}

func __PackValueArray(cmd aerospikeBuffer, list ValueArray) (int, error) {
	size := 0
	n, err := __PackArrayBegin(cmd, len(list))
	if err != nil {
		return n, err
	}
	size += n

	for i := range list {
		n, err := list[i].pack(cmd)
		if err != nil {
			return 0, err
		}
		size += n
	}

	return size, err
}

func __PackArrayBegin(cmd aerospikeBuffer, size int) (int, error) {
	if size < 16 {
		return __PackAByte(cmd, 0x90|byte(size))
	} else if size <= math.MaxUint16 {
		return __PackShort(cmd, 0xdc, int16(size))
	} else {
		return __PackInt(cmd, 0xdd, int32(size))
	}
}

func __PackIfcMap(cmd aerospikeBuffer, theMap map[interface{}]interface{}) (int, error) {
	size := 0
	n, err := __PackMapBegin(cmd, len(theMap))
	if err != nil {
		return n, err
	}
	size += n

	for k, v := range theMap {
		n, err := __PackObject(cmd, k, true)
		if err != nil {
			return 0, err
		}
		size += n
		n, err = __PackObject(cmd, v, false)
		if err != nil {
			return 0, err
		}
		size += n
	}

	return size, err
}

func __PackJsonMap(cmd aerospikeBuffer, theMap map[string]interface{}) (int, error) {
	size := 0
	n, err := __PackMapBegin(cmd, len(theMap))
	if err != nil {
		return n, err
	}
	size += n

	for k, v := range theMap {
		n, err := __PackString(cmd, k)
		if err != nil {
			return 0, err
		}
		size += n
		n, err = __PackObject(cmd, v, false)
		if err != nil {
			return 0, err
		}
		size += n
	}

	return size, err
}

func __PackMap(cmd aerospikeBuffer, theMap MapIter) (int, error) {
	size := 0
	n, err := __PackMapBegin(cmd, theMap.Len())
	if err != nil {
		return n, err
	}
	size += n

	f := func(k, v interface{}) error {
		n, err := __PackObject(cmd, k, true)
		if err != nil {
			return err
		}
		size += n
		n, err = __PackObject(cmd, v, false)
		if err != nil {
			return err
		}
		size += n

		return nil
	}

	err = theMap.Range(f)
	return size, err
}

func __PackMapBegin(cmd aerospikeBuffer, size int) (int, error) {
	if size < 16 {
		return __PackAByte(cmd, 0x80|byte(size))
	} else if size <= math.MaxUint16 {
		return __PackShort(cmd, 0xde, int16(size))
	} else {
		return __PackInt(cmd, 0xdf, int32(size))
	}
}

func __PackBytes(cmd aerospikeBuffer, b []byte) (int, error) {
	size := 0
	n, err := __PackByteArrayBegin(cmd, len(b)+1)
	if err != nil {
		return n, err
	}
	size += n

	n, err = __PackAByte(cmd, ParticleType.BLOB)
	if err != nil {
		return size + n, err
	}
	size += n

	n, err = __PackByteArray(cmd, b)
	if err != nil {
		return size + n, err
	}
	size += n

	return size, nil
}

func __PackByteArrayBegin(cmd aerospikeBuffer, length int) (int, error) {
	if length < 32 {
		return __PackAByte(cmd, 0xa0|byte(length))
	} else if length < 65536 {
		return __PackShort(cmd, 0xda, int16(length))
	} else {
		return __PackInt(cmd, 0xdb, int32(length))
	}
}

func __PackObject(cmd aerospikeBuffer, obj interface{}, mapKey bool) (int, error) {
	switch v := obj.(type) {
	case Value:
		return v.pack(cmd)
	case string:
		return __PackString(cmd, v)
	case []byte:
		return __PackBytes(cmd, obj.([]byte))
	case int8:
		return __PackAInt(cmd, int(v))
	case uint8:
		return __PackAInt(cmd, int(v))
	case int16:
		return __PackAInt(cmd, int(v))
	case uint16:
		return __PackAInt(cmd, int(v))
	case int32:
		return __PackAInt(cmd, int(v))
	case uint32:
		return __PackAInt(cmd, int(v))
	case int:
		if Buffer.Arch32Bits {
			return __PackAInt(cmd, v)
		}
		return __PackAInt64(cmd, int64(v))
	case uint:
		if Buffer.Arch32Bits {
			return __PackAInt(cmd, int(v))
		}
		return __PackAUInt64(cmd, uint64(v))
	case int64:
		return __PackAInt64(cmd, v)
	case uint64:
		return __PackAUInt64(cmd, v)
	case time.Time:
		return __PackAInt64(cmd, v.UnixNano())
	case nil:
		return __PackNil(cmd)
	case bool:
		return __PackBool(cmd, v)
	case float32:
		return __PackFloat32(cmd, v)
	case float64:
		return __PackFloat64(cmd, v)
	case struct{}:
		if mapKey {
			panic(fmt.Sprintf("Maps, Slices, and bounded arrays other than Bounded Byte Arrays are not supported as Map keys. Value: %#v", v))
		}
		return __PackIfcMap(cmd, map[interface{}]interface{}{})
	case []interface{}:
		if mapKey {
			panic(fmt.Sprintf("Maps, Slices, and bounded arrays other than Bounded Byte Arrays are not supported as Map keys. Value: %#v", v))
		}
		return __PackIfcList(cmd, v)
	case map[interface{}]interface{}:
		if mapKey {
			panic(fmt.Sprintf("Maps, Slices, and bounded arrays other than Bounded Byte Arrays are not supported as Map keys. Value: %#v", v))
		}
		return __PackIfcMap(cmd, v)
	case ListIter:
		if mapKey {
			panic(fmt.Sprintf("Maps, Slices, and bounded arrays other than Bounded Byte Arrays are not supported as Map keys. Value: %#v", v))
		}
		return __PackList(cmd, obj.(ListIter))
	case MapIter:
		if mapKey {
			panic(fmt.Sprintf("Maps, Slices, and bounded arrays other than Bounded Byte Arrays are not supported as Map keys. Value: %#v", v))
		}
		return __PackMap(cmd, obj.(MapIter))
	}

	if __packObjectReflect != nil {
		return __packObjectReflect(cmd, obj, mapKey)
	}

	panic(fmt.Sprintf("Type `%#v` not supported to pack. ", obj))
}

func __PackAUInt64(cmd aerospikeBuffer, val uint64) (int, error) {
	return __PackUInt64(cmd, val)
}

func __PackAInt64(cmd aerospikeBuffer, val int64) (int, error) {
	if val >= 0 {
		if val < 128 {
			return __PackAByte(cmd, byte(val))
		}

		if val <= math.MaxUint8 {
			return __PackByte(cmd, 0xcc, byte(val))
		}

		if val <= math.MaxUint16 {
			return __PackShort(cmd, 0xcd, int16(val))
		}

		if val <= math.MaxUint32 {
			return __PackInt(cmd, 0xce, int32(val))
		}
		return __PackInt64(cmd, 0xd3, val)
	} else {
		if val >= -32 {
			return __PackAByte(cmd, 0xe0|(byte(val)+32))
		}

		if val >= math.MinInt8 {
			return __PackByte(cmd, 0xd0, byte(val))
		}

		if val >= math.MinInt16 {
			return __PackShort(cmd, 0xd1, int16(val))
		}

		if val >= math.MinInt32 {
			return __PackInt(cmd, 0xd2, int32(val))
		}
		return __PackInt64(cmd, 0xd3, val)
	}
}

func __PackAInt(cmd aerospikeBuffer, val int) (int, error) {
	if val >= 0 {
		if val < 128 {
			return __PackAByte(cmd, byte(val))
		}

		if val < 256 {
			return __PackByte(cmd, 0xcc, byte(val))
		}

		if val < 65536 {
			return __PackShort(cmd, 0xcd, int16(val))
		}
		return __PackInt(cmd, 0xce, int32(val))
	} else {
		if val >= -32 {
			return __PackAByte(cmd, 0xe0|(byte(val)+32))
		}

		if val >= math.MinInt8 {
			return __PackByte(cmd, 0xd0, byte(val))
		}

		if val >= math.MinInt16 {
			return __PackShort(cmd, 0xd1, int16(val))
		}
		return __PackInt(cmd, 0xd2, int32(val))
	}
}

func __PackString(cmd aerospikeBuffer, val string) (int, error) {
	size := 0
	slen := len(val) + 1
	n, err := __PackByteArrayBegin(cmd, slen)
	if err != nil {
		return n, err
	}
	size += n

	if cmd != nil {
		n, err = cmd.WriteByte(byte(ParticleType.STRING))
		if err != nil {
			return size + n, err
		}
		size += n

		n, err = cmd.WriteString(val)
		if err != nil {
			return size + n, err
		}
		size += n
	} else {
		size += 1 + len(val)
	}

	return size, nil
}

func __PackGeoJson(cmd aerospikeBuffer, val string) (int, error) {
	size := 0
	slen := len(val) + 1
	n, err := __PackByteArrayBegin(cmd, slen)
	if err != nil {
		return n, err
	}
	size += n

	if cmd != nil {
		n, err = cmd.WriteByte(byte(ParticleType.GEOJSON))
		if err != nil {
			return size + n, err
		}
		size += n

		n, err = cmd.WriteString(val)
		if err != nil {
			return size + n, err
		}
		size += n
	} else {
		size += 1 + len(val)
	}

	return size, nil
}

func __PackByteArray(cmd aerospikeBuffer, src []byte) (int, error) {
	if cmd != nil {
		return cmd.Write(src)
	}
	return len(src), nil
}

func __PackInt64(cmd aerospikeBuffer, valType int, val int64) (int, error) {
	if cmd != nil {
		size, err := cmd.WriteByte(byte(valType))
		if err != nil {
			return size, err
		}

		n, err := cmd.WriteInt64(val)
		return size + n, err
	}
	return 1 + 8, nil
}

func __PackUInt64(cmd aerospikeBuffer, val uint64) (int, error) {
	if cmd != nil {
		size, err := cmd.WriteByte(byte(0xcf))
		if err != nil {
			return size, err
		}

		n, err := cmd.WriteInt64(int64(val))
		return size + n, err
	}
	return 1 + 8, nil
}

func __PackInt(cmd aerospikeBuffer, valType int, val int32) (int, error) {
	if cmd != nil {
		size, err := cmd.WriteByte(byte(valType))
		if err != nil {
			return size, err
		}
		n, err := cmd.WriteInt32(val)
		return size + n, err
	}
	return 1 + 4, nil
}

func __PackShort(cmd aerospikeBuffer, valType int, val int16) (int, error) {
	if cmd != nil {
		size, err := cmd.WriteByte(byte(valType))
		if err != nil {
			return size, err
		}

		n, err := cmd.WriteInt16(val)
		return size + n, err
	}
	return 1 + 2, nil
}

// This method is not compatible with MsgPack specs and is only used by aerospike client<->server
// for wire transfer only
func __PackShortRaw(cmd aerospikeBuffer, val int16) (int, error) {
	if cmd != nil {
		return cmd.WriteInt16(val)
	}
	return 2, nil
}

func __PackByte(cmd aerospikeBuffer, valType int, val byte) (int, error) {
	if cmd != nil {
		size := 0
		n, err := cmd.WriteByte(byte(valType))
		if err != nil {
			return n, err
		}
		size += n

		n, err = cmd.WriteByte(val)
		if err != nil {
			return size + n, err
		}
		size += n

		return size, nil
	}
	return 1 + 1, nil
}

func __PackNil(cmd aerospikeBuffer) (int, error) {
	if cmd != nil {
		return cmd.WriteByte(0xc0)
	}
	return 1, nil
}

func __PackBool(cmd aerospikeBuffer, val bool) (int, error) {
	if cmd != nil {
		if val {
			return cmd.WriteByte(0xc3)
		}
		return cmd.WriteByte(0xc2)
	}
	return 1, nil
}

func __PackFloat32(cmd aerospikeBuffer, val float32) (int, error) {
	if cmd != nil {
		size := 0
		n, err := cmd.WriteByte(0xca)
		if err != nil {
			return n, err
		}
		size += n
		n, err = cmd.WriteFloat32(val)
		return size + n, err
	}
	return 1 + 4, nil
}

func __PackFloat64(cmd aerospikeBuffer, val float64) (int, error) {
	if cmd != nil {
		size := 0
		n, err := cmd.WriteByte(0xcb)
		if err != nil {
			return n, err
		}
		size += n
		n, err = cmd.WriteFloat64(val)
		return size + n, err
	}
	return 1 + 8, nil
}

func __PackAByte(cmd aerospikeBuffer, val byte) (int, error) {
	if cmd != nil {
		return cmd.WriteByte(val)
	}
	return 1, nil
}

// packer implements a buffered packer
type packer struct {
	bytes.Buffer
	tempBuffer [8]byte
}

func newPacker() *packer {
	return &packer{}
}

// Int64ToBytes converts an int64 into slice of Bytes.
func (vb *packer) WriteInt64(num int64) (int, error) {
	return vb.WriteUint64(uint64(num))
}

// Uint64ToBytes converts an uint64 into slice of Bytes.
func (vb *packer) WriteUint64(num uint64) (int, error) {
	binary.BigEndian.PutUint64(vb.tempBuffer[:8], num)
	vb.Write(vb.tempBuffer[:8])
	return 8, nil
}

// Int32ToBytes converts an int32 to a byte slice of size 4
func (vb *packer) WriteInt32(num int32) (int, error) {
	return vb.WriteUint32(uint32(num))
}

// Uint32ToBytes converts an uint32 to a byte slice of size 4
func (vb *packer) WriteUint32(num uint32) (int, error) {
	binary.BigEndian.PutUint32(vb.tempBuffer[:4], num)
	vb.Write(vb.tempBuffer[:4])
	return 4, nil
}

// Int16ToBytes converts an int16 to slice of bytes
func (vb *packer) WriteInt16(num int16) (int, error) {
	return vb.WriteUint16(uint16(num))
}

// UInt16ToBytes converts an iuint16 to slice of bytes
func (vb *packer) WriteUint16(num uint16) (int, error) {
	binary.BigEndian.PutUint16(vb.tempBuffer[:2], num)
	vb.Write(vb.tempBuffer[:2])
	return 2, nil
}

func (vb *packer) WriteFloat32(float float32) (int, error) {
	bits := math.Float32bits(float)
	binary.BigEndian.PutUint32(vb.tempBuffer[:4], bits)
	vb.Write(vb.tempBuffer[:4])
	return 4, nil
}

func (vb *packer) WriteFloat64(float float64) (int, error) {
	bits := math.Float64bits(float)
	binary.BigEndian.PutUint64(vb.tempBuffer[:8], bits)
	vb.Write(vb.tempBuffer[:8])
	return 8, nil
}

func (vb *packer) WriteByte(b byte) (int, error) {
	vb.Write([]byte{b})
	return 1, nil
}

func (vb *packer) WriteString(s string) (int, error) {
	// To avoid allocating memory, write the strings in small chunks
	l := len(s)
	const size = 128
	b := [size]byte{}
	cnt := 0
	for i := 0; i < l; i++ {
		b[cnt] = s[i]
		cnt++

		if cnt == size {
			vb.Write(b[:])
			cnt = 0
		}
	}

	if cnt > 0 {
		vb.Write(b[:cnt])
	}

	return len(s), nil
}
