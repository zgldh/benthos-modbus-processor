package processor

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/snksoft/crc"
)

type CrcType string

const (
	X25        CrcType = "X25"
	CCITT      CrcType = "CCITT"
	CRC16      CrcType = "CRC16"
	MODBUS     CrcType = "MODBUS"
	XMODEM     CrcType = "XMODEM"
	XMODEM2    CrcType = "XMODEM2"
	CRC32      CrcType = "CRC32"
	IEEE       CrcType = "IEEE"
	Castagnoli CrcType = "Castagnoli"
	CRC32C     CrcType = "CRC32C"
	Koopman    CrcType = "Koopman"
	CRC64ISO   CrcType = "CRC64ISO"
	CRC64ECMA  CrcType = "CRC64ECMA"
)

type RawType string

const (
	Int8    RawType = "Int8"
	Int16   RawType = "Int16"
	Int32   RawType = "Int32"
	Float32 RawType = "Float32"
	Float64 RawType = "Float64"
)

type ValueType string

const (
	Int   RawType = "Int"
	Float RawType = "Float"
)

type ConfigDataLength struct {
	ByteIndex int
	NumBytes  int
	BigEndian bool
}

type ConfigCrcChecking struct {
	Enabled   bool
	Type      CrcType // X25, CCITT, MODBUS, CRC16, XMODEM, XMODEM2, CRC32, IEEE, Castagnoli, CRC32C, Koopman, CRC64ISO, CRC64ECMA
	BigEndian bool
}

type ConfigFieldAttributes struct {
	StartingAddress int
	RawType         RawType //  "Int8", "Int16", "Int32", "Float32", "Float64"
	BigEndian       bool
}

type ConfigFieldProperties struct {
	ValueType ValueType // "Int", "Float"
	Scale     float64
}

type ConfigField struct {
	Name       string
	Attributes ConfigFieldAttributes
	Properties ConfigFieldProperties
}

type ModbusProcessorConfig struct {
	BytesPerAddress int
	DataLength      ConfigDataLength
	CrcChecking     ConfigCrcChecking
	Fields          []ConfigField
}
type ModbusProcessor struct {
	config                    *ModbusProcessorConfig
	crcHash                   *crc.Hash
	crcHashMatchingBytesCount int
	logger                    *service.Logger
}

func newModbusProcessor(conf *service.ParsedConfig, logger *service.Logger) (*ModbusProcessor, error) {
	bytes_per_address, err := conf.FieldInt("bytes_per_address")
	if err != nil {
		bytes_per_address = 2
	}

	data_length__byte_index, err := conf.FieldInt("data_length", "byte_index")
	if err != nil {
		data_length__byte_index = 2
	}
	data_length__num_bytes, err := conf.FieldInt("data_length", "num_bytes")
	if err != nil {
		data_length__num_bytes = 2
	}
	data_length__big_endian, err := conf.FieldBool("data_length", "big_endian")
	if err != nil {
		data_length__big_endian = true
	}

	crc_checking__enabled, err := conf.FieldBool("crc_checking", "enabled")
	if err != nil {
		crc_checking__enabled = true
	}
	crc_checking__type, err := conf.FieldString("crc_checking", "type")
	if err != nil {
		crc_checking__type = string(CRC16)
	}

	crc_checking__big_endian, err := conf.FieldBool("crc_checking", "big_endian")
	if err != nil {
		crc_checking__big_endian = true
	}

	configFields, err := conf.FieldObjectList("fields")
	if err != nil {
		return nil, errors.New("fields are required")
	}
	fields := make([]ConfigField, len(configFields))

	for index, configField := range configFields {
		name, err := configField.FieldString("name")
		if err != nil {
			return nil, errors.New("field name is required")
		}
		attributes_starting_address, err := configField.FieldInt("attributes", "starting_address")
		if err != nil {
			return nil, fmt.Errorf("field '%v' attributes.starting_address is required", name)
		}
		attributes_raw_type, err := configField.FieldString("attributes", "raw_type")
		if err != nil {
			return nil, fmt.Errorf("field '%v' attributes.raw_type is required", name)
		}
		attributes_big_endian, err := configField.FieldBool("attributes", "big_endian")
		if err != nil {
			attributes_big_endian = true
		}

		properties_value_type, err := configField.FieldString("properties", "value_type")
		if err != nil {
			return nil, fmt.Errorf("field '%v' properties.value_type is required", name)
		}
		properties_scale, err := configField.FieldFloat("properties", "scale")
		if err != nil {
			properties_scale = 1
		}

		fields[index] = ConfigField{
			Name: name,
			Attributes: ConfigFieldAttributes{
				StartingAddress: attributes_starting_address,
				RawType:         RawType(attributes_raw_type),
				BigEndian:       attributes_big_endian,
			},
			Properties: ConfigFieldProperties{
				ValueType: ValueType(properties_value_type),
				Scale:     properties_scale,
			},
		}
	}
	crcParameter, err := getCrcParameterByType(CrcType(crc_checking__type))
	if err != nil {
		crcParameter = crc.CRC16
	}
	hash := crc.NewHash(crcParameter)
	return &ModbusProcessor{
		config: &ModbusProcessorConfig{
			BytesPerAddress: bytes_per_address,
			DataLength: ConfigDataLength{
				ByteIndex: data_length__byte_index,
				NumBytes:  data_length__num_bytes,
				BigEndian: data_length__big_endian,
			},
			CrcChecking: ConfigCrcChecking{
				Enabled:   crc_checking__enabled,
				Type:      CrcType(crc_checking__type),
				BigEndian: crc_checking__big_endian,
			},
			Fields: fields,
		},
		crcHash:                   hash,
		crcHashMatchingBytesCount: getCrcMatchingByteLength(CrcType(crc_checking__type)),
		logger:                    logger,
	}, nil
}

