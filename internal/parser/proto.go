package parser

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
)

// protoProfile represents the protobuf profile structure
type protoProfile struct {
	SampleType   []*protoValueType
	Sample       []*protoSample
	Mapping      []*protoMapping
	Location     []*protoLocation
	Function     []*protoFunction
	StringTable  []string
	DropFrames   int64
	KeepFrames   int64
	TimeNanos    int64
	DurationNanos int64
	PeriodType   *protoValueType
	Period       int64
	Comment      []int64
	DefaultSampleType int64
}

type protoValueType struct {
	Type int64
	Unit int64
}

type protoSample struct {
	LocationIdx []uint64
	Value       []int64
	Label       []*protoLabel
}

type protoLabel struct {
	Key int64
	Str int64
	Num int64
}

type protoMapping struct {
	Id              uint64
	MemoryStart     uint64
	MemoryLimit     uint64
	FileOffset      uint64
	Filename        int64
	BuildId         int64
	HasFunctions    bool
	HasFilenames    bool
	HasLineNumbers  bool
	HasInlineFrames bool
}

type protoLocation struct {
	Id       uint64
	MappingId uint64
	Address  uint64
	Line     []*protoLine
}

type protoLine struct {
	FunctionId uint64
	Line       int64
}

type protoFunction struct {
	Id         uint64
	Name       int64
	SystemName int64
	Filename   int64
	StartLine  int64
}

// parseProtoProfile parses a gzipped protobuf profile
func parseProtoProfile(data []byte) (*protoProfile, error) {
	// Try to decompress gzip
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		// Not gzipped, try raw
		return parseRawProto(data)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}

	return parseRawProto(decompressed)
}

// parseRawProto parses protobuf data (simplified parser)
func parseRawProto(data []byte) (*protoProfile, error) {
	prof := &protoProfile{
		StringTable: []string{""}, // First string is always empty
	}

	reader := bytes.NewReader(data)
	for reader.Len() > 0 {
		// Read tag and wire type
		tag, err := readVarint(reader)
		if err != nil {
			break
		}

		fieldNum := int(tag >> 3)
		wireType := tag & 7

		switch wireType {
		case 0: // Varint
			v, _ := readVarint(reader)
			prof.parseVarintField(fieldNum, v)
		case 1: // 64-bit
			var v uint64
			binary.Read(reader, binary.LittleEndian, &v)
			prof.parse64BitField(fieldNum, v)
		case 2: // Length-delimited
			length, _ := readVarint(reader)
			data := make([]byte, length)
			io.ReadFull(reader, data)
			prof.parseLengthDelimitedField(fieldNum, data)
		case 5: // 32-bit
			var v uint32
			binary.Read(reader, binary.LittleEndian, &v)
		}
	}

	return prof, nil
}

func readVarint(r io.ByteReader) (uint64, error) {
	var result uint64
	var shift uint
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			return result, nil
		}
		shift += 7
	}
}

func (p *protoProfile) parseVarintField(fieldNum int, value uint64) {
	switch fieldNum {
	case 10: // drop_frames
		p.DropFrames = int64(value)
	case 11: // keep_frames
		p.KeepFrames = int64(value)
	case 13: // period
		p.Period = int64(value)
	case 15: // comment
		p.Comment = append(p.Comment, int64(value))
	case 16: // default_sample_type
		p.DefaultSampleType = int64(value)
	}
}

func (p *protoProfile) parse64BitField(fieldNum int, value uint64) {
	switch fieldNum {
	case 9: // time_nanos
		p.TimeNanos = int64(value)
	}
}

