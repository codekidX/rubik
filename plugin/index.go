package plugin

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"

	"github.com/rubikorg/rubik"
)

func GetPluginData() (*rubik.PluginData, error) {
	rubikSock := "/tmp/rubik.sock"
	os.Remove(rubikSock)
	l, err := net.Listen("unix", rubikSock)
	if err != nil {
		return nil, err
	}

	for {
		fd, err := l.Accept()
		if err != nil {
			return nil, err
		}

		for {
			var bytebuf bytes.Buffer
			buf := make([]byte, 4096)
			_, err := fd.Read(buf)
			_, err = bytebuf.Write(buf)
			if err != nil {
				return nil, err
			}

			var pd rubik.PluginData
			dec := gob.NewDecoder(&bytebuf)
			err = dec.Decode(&pd)
			if err != nil {
				return nil, err
			}
			fmt.Printf("Server got: %#v\n", pd)
			return &pd, nil
		}
	}

}
