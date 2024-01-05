package build

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/40t/go-sniffer/plugSrc/mongodb/build/bson"
	"io"
	"time"
)

func GetNowStr(isClient bool) string {
	var msg string
	layout := "01/02 15:04:05.000000"
	msg += time.Now().Format(layout)
	if isClient {
		msg += "| cli -> ser |"
	} else {
		msg += "| ser -> cli |"
	}
	return msg
}

func ReadInt32(r io.Reader) (n int32) {
	binary.Read(r, binary.LittleEndian, &n)
	return
}

func ReadInt64(r io.Reader) int64 {
	var n int64
	binary.Read(r, binary.LittleEndian, &n)
	return n
}

func ReadString(r io.Reader) string {

	var result []byte
	var b = make([]byte, 1)
	for {

		_, err := r.Read(b)

		if err != nil {
			panic(err)
		}

		if b[0] == '\x00' {
			break
		}

		result = append(result, b[0])
	}

	return string(result)
}

func ReadBson2Json(r io.Reader) string {

	//read len
	docLen := ReadInt32(r)
	if docLen == 0 {
		return ""
	}

	//document []byte
	docBytes := make([]byte, int(docLen))
	binary.LittleEndian.PutUint32(docBytes, uint32(docLen))
	if _, err := io.ReadFull(r, docBytes[4:]); err != nil {
		panic(err)
	}

	//resolve document
	var bsn bson.M
	err := bson.Unmarshal(docBytes, &bsn)
	if err != nil {
		panic(err)
	}

	//format to Json
	jsonStr, err := json.Marshal(bsn)
	if err != nil {
		return fmt.Sprintf("{\"error\":%s}", err.Error())
	}
	return string(jsonStr)
}

func ReadCString(message []byte) string {
	index := bytes.IndexByte(message, 0x00)
	return string(message[:index])
}

func ParseOpMsgSections(message []byte) {
	// 解析 OP_MSG 消息的 section
	for len(message) > 0 {
		kind := message[0]
		message = message[1:]

		switch kind {
		case 0:
			// Body section
			if len(message) < 4 {
				fmt.Println("Invalid OP_MSG Body section: message too short")
				return
			}
			bodySize := binary.LittleEndian.Uint32(message[:4])
			message = message[4:]
			if len(message) < int(bodySize) {
				fmt.Println("Invalid OP_MSG Body section: message too short")
				fmt.Println(string(message))
				return
			}
			body := message[:bodySize]
			message = message[bodySize:]

			// 解析 Body section
			var bodyDoc interface{}
			err := bson.Unmarshal(body, &bodyDoc)
			if err != nil {
				fmt.Printf("Failed to unmarshal body section: %v\n", err)
			}
			fmt.Printf("OP_MSG Body: %v\n", bodyDoc)

		case 1:
			// Document Sequence section
			sequenceSize := binary.LittleEndian.Uint32(message[:4])
			sequenceIdentifier := ReadCString(message[4:])
			sequence := message[4+len(sequenceIdentifier)+1 : 4+len(sequenceIdentifier)+1+int(sequenceSize)]
			message = message[4+len(sequenceIdentifier)+1+int(sequenceSize):]

			// 解析 Document Sequence section
			var sequenceDocs []interface{}
			for len(sequence) > 0 {
				var doc interface{}
				docSize := binary.LittleEndian.Uint32(sequence[:4])
				err := bson.Unmarshal(sequence[4:4+docSize], &doc)
				if err != nil {
					fmt.Printf("Failed to unmarshal document in sequence: %v\n", err)
				}
				sequenceDocs = append(sequenceDocs, doc)
				sequence = sequence[4+docSize:]
			}
			fmt.Printf("OP_MSG Document Sequence Identifier: %s\n", sequenceIdentifier)
			fmt.Printf("OP_MSG Document Sequence: %v\n", sequenceDocs)

		default:
			fmt.Printf("Unknown section kind: %d\n", kind)
		}
	}
}
