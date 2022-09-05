package input

import (
	"bytes"
	"context"

	"github.com/benthosdev/benthos/v4/public/service"
)

type ConfigSpecDataLength struct {
	starting_address int
	num_bytes        int
	big_endian       bool
}

type ConfigSpecCrc16 struct {
	enabled    bool
	big_endian bool
}
type ModbusProcessor struct {
	logger *service.Logger
}

func newModbusProcessor(conf *service.ParsedConfig, logger *service.Logger) *ModbusProcessor {
	return &ModbusProcessor{
		logger: logger,
	}
}

func init() {
	modbusConfigSpec := service.NewConfigSpec().
		Summary("Creates a processor for Modbus data.").
		Field(service.NewIntField("bytes_per_address").Advanced().Default(2)).
		Field(service.NewObjectField("data_length",
			service.NewIntField("starting_address").Default(0x02),
			service.NewIntField("num_bytes").Default(2),
			service.NewBoolField("big_endian").Default(true),
		).Advanced().Default(ConfigSpecDataLength{2, 2, true})).
		Field(service.NewObjectField("crc16",
			service.NewBoolField("enabled").Default(true),
			service.NewBoolField("big_endian").Default(true),
		).Advanced().Default(ConfigSpecCrc16{true, true})).
		Field(service.NewObjectListField("fields",
			service.NewStringField("name"),
			service.NewObjectField("attributes",
				service.NewIntField("starting_address"),
				service.NewStringEnumField("raw_type", "Int8", "Int16", "Int32", "Float32", "Float64"), // TODO
				service.NewBoolField("big_endian").Default(true),
			),
			service.NewObjectField("properties",
				service.NewStringEnumField("value_type", "Int", "Float"), // TODO
				service.NewFloatField("scale").Default(1),
			),
		))

	constructor := func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
		return newModbusProcessor(conf, mgr.Logger()), nil
	}

	err := service.RegisterProcessor("modbus", modbusConfigSpec, constructor)
	if err != nil {
		panic(err)
	}
}

func (r *ModbusProcessor) Process(ctx context.Context, m *service.Message) (service.MessageBatch, error) {
	bytesContent, err := m.AsBytes()
	if err != nil {
		return nil, err
	}

	newBytes := make([]byte, len(bytesContent))
	for i, b := range bytesContent {
		newBytes[len(newBytes)-i-1] = b
	}

	if bytes.Equal(newBytes, bytesContent) {
		r.logger.Infof("Woah! This is like totally a palindrome: %s", bytesContent)
	}

	m.SetBytes(newBytes)
	return []*service.Message{m}, nil
}

func (r *ModbusProcessor) Close(ctx context.Context) error {
	return nil
}
