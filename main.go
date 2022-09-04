package main

import (
	"context"

	"github.com/benthosdev/benthos/v4/public/service"

	// Import all standard Benthos components
	_ "github.com/benthosdev/benthos/v4/public/components/all"

	// Add your plugin packages here
	_ "github.com/zgldh/benthos_modbus_processor/processor"
)

func main() {
	service.RunCLI(context.Background())
}
