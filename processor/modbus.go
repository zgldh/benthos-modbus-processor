package processor

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
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
	Int64   RawType = "Int64"
	UInt8   RawType = "UInt8"
	UInt16  RawType = "UInt16"
	UInt32  RawType = "UInt32"
	UInt64  RawType = "UInt64"
	Float32 RawType = "Float32"
	Float64 RawType = "Float64"
)

type ConfigDataLength struct {
	ByteIndex int
	BytesNum  int
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
	SIUnit  string
	Mapping *bloblang.Executor
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

type FieldValue struct {
	Value  interface{}
	SIUnit string
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
	data_length__bytes_num, err := conf.FieldInt("data_length", "bytes_num")
	if err != nil {
		data_length__bytes_num = 2
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

		properties_si_unit, err := configField.FieldString("properties", "si_unit")
		if err != nil {
			properties_si_unit = ""
		}

		properties_mapping, err := configField.FieldBloblang("properties", "mapping")
		if err != nil {
			properties_mapping = nil
		}

		fields[index] = ConfigField{
			Name: name,
			Attributes: ConfigFieldAttributes{
				StartingAddress: attributes_starting_address,
				RawType:         RawType(attributes_raw_type),
				BigEndian:       attributes_big_endian,
			},
			Properties: ConfigFieldProperties{
				SIUnit:  properties_si_unit,
				Mapping: properties_mapping,
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
				BytesNum:  data_length__bytes_num,
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
			service.NewIntField("bytes_num").Default(1).LintRule(`root = if false == [1,2,4,8].contains(this.number()) { [ "data_length.bytes_num can only be 1, 2, 4 or 8." ] }`),
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
				service.NewStringEnumField("raw_type", "Int8", "Int16", "Int32", "Int64", "UInt8", "UInt16", "UInt32", "UInt64", "Float32", "Float64"), // TODO
				service.NewBoolField("big_endian").Default(true),
			),
			service.NewObjectField("properties",
				service.NewStringField("si_unit").Optional(),
				service.NewBloblangField("mapping").Optional(),
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
	err = r.processCRC(bytesContent, m)
	if err != nil {
		return nil, err
	}

	// Slave address
	slaveAddress := uint(bytesContent[0])
	m.MetaSet("modbus_slave_address", fmt.Sprintf("%v", slaveAddress))

	// Function number
	functionNumber := uint(bytesContent[1])
	m.MetaSet("modbus_function", fmt.Sprintf("%v", functionNumber))

	// Data length
	dataLength, err := r.processDataLength(bytesContent, m)
	if err != nil {
		return nil, err
	}

	// Data fields parsing
	result, err := r.processDataFields(bytesContent, m, dataLength)
	if err != nil {
		return nil, err
	}

	m.SetStructured(result)
	return []*service.Message{m}, nil
}

func (r *ModbusProcessor) Close(ctx context.Context) error {
	return nil
}

func (r *ModbusProcessor) processCRC(bytesContent []byte, m *service.Message) error {
	if r.config.CrcChecking.Enabled {
		isCrcChecked, err := r.getCrcCheckingResult(bytesContent)
		if err != nil {
			isCrcChecked = false
		}
		if isCrcChecked {
			m.MetaSet("modbus_crc_checked", "true")
		} else {
			m.MetaSet("modbus_crc_checked", "false")
			return errors.New("CRC checking failed")
		}
	}
	return nil
}

func (r *ModbusProcessor) processDataLength(bytesContent []byte, m *service.Message) (uint64, error) {
	var length uint64
	var err error
	lengthRawBytes := bytesContent[r.config.DataLength.ByteIndex : r.config.DataLength.ByteIndex+r.config.DataLength.BytesNum]
	if r.config.DataLength.BigEndian {
		if r.config.DataLength.BytesNum == 1 {
			length = uint64(lengthRawBytes[0])
		} else if r.config.DataLength.BytesNum == 2 {
			length = uint64(binary.BigEndian.Uint16(lengthRawBytes))
		} else if r.config.DataLength.BytesNum == 4 {
			length = uint64(binary.BigEndian.Uint32(lengthRawBytes))

		} else if r.config.DataLength.BytesNum == 8 {
			length = binary.BigEndian.Uint64(lengthRawBytes)
		} else {
			err = errors.New("Data length parse failed")
		}
	} else {
		if r.config.DataLength.BytesNum == 1 {
			length = uint64(lengthRawBytes[0])
		} else if r.config.DataLength.BytesNum == 2 {
			length = uint64(binary.LittleEndian.Uint16(lengthRawBytes))
		} else if r.config.DataLength.BytesNum == 4 {
			length = uint64(binary.LittleEndian.Uint32(lengthRawBytes))

		} else if r.config.DataLength.BytesNum == 8 {
			length = binary.LittleEndian.Uint64(lengthRawBytes)
		} else {
			err = errors.New("Data length parse failed")
		}
	}

	if err != nil {
		return 0, err
	} else {
		m.MetaSet("modbus_data_length", fmt.Sprintf("%v", length))
		return length, nil
	}
}

func (r *ModbusProcessor) processDataFields(bytesContent []byte, m *service.Message, bytesLength uint64) (map[string]interface{}, error) {
	var result map[string]interface{} = map[string]interface{}{}
	var err error
	var byteOrder binary.ByteOrder
	offset := r.config.DataLength.ByteIndex + r.config.DataLength.BytesNum

	for _, field := range r.config.Fields {
		rawBytesNum := getRawTypeByteLength(field.Attributes.RawType)
		sliceStart := field.Attributes.StartingAddress*r.config.BytesPerAddress + offset
		sliceEnd := sliceStart + rawBytesNum
		bytesSlice := bytesContent[sliceStart:sliceEnd]
		var rawValue float64
		var fieldValue FieldValue

		if field.Attributes.BigEndian {
			byteOrder = binary.BigEndian
		} else {
			byteOrder = binary.LittleEndian
		}

		buff := bytes.NewReader(bytesSlice)
		if field.Attributes.RawType == Int8 {
			var readedValue int8
			err = binary.Read(buff, byteOrder, &readedValue)
			rawValue = float64(readedValue)
		} else if field.Attributes.RawType == Int16 {
			var readedValue int16
			err = binary.Read(buff, byteOrder, &readedValue)
			rawValue = float64(readedValue)
		} else if field.Attributes.RawType == Int32 {
			var readedValue int32
			err = binary.Read(buff, byteOrder, &readedValue)
			rawValue = float64(readedValue)
		} else if field.Attributes.RawType == Int64 {
			var readedValue int64
			err = binary.Read(buff, byteOrder, &readedValue)
			rawValue = float64(readedValue)
		} else if field.Attributes.RawType == UInt8 {
			var readedValue uint8
			err = binary.Read(buff, byteOrder, &readedValue)
			rawValue = float64(readedValue)
		} else if field.Attributes.RawType == UInt16 {
			var readedValue uint16
			err = binary.Read(buff, byteOrder, &readedValue)
			rawValue = float64(readedValue)
		} else if field.Attributes.RawType == UInt32 {
			var readedValue uint32
			err = binary.Read(buff, byteOrder, &readedValue)
			rawValue = float64(readedValue)
		} else if field.Attributes.RawType == UInt64 {
			var readedValue uint64
			err = binary.Read(buff, byteOrder, &readedValue)
			rawValue = float64(readedValue)
		} else if field.Attributes.RawType == Float32 {
			var readedValue float32
			err = binary.Read(buff, byteOrder, &readedValue)
			rawValue = float64(readedValue)
		} else if field.Attributes.RawType == Float64 {
			var readedValue float64
			err = binary.Read(buff, byteOrder, &readedValue)
			rawValue = float64(readedValue)
		}
		if err != nil {
			return nil, err
		}

		if field.Properties.Mapping != nil {
			mappedResult, err := field.Properties.Mapping.Query(map[string]interface{}{
				"rawValue": rawValue,
			})
			if err != nil {
				return nil, err
			}
			fieldValue = FieldValue{
				Value: mappedResult,
			}
		} else {
			fieldValue = FieldValue{
				Value: rawValue,
			}
		}

		if field.Properties.SIUnit != "" {
			fieldValue.SIUnit = field.Properties.SIUnit
		}

		result[field.Name] = fieldValue
	}
	return result, nil
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

func getRawTypeByteLength(rawType RawType) int {
	if rawType == Int8 {
		return 1
	}
	if rawType == Int16 {
		return 2
	}
	if rawType == Int32 {
		return 4
	}
	if rawType == Int64 {
		return 8
	}
	if rawType == UInt8 {
		return 1
	}
	if rawType == UInt16 {
		return 2
	}
	if rawType == UInt32 {
		return 4
	}
	if rawType == UInt64 {
		return 8
	}
	if rawType == Float32 {
		return 4
	}
	if rawType == Float64 {
		return 8
	}
	return 1
}
func getRawValueByType(rawType RawType) interface{} {
	if rawType == Int8 {
		return int8(0)
	}
	if rawType == Int16 {
		return int16(0)
	}
	if rawType == Int32 {
		return int32(0)
	}
	if rawType == Int64 {
		return int64(0)
	}
	if rawType == UInt8 {
		return uint8(0)
	}
	if rawType == UInt16 {
		return uint16(0)
	}
	if rawType == UInt32 {
		return uint32(0)
	}
	if rawType == UInt64 {
		return uint64(0)
	}
	if rawType == Float32 {
		return float32(0)
	}
	if rawType == Float64 {
		return float64(0)
	}
	return int8(0)
}
