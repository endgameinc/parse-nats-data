package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	proto "github.com/gogo/protobuf/proto"
	spb "github.com/nats-io/nats-streaming-server/spb"
)

// Record types for subscription file
type recordType byte

const (
	subRecNew = recordType(iota) + 1
	subRecUpdate
	subRecDel
	subRecAck
	subRecMsg
)

// Record types for client store
const (
	addClient = recordType(iota) + 1
	delClient
)

func readNextBytes(file *os.File, number int) ([]byte, error) {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func readNextBytesAsUint32(f *os.File) (uint32, error) {
	b, err := readNextBytes(f, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func readNextBytesAsUint64(f *os.File) (uint64, error) {
	b, err := readNextBytes(f, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

func getMsgSize(d []byte) int {
	if len(d) <= 0 {
		return 0
	}
	size := int(uint(d[0]) | uint(d[1])<<8 | uint(d[2])<<16)
	return size
}

func getNextMsgOffsetFromIndex(idxFile *os.File) (uint32, uint64, error) {
	var err error

	seq, err := readNextBytesAsUint64(idxFile)
	if err != nil {
		return 0, 0, err
	}

	offset, err := readNextBytesAsUint64(idxFile)
	if err != nil {
		return 0, 0, err
	}

	timestamp, err := readNextBytesAsUint64(idxFile)
	if err != nil {
		return 0, 0, err
	}

	idxMsgSize, err := readNextBytesAsUint32(idxFile)
	if err != nil {
		return 0, 0, err
	}

	crc32, err := readNextBytesAsUint32(idxFile)
	if err != nil {
		return 0, 0, err
	}

	t := time.Unix(0, int64(timestamp))

	location, err := time.LoadLocation("UTC")
	if err != nil {
		return 0, 0, err
	}

	fmt.Printf("[%s] seq: %d | size: %d | offset: %d | crc32: %x", t.In(location), seq, idxMsgSize, offset, crc32)

	return idxMsgSize, offset, err
}

func printMsgAtOffset(datFile *os.File, size uint32, offset uint64) error {
	var err error

	bytes := make([]byte, size)

	_, err = datFile.Seek(int64(offset), 0)
	if err != nil {
		log.Fatal(err)
	}

	_, err = datFile.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}

	msgSize := getMsgSize(bytes)
	if int(msgSize) != int(size) {
		return errors.New(fmt.Sprintf("error: size mismatch (%u msg!= %u idx)", int(msgSize), int(size)))
	}
	return err
}

func parseMessages(idxFile *os.File, datFile *os.File) error {
	var err error
	size := uint32(0)
	offset := uint64(0)

	for {
		size, offset, err = getNextMsgOffsetFromIndex(idxFile)
		if err != nil {
			break
		}

		err = printMsgAtOffset(datFile, size, offset)
		if err != nil {
			break
		}
		fmt.Printf("\n")
	}

	return err
}

func checkVersion(f *os.File) error {
	fileVersion, err := readNextBytesAsUint32(f)
	if fileVersion != 1 || err != nil {
		return errors.New("invalid file version")
	}

	return nil
}

func parseData(idxPath string, datPath string) error {
	var err error

	idxFile, err := os.Open(idxPath)
	if err != nil {
		return err
	}

	datFile, err := os.Open(datPath)
	if err != nil {
		log.Fatalf("Cannot open file '%s' err:%s", datPath, err)
	}
	defer idxFile.Close()

	if checkVersion(idxFile) != nil {
		return errors.New("invalid index file")
	}

	if checkVersion(datFile) != nil {
		return errors.New("invalid dat file")
	}

	err = parseMessages(idxFile, datFile)
	return err
}

func parseSubRecNew(msgType string, msg *[]byte) {
	newSub := &spb.SubState{}
	err := proto.Unmarshal(*msg, newSub)
	if err != nil {
		fmt.Printf("unmarshalling error: %s\n", err)
		return
	}

	fmt.Printf("ID: %d '%s' Type: %s LastSent: %d qGroup: %s Inbox: %s AckInbox: %s MaxInFlight: %d AckWaitInSecs: %d Durable: %s IsDurable: %t IsClosed: %t\n",
		newSub.ID,
		newSub.ClientID,
		msgType,
		newSub.LastSent,
		newSub.QGroup,
		newSub.Inbox,
		newSub.AckInbox,
		newSub.MaxInFlight,
		newSub.AckWaitInSecs,
		newSub.DurableName,
		newSub.IsDurable,
		newSub.IsClosed)
}

func parseSubReqDel(msgType string, msg *[]byte) {
	delSub := &spb.SubStateDelete{}
	err := proto.Unmarshal(*msg, delSub)
	if err != nil {
		fmt.Printf("unmarshalling error: %s\n", err)
		return
	}
	fmt.Printf("ID: %d Tyep: %s\n", delSub.ID, msgType)
}

func parseSubReqAck(msgType string, msg *[]byte) {
	updateSub := &spb.SubStateUpdate{}
	err := proto.Unmarshal(*msg, updateSub)
	if err != nil {
		fmt.Printf("unmarshalling error: %s\n", err)
		return
	}
	fmt.Printf("ID: %d SeqNo: %d %s\n", updateSub.ID, updateSub.Seqno, msgType)
}

func parseAddClient(msgType string, msg *[]byte) {
	addClient := &spb.ClientInfo{}
	err := proto.Unmarshal(*msg, addClient)
	if err != nil {
		fmt.Printf("unmarshalling error: %s\n", err)
		return
	}
	fmt.Printf("CID: %s Inbox: %s ConnId: %d Protocol: %d PingInterval: %d PingMaxTimeout: %d %s\n", addClient.ID, addClient.HbInbox, addClient.ConnID, addClient.Protocol, addClient.PingInterval, addClient.PingMaxOut, msgType)
}

func parseDelClient(msgType string, msg *[]byte) {
	delClient := &spb.ClientInfo{}
	err := proto.Unmarshal(*msg, delClient)
	if err != nil {
		fmt.Printf("unmarshalling error: %s\n", err)
		return
	}
	fmt.Printf("CID: %s %s\n", delClient.ID, msgType)
	return
}

func parseClients(clientsPath string) error {
	var err error

	clientsFile, err := os.Open(clientsPath)
	if err != nil {
		return err
	}
	defer clientsFile.Close()

	if checkVersion(clientsFile) != nil {
		return errors.New("invalid sub file")
	}

	msgSizeAndType := []byte{}
	msg := []byte{}

	for {
		msgSizeAndType, err = readNextBytes(clientsFile, 4)
		if err != nil {
			return err
		}

		msgSize := getMsgSize(msgSizeAndType)
		_, err = readNextBytes(clientsFile, 4)
		if err != nil {
			return err
		}

		msgType := recordType(msgSizeAndType[3])
		msg, err = readNextBytes(clientsFile, msgSize)
		if err != nil {
			return err
		}

		switch msgType {
		case addClient:
			parseAddClient("addClient", &msg)
		case delClient:
			parseDelClient("delClient", &msg)
		default:
			return errors.New("unknown message type")
		}
	}

	return err
}

func parseSubscriptions(subPath string) error {
	var err error

	subFile, err := os.Open(subPath)
	if err != nil {
		return err
	}
	defer subFile.Close()

	if checkVersion(subFile) != nil {
		return errors.New("invalid sub file")
	}

	msgSizeAndType := []byte{}
	msg := []byte{}

	for true {

		msgSizeAndType, err = readNextBytes(subFile, 4)
		if err != nil {
			return err
		}

		msgSize := getMsgSize(msgSizeAndType)
		_, err = readNextBytes(subFile, 4)
		if err != nil {
			return err
		}

		msgType := recordType(msgSizeAndType[3])
		msg, err = readNextBytes(subFile, msgSize)
		if err != nil {
			return err
		}

		switch msgType {
		case subRecNew:
			parseSubRecNew("subRecNew", &msg)
		case subRecUpdate:
			parseSubRecNew("subRecUpdate", &msg)
		case subRecDel:
			parseSubReqDel("subRecDel", &msg)
		case subRecAck:
			parseSubReqAck("subRecAck", &msg)
		case subRecMsg:
			parseSubReqAck("subRecMsg", &msg)
		default:
			return errors.New("unknown message type")
		}
	}

	return err
}

var usageStr = `
Usage: %s [options]

Options:
	-d, --msgsDat		msgs.dat file
	-i, --msgsIdx		msgs.idx file
	-s, --subsDat		subs.dat file
	-c, --clientsDat	clients.dat file
`

func usage() {
	fmt.Printf("%s\n", fmt.Sprintf(usageStr, os.Args[0]))
	os.Exit(0)
}

func main() {
	var err error

	msgDatFile := ""
	flag.StringVar(&msgDatFile, "d", "", "msgs.dat file")
	flag.StringVar(&msgDatFile, "msgsDat", "", "msgs.dat file")

	msgIdxFile := ""
	flag.StringVar(&msgIdxFile, "i", "", "msgs.idx file")
	flag.StringVar(&msgIdxFile, "msgsIdx", "", "msgs.idx file")

	subDatFile := ""
	flag.StringVar(&subDatFile, "s", "", "subs.dat file")
	flag.StringVar(&subDatFile, "subsDat", "", "subs.dat file")

	clientsDatFile := ""
	flag.StringVar(&clientsDatFile, "c", "", "clients.dat file")
	flag.StringVar(&clientsDatFile, "clientsDat", "", "clients.dat file")

	flag.Usage = usage
	flag.Parse()

	if flag.NFlag() < 1 {
		usage()
		os.Exit(0)
	}

	if subDatFile != "" {
		err = parseSubscriptions(subDatFile)
	} else if clientsDatFile != "" {
		err = parseClients(clientsDatFile)
	} else {
		if msgDatFile == "" {
			fmt.Println("error: missing msgs.dat")
			os.Exit(1)
		}
		if msgIdxFile == "" {
			fmt.Println("error: missing msgs.idx")
			os.Exit(1)
		}
		err = parseData(msgIdxFile, msgDatFile)
	}

	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
