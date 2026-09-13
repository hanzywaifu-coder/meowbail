package media

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
)

const (
	signatureLocalFile    = 0x04034b50
	signatureCentralDir   = 0x02014b50
	signatureEndOfCentral = 0x06054b50
	compressionStored     = 0
	versionNeeded         = 20
	localHeaderSize       = 30
	centralHeaderSize     = 46
	eocdSize              = 22
)

type zipEntry struct {
	name []byte
	crc  uint32
	size uint32
	data []byte
}

type whatsappZipWriter struct {
	entries []*zipEntry
}

func NewWhatsAppZipWriter() *whatsappZipWriter {
	return &whatsappZipWriter{}
}

func (w *whatsappZipWriter) AddFile(fileName string, data []byte) {
	crc := crc32.ChecksumIEEE(data)
	w.entries = append(w.entries, &zipEntry{
		name: []byte(fileName),
		crc:  crc,
		size: uint32(len(data)),
		data: data,
	})
}

// Bytes mengompilasi ZIP container murni (metode Store, flags 0, no data descriptors)
// yang secara ketat kompatibel dengan parser ZIP C++ WhatsApp Android & iOS
func (w *whatsappZipWriter) Bytes() []byte {
	var buf bytes.Buffer
	type centralRef struct {
		name   []byte
		crc    uint32
		size   uint32
		offset uint32
	}
	var centrals []centralRef

	// 1. Tulis Local File Headers & Payload
	for _, entry := range w.entries {
		localOffset := uint32(buf.Len())
		centrals = append(centrals, centralRef{
			name:   entry.name,
			crc:    entry.crc,
			size:   entry.size,
			offset: localOffset,
		})

		header := make([]byte, localHeaderSize+len(entry.name))
		binary.LittleEndian.PutUint32(header[0:4], signatureLocalFile)
		binary.LittleEndian.PutUint16(header[4:6], versionNeeded)
		binary.LittleEndian.PutUint16(header[6:8], 0) // General Purpose Bit Flag = 0 (No descriptor!)
		binary.LittleEndian.PutUint16(header[8:10], compressionStored)
		binary.LittleEndian.PutUint16(header[10:12], 0) // Last mod time
		binary.LittleEndian.PutUint16(header[12:14], 0) // Last mod date
		binary.LittleEndian.PutUint32(header[14:18], entry.crc)
		binary.LittleEndian.PutUint32(header[18:22], entry.size) // Compressed size
		binary.LittleEndian.PutUint32(header[22:26], entry.size) // Uncompressed size
		binary.LittleEndian.PutUint16(header[26:28], uint16(len(entry.name)))
		binary.LittleEndian.PutUint16(header[28:30], 0) // Extra field length
		copy(header[localHeaderSize:], entry.name)

		buf.Write(header)
		buf.Write(entry.data)
	}

	centralStartOffset := uint32(buf.Len())

	// 2. Tulis Central Directory Headers
	for _, c := range centrals {
		cHeader := make([]byte, centralHeaderSize+len(c.name))
		binary.LittleEndian.PutUint32(cHeader[0:4], signatureCentralDir)
		binary.LittleEndian.PutUint16(cHeader[4:6], versionNeeded) // Version made by
		binary.LittleEndian.PutUint16(cHeader[6:8], versionNeeded) // Version needed
		binary.LittleEndian.PutUint16(cHeader[8:10], 0)            // Bit flag = 0
		binary.LittleEndian.PutUint16(cHeader[10:12], compressionStored)
		binary.LittleEndian.PutUint16(cHeader[12:14], 0)
		binary.LittleEndian.PutUint16(cHeader[14:16], 0)
		binary.LittleEndian.PutUint32(cHeader[16:20], c.crc)
		binary.LittleEndian.PutUint32(cHeader[20:24], c.size)
		binary.LittleEndian.PutUint32(cHeader[24:28], c.size)
		binary.LittleEndian.PutUint16(cHeader[28:30], uint16(len(c.name)))
		binary.LittleEndian.PutUint16(cHeader[30:32], 0)
		binary.LittleEndian.PutUint16(cHeader[32:34], 0)
		binary.LittleEndian.PutUint16(cHeader[34:36], 0)
		binary.LittleEndian.PutUint16(cHeader[36:38], 0)
		binary.LittleEndian.PutUint32(cHeader[38:42], 0) // External file attributes
		binary.LittleEndian.PutUint32(cHeader[42:46], c.offset)
		copy(cHeader[centralHeaderSize:], c.name)

		buf.Write(cHeader)
	}

	centralDirSize := uint32(buf.Len()) - centralStartOffset

	// 3. Tulis End of Central Directory (EOCD)
	eocd := make([]byte, eocdSize)
	binary.LittleEndian.PutUint32(eocd[0:4], signatureEndOfCentral)
	binary.LittleEndian.PutUint16(eocd[4:6], 0)
	binary.LittleEndian.PutUint16(eocd[6:8], 0)
	binary.LittleEndian.PutUint16(eocd[8:10], uint16(len(w.entries)))
	binary.LittleEndian.PutUint16(eocd[10:12], uint16(len(w.entries)))
	binary.LittleEndian.PutUint32(eocd[12:16], centralDirSize)
	binary.LittleEndian.PutUint32(eocd[16:20], centralStartOffset)
	binary.LittleEndian.PutUint16(eocd[20:22], 0)

	buf.Write(eocd)
	return buf.Bytes()
}