func init() {
	modbusConfigSpec := service.NewConfigSpec().
		Summary("Creates a processor for Modbus data.").
		Field(service.NewIntField("bytes_per_address").Advanced().Default(2)).
		Field(service.NewObjectField("data_length",
			service.NewIntField("byte_index").Default(0x02),
			service.NewIntField("num_bytes").Default(2),
			service.NewBoolField("big_endian").Default(true),
		).Advanced().Default(ConfigDataLength{2, 2, true})).
		Field(service.NewObjectField("crc_checking",
			service.NewBoolField("enabled").Default(true),
			service.NewStringEnumField("type",
				"X25", "CCITT", "MODBUS", "CRC16", "XMODEM", "XMODEM2", "CRC32", "IEEE", "Castagnoli", "CRC32C", "Koopman", "CRC64ISO", "CRC64ECMA",
			).Default(CRC16),
			service.NewBoolField("big_endian").Default(true),
		).Advanced().Default(ConfigCrcChecking{true, CRC16, true})).
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
		return newModbusProcessor(conf, mgr.Logger())
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
	// newBytes, err := json.Marshal(*r.config)
	// if err != nil {
	// 	panic(err)
	// }

	// CRC checking
	if r.config.CrcChecking.Enabled {
		isCrcChecked, err := r.getCrcCheckingResult(bytesContent)
		if err != nil {
			isCrcChecked = false
		}
		if isCrcChecked {
			m.MetaSet("modbus_crc_checked", "true")
		} else {
			m.MetaSet("modbus_crc_checked", "false")
			return nil, errors.New("CRC checking failed")
		}
	}

	m.SetBytes([]byte{})
	return []*service.Message{m}, nil
}

func (r *ModbusProcessor) Close(ctx context.Context) error {
	return nil
}

func (r *ModbusProcessor) getCrcCheckingResult(bytesContent []byte) (bool, error) {
	length := len(bytesContent)
	payloadBytes := bytesContent[0 : length-int(r.crcHashMatchingBytesCount)]
	calculatedCrc := r.crcHash.CalculateCRC(payloadBytes)

	var matchingCrc uint64
	matchingBytes := bytesContent[length-r.crcHashMatchingBytesCount : length]
	if r.config.CrcChecking.BigEndian {
		if r.crcHashMatchingBytesCount == 2 {
			matchingCrc = uint64(binary.BigEndian.Uint16(matchingBytes))
		} else if r.crcHashMatchingBytesCount == 4 {
			matchingCrc = uint64(binary.BigEndian.Uint32(matchingBytes))
		} else {
			matchingCrc = binary.BigEndian.Uint64(matchingBytes)
		}
	} else {
		if r.crcHashMatchingBytesCount == 2 {
			matchingCrc = uint64(binary.LittleEndian.Uint16(matchingBytes))
		} else if r.crcHashMatchingBytesCount == 4 {
			matchingCrc = uint64(binary.LittleEndian.Uint32(matchingBytes))
		} else {
			matchingCrc = binary.LittleEndian.Uint64(matchingBytes)
		}
	}
	return matchingCrc == calculatedCrc, nil
}

func getCrcParameterByType(typeName CrcType) (*crc.Parameters, error) {
	if X25 == typeName {
		return crc.X25, nil
	}
	if CCITT == typeName {
		return crc.CCITT, nil
	}
	if MODBUS == typeName {
		return &crc.Parameters{Width: 16, Polynomial: 0x8005, Init: 0xFFFF, ReflectIn: true, ReflectOut: true, FinalXor: 0x0}, nil
	}
	if CRC16 == typeName {
		return crc.CRC16, nil
	}
	if XMODEM == typeName {
		return crc.XMODEM, nil
	}
	if XMODEM2 == typeName {
		return crc.XMODEM2, nil
	}
	if CRC32 == typeName {
		return crc.CRC32, nil
	}
	if IEEE == typeName {
		return crc.IEEE, nil
	}
	if Castagnoli == typeName {
		return crc.Castagnoli, nil
	}
	if CRC32C == typeName {
		return crc.CRC32C, nil
	}
	if Koopman == typeName {
		return crc.Koopman, nil
	}
	if CRC64ISO == typeName {
		return crc.CRC64ISO, nil
	}
	if CRC64ECMA == typeName {
		return crc.CRC64ECMA, nil
	}
	return nil, fmt.Errorf("unknown crc type %s", typeName)
}

func getCrcMatchingByteLength(typeName CrcType) int {
	if X25 == typeName {
		return 2
	}
	if CCITT == typeName {
		return 2
	}
	if MODBUS == typeName {
		return 2
	}
	if CRC16 == typeName {
		return 2
	}
	if XMODEM == typeName {
		return 2
	}
	if XMODEM2 == typeName {
		return 2
	}
	if CRC32 == typeName {
		return 4
	}
	if IEEE == typeName {
		return 4
	}
	if Castagnoli == typeName {
		return 4
	}
	if CRC32C == typeName {
		return 4
	}
	if Koopman == typeName {
		return 4
	}
	if CRC64ISO == typeName {
		return 8
	}
	if CRC64ECMA == typeName {
		return 8
	}
	return 2
}
