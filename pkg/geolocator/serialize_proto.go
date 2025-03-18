package geolocator

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (g *GeoLocator) SerializeProto(p protoreflect.ProtoMessage, initial []byte) ([]byte, error) {

	if p == nil {
		panic("protobuf is nil")
	}
	b, err := proto.Marshal(p)
	if err != nil {
		return nil, err
	}
	if initial != nil {
		b = append(initial, append([]byte{byte(len(b))}, b...)...)
	}

	return b, nil

}
