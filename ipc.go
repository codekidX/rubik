package rubik

import (
	"fmt"
	"reflect"
	"time"

	"github.com/rubikorg/rubik/pkg"
)

// IpcMessage is the structure using which Rubik services
// communicates with each other
type IpcMessage struct {
	Type interface{}
	Func func(interface{})
}

type ipcModem struct {
	wsMap map[string]string
	msgRx map[string]IpcMessage
}

// Send transmits a message using the ipcModem to the given service
// the message is identified by the receiver using the msgType argument
func (ipc ipcModem) Send(msgType string, service string, message interface{}) {
	if s, ok := ipc.wsMap[service]; ok {
		txClient := NewClient(s, time.Second*30)
		ipcRxEn := IpcRxEntity{
			Message: msgType,
			Body:    message,
		}
		ipcRxEn.PointTo = "/rubik/msg/rx"
		r, err := txClient.Post(ipcRxEn)
		if err != nil {
			fmt.Printf("Message: %s, failed. Response: %s\n", msgType, r.StringBody)
		}
		return
	}
	pkg.ErrorMsg(fmt.Sprintf("%s is not present in this workspace", service))
}

// OnMessage registers a IpcMessage handler for the given
// message type
func (ipc ipcModem) OnMessage(msgType string, ipcMp IpcMessage) {
	if reflect.TypeOf(ipcMp.Type).Kind() != reflect.Ptr {
		panic(fmt.Errorf("OnMessage: %s has IpcMessage type as non-pointer", msgType))
	}
	ipc.msgRx[msgType] = ipcMp
}

var Ipc = ipcModem{
	wsMap: make(map[string]string),
	msgRx: make(map[string]IpcMessage),
}
