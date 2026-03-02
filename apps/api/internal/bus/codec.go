package bus

import (
	"encoding/json"
	"fmt"
)

type CodecType string

const (
	CodecJSON  CodecType = "json"
	CodecProto CodecType = "proto"
	CodecBytes CodecType = "bytes"
)

type Codec interface {
	Name() CodecType
	Encode(v any) ([]byte, error)
	Decode(data []byte, v any) error
}

type JSONCodec struct{}

func (JSONCodec) Name() CodecType { return CodecJSON }

func (JSONCodec) Encode(v any) ([]byte, error) { return json.Marshal(v) }

func (JSONCodec) Decode(data []byte, v any) error { return json.Unmarshal(data, v) }

type CodecRegistry struct {
	codecs map[CodecType]Codec
}

func NewRegistry(codecs ...Codec) *CodecRegistry {
	m := make(map[CodecType]Codec, len(codecs))
	for _, c := range codecs {
		m[c.Name()] = c
	}
	return &CodecRegistry{codecs: m}
}

func (r *CodecRegistry) Get(cType CodecType) (Codec, error) {
	c, ok := r.codecs[cType]
	if !ok {
		return nil, fmt.Errorf("unknown codec: %s", cType)
	}
	return c, nil
}
