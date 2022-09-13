(Work In Progress)
=================
Benthos with Modbus Processor plugin
======================

## Build

```sh
go build
```

Alternatively build it as a Docker image with:

```sh
go mod vendor
docker build . -t benthos-modbus-processor
```

## Usage

`docker run --rm -v /path/to/yaml:/benthos.yaml zgldh/benthos-modbus-processor:v1.0.2`

```yaml
pipeline:
  processors:
    - modbus:
        bytes_per_address: 2      # Optional, default 2. How many bytes per address for fields.
        data_length:              # Bytes length not Address length.
          byte_index: 0x02        # Optional, default 0x02. The index of the first byte for data length.
          bytes_num: 1            # Optional, default 1. How many bytes to define the data length. 1,2,4,8
          big_endian: true        # Optional, default true.
        crc_checking:
          enabled: true           # Optional, default true. The CRC matching value is the last 2, 4 or 8  bytes. It will take all bytes before the CRC matching value to get calculated CRC value.  It will throw exception if enabled and crc checking failed.
          type: MODBUS            # Optional, default MODBUS. Options: X25, CCITT, MODBUS, CRC16, XMODEM, XMODEM2, CRC32, IEEE, Castagnoli, CRC32C, Koopman, CRC64ISO, CRC64ECMA
          big_endian: false       # Optional, default true
        fields:
          -
            name: "SwitchOn"
            attributes:
              starting_address: 0x00
              raw_type: "Int16"   # Int8, Int16, Int32, Int64, UInt8, UInt16, UInt32, UInt64, Float32, Float64
              big_endian: true    # Optional, default true.
            properties:
              mapping: root = rawValue == 240
          -
            name: "F"
            attributes:
              starting_address: 0x04
              raw_type: "UInt16"
              big_endian: true      # Optional, default true.
            properties:
              si_unit: "Hz"
              mapping: root = rawValue / 10   # Optional, default empty. 
                                              # The mapping input `rawValue` will be the raw_type that conveted from bytes array. 
                                              # The mapping result `root` will be converted to the value_type.
          -
            name: "AU"
            attributes:
              starting_address: 0x08
              raw_type: "UInt16"
            properties:
              si_unit: "V"
          -
            name: "A_Power"
            attributes:
              starting_address: 14
              raw_type: "UInt32"
              big_endian: true    # Optional, default true.
            properties:
              si_unit: "kW·h"
              mapping: root = rawValue / 1000
```


```json
Meta Data
{
  "modbus_crc_checked": true,
  "modbus_slave_address": 1,
  "modbus_function": 3,
  "modbus_data_length": 123,
}

Payload
{
  "SwitchOn": {
    "Value": true,
    "SIUnit": ""
  },
  "F": {
    "Value": 49.9,
    "SIUnit": "Hz"
  },
  "AU": {
    "Value": 230,
    "SIUnit": "V"
  },
  "A_Power": {
    "Value": 775.65,
    "SIUnit": "kW·h"
  }
}
```