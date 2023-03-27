package fsmp

import (
	"fmt"
	"unsafe"

	"github.com/tidwall/spinlock"
)

type Pool struct {
	spinlock.Locker
	m_numOfBlocks,
	m_sizeOfEachBlock,
	configuredBlockSize,
	m_numFreeBlocks,
	m_numInitialized uint
	buff []byte
	m_memStart,
	m_next *uint // index number / slot number
}

var (
	ErrOutOfMemory error = fmt.Errorf("no space left")
	ErrOutOfBound  error = fmt.Errorf("index is out of bound")
)

func addrFromSlice(b []byte) *uint {
	return (*uint)(unsafe.Pointer(&b[0]))
}

func (p *Pool) addrFromIndex(index uint) *uint {
	start := uintptr(unsafe.Pointer(p.m_memStart))
	return (*uint)(unsafe.Pointer(start + uintptr(index*p.m_sizeOfEachBlock)))
}

func (p *Pool) indexFromAddr(addr *uint) uint {
	start := uintptr(unsafe.Pointer(p.m_memStart))
	indexPtr := uintptr(unsafe.Pointer(addr))
	relativeOffset := indexPtr - start
	return uint(relativeOffset) / p.m_sizeOfEachBlock
}

// CreatePool creates a new memory pool with a fixed number of blocks (numBlocks) and size of each block (blockSize).
func CreatePool(numBlocks, blockSize uint) *Pool {
	configuredBlockSize := blockSize
	if uintptr(blockSize) < unsafe.Sizeof(numBlocks) {
		blockSize = uint(unsafe.Sizeof(numBlocks))
	}

	b := make([]byte, numBlocks*blockSize)
	return &Pool{
		// look into the paper for details!
		// Used equal naming for variables and functions
		m_numOfBlocks:     numBlocks,
		m_sizeOfEachBlock: blockSize,
		m_numFreeBlocks:   numBlocks,
		m_numInitialized:  0,
		m_memStart:        addrFromSlice(b),
		m_next:            addrFromSlice(b),

		configuredBlockSize: configuredBlockSize,
		buff:                b, // keep holding a reference to the underling memory block
	}
}

// Allocate returns a new slice where len and cap equals `blockSize`. If the limit of `numBlocks`
// concurrent allocations is reached, ErrOutOfMemory is returned.
func (p *Pool) Allocate() ([]byte, error) {

	p.Lock()
	defer p.Unlock()

	if p.m_numInitialized < p.m_numOfBlocks {
		x := p.addrFromIndex(p.m_numInitialized)
		*x = p.m_numInitialized + 1
		p.m_numInitialized++
	}

	if p.m_numFreeBlocks > 0 {
		res := unsafe.Slice((*byte)(unsafe.Pointer(p.m_next)), p.configuredBlockSize)
		p.m_numFreeBlocks--
		if p.m_numFreeBlocks > 0 {
			p.m_next = p.addrFromIndex(*p.m_next)
		} else {
			p.m_next = nil
		}
		return res, nil
	}

	return nil, ErrOutOfMemory
}

// DeAllocate releases a previous allocated slice. In case of passing an invalid slice, ErrOutOfBound is returned.
// Furthermore the released slice must have the same start position provided by allocate.
func (p *Pool) DeAllocate(b []byte) error {

	indexPtr := addrFromSlice(b)
	uintptrIndex := uintptr(unsafe.Pointer(indexPtr))
	if uintptrIndex < uintptr(unsafe.Pointer(p.m_memStart)) ||
		uintptrIndex > uintptr(unsafe.Pointer(p.addrFromIndex(p.m_numOfBlocks))) {
		return ErrOutOfBound
	}
	p.Lock()
	defer p.Unlock()

	var index uint
	if p.m_next == nil {
		index = p.m_numOfBlocks
	} else {
		index = p.indexFromAddr(p.m_next)
	}
	*indexPtr = index
	p.m_next = indexPtr
	p.m_numFreeBlocks++

	return nil
}
