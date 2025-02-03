package serializer

import (
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
)

func WriteProtobufToJsonFile(message proto.Message, filename string) error {
	data := ProtobufToJSON(message)

	err := os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("can't write binary data to file: %s", err)
	}
	return nil
}

func WriteProtobufToBinaryFile(message proto.Message, filename string) error {
	data, err := proto.Marshal(message)

	if err != nil {
		return fmt.Errorf("can't marshal the message to binary: %s", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("can't write binary data to file: %s", err)
	}
	return nil
}

func ReadProtobufFromBinaryFile(filename string, message proto.Message) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("can't read binary data from file: %s", err)
	}

	err = proto.Unmarshal(data, message)
	if err != nil {
		return fmt.Errorf("can't unmarshal binary data to message: %s", err)
	}
	return nil
}
