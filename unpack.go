package main

import (
	"fmt"
	"io"
	"os"
)

type part struct {
	name   string
	offset int
	size   int
}

var (
	parts = [3]part{
		{"kernel_heder", 0x0, 0x40},
		{"image", 0x40, 0x200000},
		{"squashfs", 0x200040, 0x350000},
	}
)

func Unpackfirmware() {

	//technically we only need the squashfs part but we unpack all three
	for _, part := range parts {

		outfile, err := os.OpenFile("/home/tobias/Tempunpacked/"+part.name, os.O_RDWR, os.ModeAppend)
		if err != nil {
			panic(err)
		}

		firmwareFile, _ := os.Open("/home/tobias/go/src/superCam/wyzeFirmware/demo_v2_4.9.5.36.bin/demo_v2_4.9.5.36.bin")

		fmt.Println("unpacking " + part.name)
		_, _ = firmwareFile.Seek(int64(part.offset), 0x0)
		readBytes := make([]byte, part.size)

		_, err = io.ReadFull(firmwareFile, readBytes)
		_, err = outfile.Write(readBytes)

		if err != nil {
			panic(err)
		}

		outfile.Sync()

		defer outfile.Close()

	}

}
