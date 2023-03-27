package fsmp

import (
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	pool := CreatePool(100, 8)
	assert.NotNil(t, pool)
}

func TestAddrConvertions(t *testing.T) {
	pool := CreatePool(5, 8)
	assert.NotNil(t, pool)

	p_0 := pool.addrFromIndex(0)
	p_2 := pool.addrFromIndex(2)
	p_4 := pool.addrFromIndex(4)

	assert.Equal(t, uintptr(unsafe.Pointer(p_0))+uintptr(2*8), uintptr(unsafe.Pointer(p_2)))
	assert.Equal(t, uintptr(unsafe.Pointer(p_0))+uintptr(4*8), uintptr(unsafe.Pointer(p_4)))

	assert.Equal(t, uint(0), pool.indexFromAddr(p_0))
	assert.Equal(t, uint(2), pool.indexFromAddr(p_2))
	assert.Equal(t, uint(4), pool.indexFromAddr(p_4))
}

func TestAllocateUntilFull(t *testing.T) {
	pool := CreatePool(5, 8)
	assert.NotNil(t, pool)

	for i := 0; i < 5; i++ {
		b, err := pool.Allocate()
		assert.NotNil(t, b)
		assert.Nil(t, err)
	}
}

func TestAllocateUntilFullPlusOne(t *testing.T) {
	pool := CreatePool(5, 8)
	assert.NotNil(t, pool)

	for i := 0; i < 5; i++ {
		b, err := pool.Allocate()
		assert.NotNil(t, b)
		assert.Nil(t, err)
	}
	b, err := pool.Allocate()
	assert.Nil(t, b)
	assert.Equal(t, ErrOutOfMemory, err)
}

func TestDeAlloc(t *testing.T) {
	pool := CreatePool(1, 8)
	assert.NotNil(t, pool)

	b, err := pool.Allocate()
	assert.NotNil(t, b)
	assert.Nil(t, err)

	assert.Nil(t, pool.DeAllocate(b))

	b, err = pool.Allocate()
	assert.NotNil(t, b)
	assert.Nil(t, err)
}

func TestDeAllocOutOfBound(t *testing.T) {
	pool := CreatePool(10, 8)
	assert.NotNil(t, pool)

	b := make([]byte, 8)

	assert.Equal(t, ErrOutOfBound, pool.DeAllocate(b))
}

func fill(buff []byte, s uint8) {
	for i := range buff {
		buff[i] = s
	}
}

func TestAllocAndDeAllocLoop(t *testing.T) {
	const size = 89
	const blockSize = 97
	pool := CreatePool(size, blockSize)
	assert.NotNil(t, pool)
	arr := make([][]byte, size)

	for i := 0; i < 3; i++ {
		var err error
		for j := 0; j < size; j++ {
			arr[j], err = pool.Allocate()
			assert.NotNil(t, arr[j])
			assert.Nil(t, err)
			assert.Equal(t, blockSize, len(arr[j]))
			assert.Equal(t, blockSize, cap(arr[j]))
			fill(arr[j], uint8(j))
		}
		b, err := pool.Allocate()
		assert.Nil(t, b)
		assert.Equal(t, ErrOutOfMemory, err)
		for j := 0; j < size; j++ {
			tmp := make([]byte, blockSize)
			fill(tmp, uint8(j))
			assert.Equal(t, arr[j], tmp)
		}
		for j := 0; j < size; j++ {
			assert.Nil(t, pool.DeAllocate(arr[j]))
		}
	}
}

func TestAllocConfiguredBlockSize(t *testing.T) {
	const size = 20
	pool := CreatePool(size, 2)
	assert.NotNil(t, pool)
	arr := make([][]byte, size)

	var err error
	for j := 0; j < size; j++ {
		arr[j], err = pool.Allocate()
		assert.NotNil(t, arr[j])
		assert.Nil(t, err)
		binary.LittleEndian.PutUint16(arr[j], uint16(j))
	}
	for j := 0; j < size; j++ {
		tmp := make([]byte, 2)
		binary.LittleEndian.PutUint16(tmp, uint16(j))
		assert.Equal(t, tmp, arr[j])
	}
	for j := 0; j < size; j++ {
		assert.Nil(t, pool.DeAllocate(arr[j]))
	}
}
