package collect

import (
	"encoding/binary"
	"io"
	"os"

	"gopkg.in/music-theory.v0/key"
)

// parseMIDI reads tempo (BPM) and key signature from a MIDI file's meta events.
// No external dependency — parses the binary format directly.
// Returns tempo in BPM (0 if not found) and key (zero value if not found).
func parseMIDI(path string) (bpm int, k key.Key, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, key.Key{}, err
	}
	defer f.Close()

	// Read MIDI header chunk
	var hdr [14]byte
	if _, err = io.ReadFull(f, hdr[:]); err != nil {
		return 0, key.Key{}, err
	}
	// MThd magic
	if string(hdr[0:4]) != "MThd" {
		return 0, key.Key{}, nil
	}
	numTracks := int(binary.BigEndian.Uint16(hdr[10:12]))

	// Walk tracks looking for meta events
	for track := 0; track < numTracks; track++ {
		var chunkHdr [8]byte
		if _, err = io.ReadFull(f, chunkHdr[:]); err != nil {
			break
		}
		if string(chunkHdr[0:4]) != "MTrk" {
			break
		}
		chunkLen := int(binary.BigEndian.Uint32(chunkHdr[4:8]))
		chunkData := make([]byte, chunkLen)
		if _, err = io.ReadFull(f, chunkData); err != nil {
			break
		}

		bpm, k = parseMIDITrack(chunkData, bpm, k)
		if bpm > 0 && k.Root != 0 {
			break // got what we need
		}
	}
	return bpm, k, nil
}

// parseMIDITrack scans a track's raw bytes for tempo and key signature meta events.
func parseMIDITrack(data []byte, bpm int, k key.Key) (int, key.Key) {
	i := 0
	for i < len(data) {
		// Read variable-length delta time
		for i < len(data) {
			b := data[i]
			i++
			if b&0x80 == 0 {
				break
			}
		}
		if i >= len(data) {
			break
		}

		status := data[i]
		i++

		// Meta event: 0xFF
		if status == 0xFF {
			if i >= len(data) {
				break
			}
			metaType := data[i]
			i++

			// Read variable-length meta event length
			metaLen := 0
			for i < len(data) {
				b := data[i]
				i++
				metaLen = (metaLen << 7) | int(b&0x7F)
				if b&0x80 == 0 {
					break
				}
			}

			if i+metaLen > len(data) {
				break
			}
			metaData := data[i : i+metaLen]
			i += metaLen

			switch metaType {
			case 0x51: // Set Tempo — 3 bytes, microseconds per beat
				if len(metaData) == 3 && bpm == 0 {
					uspb := int(metaData[0])<<16 | int(metaData[1])<<8 | int(metaData[2])
					if uspb > 0 {
						bpm = 60_000_000 / uspb
					}
				}

			case 0x59: // Key Signature — 2 bytes: sharps/flats, major/minor
				if len(metaData) == 2 && k.Root == 0 {
					k = midiKeySignature(int8(metaData[0]), metaData[1])
				}
			}
			continue
		}

		// Skip non-meta events by inferring data length from status byte
		i += midiEventDataLen(status, data, i)
	}
	return bpm, k
}

// midiKeySignature converts MIDI key signature bytes to a music-theory key.
// sf: number of sharps (positive) or flats (negative). mi: 0=major, 1=minor.
func midiKeySignature(sf int8, mi byte) key.Key {
	// Order of sharps: F C G D A E B
	// Order of flats:  B E A D G C F
	noteNames := []string{"Cb", "Gb", "Db", "Ab", "Eb", "Bb", "F", "C", "G", "D", "A", "E", "B", "F#", "C#"}
	idx := int(sf) + 7
	if idx < 0 { idx = 0 }
	if idx >= len(noteNames) { idx = len(noteNames) - 1 }
	mode := "major"
	if mi == 1 { mode = "minor" }
	return key.Of(noteNames[idx] + " " + mode)
}

// midiEventDataLen returns how many data bytes follow a given status byte.
func midiEventDataLen(status byte, data []byte, pos int) int {
	switch status & 0xF0 {
	case 0x80, 0x90, 0xA0, 0xB0, 0xE0:
		return 2
	case 0xC0, 0xD0:
		return 1
	case 0xF0:
		switch status {
		case 0xF0, 0xF7: // SysEx — variable length
			length := 0
			for pos < len(data) {
				b := data[pos]
				pos++
				length++
				if b == 0xF7 {
					break
				}
			}
			return length
		case 0xF2:
			return 2
		case 0xF3:
			return 1
		default:
			return 0
		}
	}
	return 0
}
