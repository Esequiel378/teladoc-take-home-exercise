package bigint

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"

	"teladoc/pkg/utils"
)

var (
	// ErrInvalidIntegerNumber is returned when the input string is not a valid integer number
	ErrInvalidIntegerNumber = errors.New("invalid integer number")
	// ErrConvertingChunkToInteger is returned when a chunk cannot be converted to integer
	ErrConvertingChunkToInteger = errors.New("error converting chunk to integer")
)

// Bigconstant Int represents a large integer number.
type BigInt struct {
	// magnitude is where the number is stored in chunks
	magnitude []uint32
	// length represents the number of digits in the BigInt
	length int
	// chukSize represents the number of digits in each chunk
	chukSize int
}

// IntegerNumberMatch is a regex that matches an integer number (without decimal places)
// Ex: 123, 123456789012345678901234567890
//
// [101 reference](https://regex101.com/r/3hoFC3/1)
const IntegerNumberMatch = "^[0-9]+$"

// NewBigInt creates a new BigInt from a string
// The string must be a valid integer number
// and must not contain any decimal places
//
// Ex: 123, 123456789012345678901234567890, etc.
func NewBigInt(value string) (*BigInt, error) {
	// Validate input value
	match, err := regexp.MatchString(IntegerNumberMatch, value)
	if !match || err != nil {
		return nil, ErrInvalidIntegerNumber
	}

	// Break the string into chunks of 8 digits
	// Breaking in chunks of 8 digits allows us to use uint32
	// to store and perform the addition operation on the number
	// TODO: Invsigate if we can use any other data type
	chunkSize := 9
	chunks := utils.ChunkString(value, chunkSize)

	magnitude := make([]uint32, len(chunks))

	// Convert each chunk to uint32
	for idx, chunk := range chunks {
		integer, err := utils.StringToUint32(chunk)
		if err != nil {
			return nil, ErrConvertingChunkToInteger
		}

		magnitude[idx] = integer
	}

	bigInt := &BigInt{
		magnitude: magnitude,
		length:    len(value),
		chukSize:  chunkSize,
	}

	return bigInt, nil
}

// Length returns the number of digits in the BigInt.
func (b BigInt) Length() int {
	return b.length
}

// String returns the string representation of the BigInt.
func (b BigInt) String() string {
	var result strings.Builder

	for _, chunk := range b.magnitude {
		value := strconv.FormatUint(uint64(chunk), 10)
		result.WriteString(value)
	}

	return result.String()
}

// Add adds two BigInts and returns the result.
func (b BigInt) Add(other *BigInt) *BigInt {
	lhs, rhs := b.magnitude, other.magnitude

	// Make sure the larger magnitude is always on the left
	if b.Length() < other.Length() {
		lhs, rhs = rhs, lhs
	}

	// Create a new BigInt to hold the result
	result := &BigInt{
		magnitude: make([]uint32, len(lhs)),
	}

	// Siplify the addition for single chuck setup
	if len(lhs) == 1 {
		result.magnitude[0] = lhs[0] + rhs[0]

		return result
	}

	var carry bool

	for offset := 1; offset <= len(lhs); offset++ {
		// Get the chunk index
		index := len(lhs) - offset

		// Get the chunk values, rhs may be shorter than lhs
		// so we need to check if the index is out of bounds
		// and if so, default to `0` as the value
		var (
			lhsChunk = lhs[index]
			rhsChunk uint32
		)

		// Get the chunk value from the right
		// If the right chunk does not exist, use 0
		if index < len(rhs) {
			rhsChunk = rhs[index]
		}

		// Add the two chunks
		sum := lhsChunk + rhsChunk

		// Add the carry to the sum
		if carry {
			sum++
		}

		// Count the number of digits to determine if we need to carry
		sumDigits := utils.CountDigits(int64(sum))

		// If the sum doesn't fit we need to carry to the next chunk
		carry = sumDigits > b.chukSize

		if carry {
			// Remove the carry from the sum
			exponential := math.Pow10(b.chukSize)
			// sum %= 10**b.chukSize
			sum %= uint32(exponential)
		}

		// Store the sum in the result
		result.magnitude[index] = sum
	}

	// If we have a carry left, we need to add a new chunk
	if carry {
		newMagnitude := make([]uint32, len(result.magnitude)+1)
		newMagnitude[0] = 1
		copy(newMagnitude[1:], result.magnitude)
		result.magnitude = newMagnitude
	}

	return result
}