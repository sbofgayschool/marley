package utils

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
)

// IVFWriter is used to take RTP packets and write them to an IVF on disk
type IVFWriter struct {
	ioWriter             io.Writer
	width, height, count int
	baseTimestamp        uint64
	currentFrame         []byte
}

// New builds a new IVF writer
func NewIVFWriter(fileName string, width int, height int) (*IVFWriter, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	writer, err := NewWith(f, width, height)
	if err != nil {
		return nil, err
	}
	writer.ioWriter = f
	return writer, nil
}

// NewWith initialize a new IVF writer with an io.Writer output
func NewWith(out io.Writer, width int, height int) (*IVFWriter, error) {
	if out == nil {
		return nil, fmt.Errorf("file not opened")
	}

	writer := &IVFWriter{
		ioWriter: out,
		width:    width,
		height:   height,
	}
	if err := writer.writeHeader(); err != nil {
		return nil, err
	}
	return writer, nil
}

func (i *IVFWriter) writeHeader() error {
	header := make([]byte, 32)
	copy(header[0:], []byte("DKIF"))                             // DKIF
	binary.LittleEndian.PutUint16(header[4:], 0)                 // Version
	binary.LittleEndian.PutUint16(header[6:], 32)                // Header size
	copy(header[8:], []byte("VP80"))                             // FOURCC
	binary.LittleEndian.PutUint16(header[12:], uint16(i.width))  // Width in pixels
	binary.LittleEndian.PutUint16(header[14:], uint16(i.height)) // Height in pixels
	binary.LittleEndian.PutUint32(header[16:], 90000)            // Framerate denominator
	binary.LittleEndian.PutUint32(header[20:], 1)                // Framerate numerator
	binary.LittleEndian.PutUint32(header[24:], 900)              // Frame count, will be updated on first Close() call
	binary.LittleEndian.PutUint32(header[28:], 0)                // Unused

	_, err := i.ioWriter.Write(header)
	return err
}

// WriteRTP adds a new packet and writes the appropriate headers for it
func (i *IVFWriter) WriteRTP(packet *rtp.Packet) error {
	if i.ioWriter == nil {
		return fmt.Errorf("file not opened")
	}

	vp8Packet := codecs.VP8Packet{}
	if _, err := vp8Packet.Unmarshal(packet.Payload); err != nil {
		return err
	}

	i.currentFrame = append(i.currentFrame, vp8Packet.Payload[0:]...)

	if !packet.Marker {
		return nil
	} else if len(i.currentFrame) == 0 {
		return nil
	}

	frameHeader := make([]byte, 12)
	binary.LittleEndian.PutUint32(frameHeader[0:], uint32(len(i.currentFrame))) // Frame length
	if i.baseTimestamp == 0 {
		binary.LittleEndian.PutUint64(frameHeader[4:], 0)
		i.baseTimestamp = uint64(packet.Timestamp)
	} else {
		binary.LittleEndian.PutUint64(frameHeader[4:], uint64(packet.Timestamp)-i.baseTimestamp)
	}
	i.count++

	if _, err := i.ioWriter.Write(frameHeader); err != nil {
		return err
	} else if _, err := i.ioWriter.Write(i.currentFrame); err != nil {
		return err
	}

	i.currentFrame = nil
	return nil
}

// Close stops the recording
func (i *IVFWriter) Close() error {
	if i.ioWriter == nil {
		// Returns no error as it may be convenient to call
		// Close() multiple times
		return nil
	}

	defer func() {
		i.ioWriter = nil
	}()

	if ws, ok := i.ioWriter.(io.WriteSeeker); ok {
		// Update the framecount
		if _, err := ws.Seek(24, 0); err != nil {
			return err
		}
		buff := make([]byte, 4)
		binary.LittleEndian.PutUint32(buff, uint32(i.count))
		if _, err := ws.Write(buff); err != nil {
			return err
		}
	}

	if closer, ok := i.ioWriter.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}