func (p *protoProfile) parseLengthDelimitedField(fieldNum int, data []byte) {
	reader := bytes.NewReader(data)

	switch fieldNum {
	case 1: // sample_type
		vt := &protoValueType{}
		vt.parseFrom(reader)
		p.SampleType = append(p.SampleType, vt)
	case 2: // sample
		sample := &protoSample{}
		sample.parseFrom(reader)
		p.Sample = append(p.Sample, sample)
	case 3: // mapping
		mapping := &protoMapping{}
		mapping.parseFrom(reader)
		p.Mapping = append(p.Mapping, mapping)
	case 4: // location
		loc := &protoLocation{}
		loc.parseFrom(reader)
		p.Location = append(p.Location, loc)
	case 5: // function
		fn := &protoFunction{}
		fn.parseFrom(reader)
		p.Function = append(p.Function, fn)
	case 6: // string_table
		p.StringTable = append(p.StringTable, string(data))
	case 7: // drop_frames (repeated string)
		// Ignored for now
	case 8: // keep_frames (repeated string)
		// Ignored for now
	case 14: // period_type
		vt := &protoValueType{}
		vt.parseFrom(reader)
		p.PeriodType = vt
	}
}

func (vt *protoValueType) parseFrom(r *bytes.Reader) {
	for r.Len() > 0 {
		tag, _ := readVarint(r)
		fieldNum := int(tag >> 3)
		wireType := tag & 7

		if wireType == 0 {
			v, _ := readVarint(r)
			if fieldNum == 1 {
				vt.Type = int64(v)
			} else if fieldNum == 2 {
				vt.Unit = int64(v)
			}
		} else if wireType == 2 {
			length, _ := readVarint(r)
			r.Seek(int64(length), 1) // Skip
		}
	}
}

func (s *protoSample) parseFrom(r *bytes.Reader) {
	for r.Len() > 0 {
		tag, _ := readVarint(r)
		fieldNum := int(tag >> 3)
		wireType := tag & 7

		if wireType == 0 {
			v, _ := readVarint(r)
			if fieldNum == 1 {
				s.LocationIdx = append(s.LocationIdx, v)
			} else if fieldNum == 2 {
				s.Value = append(s.Value, int64(v))
			}
		} else if wireType == 2 {
			length, _ := readVarint(r)
			data := make([]byte, length)
			io.ReadFull(r, data)
			if fieldNum == 3 { // label
				label := &protoLabel{}
				label.parseFrom(data)
				s.Label = append(s.Label, label)
			}
		}
	}
}

func (l *protoLabel) parseFrom(data []byte) {
	r := bytes.NewReader(data)
	for r.Len() > 0 {
		tag, _ := readVarint(r)
		fieldNum := int(tag >> 3)
		wireType := tag & 7

		if wireType == 0 {
			v, _ := readVarint(r)
			if fieldNum == 1 {
				l.Key = int64(v)
			} else if fieldNum == 3 {
				l.Num = int64(v)
			}
		} else if wireType == 2 {
			length, _ := readVarint(r)
			data := make([]byte, length)
			io.ReadFull(r, data)
			if fieldNum == 2 {
				l.Str = int64(binary.LittleEndian.Uint64(data))
			}
		}
	}
}

func (m *protoMapping) parseFrom(r *bytes.Reader) {
	for r.Len() > 0 {
		tag, _ := readVarint(r)
		fieldNum := int(tag >> 3)
		wireType := tag & 7

		if wireType == 0 {
			v, _ := readVarint(r)
			switch fieldNum {
			case 1:
				m.Id = v
			case 2:
				m.MemoryStart = v
			case 3:
				m.MemoryLimit = v
			case 4:
				m.FileOffset = v
			case 6:
				m.Filename = int64(v)
			case 7:
				m.BuildId = int64(v)
			case 8:
				m.HasFunctions = v > 0
			case 9:
				m.HasFilenames = v > 0
			case 10:
				m.HasLineNumbers = v > 0
			case 11:
				m.HasInlineFrames = v > 0
			}
		} else if wireType == 2 {
			length, _ := readVarint(r)
			r.Seek(int64(length), 1) // Skip
		}
	}
}

