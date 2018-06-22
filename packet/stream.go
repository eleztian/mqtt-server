package packet

import (
	"bufio"
	"bytes"
	"io"
	"errors"
)
var ErrDetectionOverflow = errors.New("detection overflow")

// ErrReadLimitExceeded can be returned during a Receive if the connection
// exceeded its read limit.
//
// Note: this error is wrapped in an Error with a NetworkError code.
var ErrReadLimitExceeded = errors.New("read limit exceeded")

type Encoder struct {
	writer *bufio.Writer
	buffer bytes.Buffer
}

func NewEncoder(writer io.Writer) *Encoder {
	return &Encoder{
		writer: bufio.NewWriter(writer),
	}
}

func (e *Encoder) Write(pkt Packet) error {
	// reset and eventually grow buffer
	packetLength := pkt.Len()
	e.buffer.Reset()
	e.buffer.Grow(packetLength)
	buf := e.buffer.Bytes()[0:packetLength]

	// encode packet
	_, err := pkt.Encode(buf)
	if err != nil {
		return err
	}

	// write buffer
	_, err = e.writer.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func (e *Encoder) Flush() error {
	return e.writer.Flush()
}

type Decoder struct {
	Limit int64

	reader *bufio.Reader
	buffer bytes.Buffer
}

func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{
		reader: bufio.NewReader(reader),
	}
}

// Read reads the next packet from the buffered reader.
func (d *Decoder) Read() (Packet, error) {
	// initial detection length
	detectionLength := 2

	for {
		// check length
		if detectionLength > 5 {
			return nil, ErrDetectionOverflow
		}

		// try read detection bytes
		header, err := d.reader.Peek(detectionLength)
		if err == io.EOF && len(header) != 0 {
			// an EOF with some data is unexpected
			return nil, io.ErrUnexpectedEOF
		} else if err != nil {
			return nil, err
		}

		// detect packet
		packetLength, packetType := DetectPacket(header)

		// on zero packet length:
		// increment detection length and try again
		if packetLength <= 0 {
			detectionLength++
			continue
		}

		// check read limit
		if d.Limit > 0 && int64(packetLength) > d.Limit {
			return nil, ErrReadLimitExceeded
		}

		// create packet
		pkt, err := packetType.New()
		if err != nil {
			return nil, err
		}

		// reset and eventually grow buffer
		d.buffer.Reset()
		d.buffer.Grow(packetLength)
		buf := d.buffer.Bytes()[0:packetLength]

		// read whole packet (will not return EOF)
		_, err = io.ReadFull(d.reader, buf)
		if err != nil {
			return nil, err
		}

		// decode buffer
		_, err = pkt.Decode(buf)
		if err != nil {
			return nil, err
		}

		return pkt, nil
	}
}

type Stream struct {
	Decoder
	Encoder
}

// NewStream creates a new Stream.
func NewStream(reader io.Reader, writer io.Writer) *Stream {
	return &Stream{
		Decoder: Decoder{
			reader: bufio.NewReader(reader),
		},
		Encoder: Encoder{
			writer: bufio.NewWriter(writer),
		},
	}
}