func (l *protoLocation) parseFrom(r *bytes.Reader) {
	for r.Len() > 0 {
		tag, _ := readVarint(r)
		fieldNum := int(tag >> 3)
		wireType := tag & 7

		if wireType == 0 {
			v, _ := readVarint(r)
			if fieldNum == 1 {
				l.Id = v
			} else if fieldNum == 2 {
				l.MappingId = v
			} else if fieldNum == 3 {
				l.Address = v
			}
		} else if wireType == 2 {
			length, _ := readVarint(r)
			data := make([]byte, length)
			io.ReadFull(r, data)
			if fieldNum == 4 { // line
				line := &protoLine{}
				line.parseFrom(data)
				l.Line = append(l.Line, line)
			}
		}
	}
}

func (l *protoLine) parseFrom(data []byte) {
	r := bytes.NewReader(data)
	for r.Len() > 0 {
		tag, _ := readVarint(r)
		fieldNum := int(tag >> 3)
		wireType := tag & 7

		if wireType == 0 {
			v, _ := readVarint(r)
			if fieldNum == 1 {
				l.FunctionId = v
			} else if fieldNum == 2 {
				l.Line = int64(v)
			}
		}
	}
}

func (f *protoFunction) parseFrom(r *bytes.Reader) {
	for r.Len() > 0 {
		tag, _ := readVarint(r)
		fieldNum := int(tag >> 3)
		wireType := tag & 7

		if wireType == 0 {
			v, _ := readVarint(r)
			if fieldNum == 1 {
				f.Id = v
			} else if fieldNum == 3 {
				f.Name = int64(v)
			} else if fieldNum == 4 {
				f.SystemName = int64(v)
			} else if fieldNum == 5 {
				f.Filename = int64(v)
			} else if fieldNum == 6 {
				f.StartLine = int64(v)
			}
		} else if wireType == 2 {
			length, _ := readVarint(r)
			r.Seek(int64(length), 1) // Skip
		}
	}
}

// detectFromSampleType detects profile type from sample_type field
func detectFromSampleType(prof *protoProfile) (ProfileType, error) {
	if len(prof.SampleType) == 0 {
		return "", fmt.Errorf("no sample type in profile")
	}

	// Get sample type names from string table
	for _, st := range prof.SampleType {
		if int(st.Type) < len(prof.StringTable) && int(st.Unit) < len(prof.StringTable) {
			typeName := prof.StringTable[st.Type]
			unitName := prof.StringTable[st.Unit]

			// Detect based on type and unit
			switch typeName {
			case "cpu", "samples":
				if unitName == "nanoseconds" || unitName == "seconds" {
					return TypeCPU, nil
				}
			case "alloc_objects", "inuse_objects":
				return TypeHeap, nil
			case "goroutines":
				return TypeGoroutine, nil
			case "contentions", "lock_duration":
				return TypeMutex, nil
			}

			// Fallback to unit detection
			switch unitName {
			case "nanoseconds", "seconds", "milliseconds":
				// Could be CPU or mutex, check type name
				if typeName == "cpu" || typeName == "samples" {
					return TypeCPU, nil
				}
			case "bytes", "objects":
				return TypeHeap, nil
			case "count", "goroutines":
				return TypeGoroutine, nil
			case "lock_ns", "contentions":
				return TypeMutex, nil
			}
		}
	}

	// Try to detect from period type
	if prof.PeriodType != nil {
		if int(prof.PeriodType.Type) < len(prof.StringTable) {
			typeName := prof.StringTable[prof.PeriodType.Type]
			if typeName == "cpu" || typeName == "samples" {
				return TypeCPU, nil
			}
		}
	}

	// Default: try to infer from sample values
	if len(prof.Sample) > 0 {
		// If samples have 2 values, likely heap (objects + bytes)
		if len(prof.Sample[0].Value) == 2 {
			return TypeHeap, nil
		}
	}

	return "", fmt.Errorf("unknown profile type")
}
