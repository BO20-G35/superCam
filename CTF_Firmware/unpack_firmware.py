#!/usr/local/bin/python3.7
import sys

firmware_file = sys.argv[1]

class FirmwarePart:
    def __init__(self, name, offset, size):
        self.name = name
        self.offset = offset
        self.size = size

firmware_parts = [
    FirmwarePart("uimage_header", 0x0, 0x40),
    FirmwarePart("uimage_kernel", 0x40, 0x200000),
    FirmwarePart("squashfs", 0x200040, 0x350000),
]


f = open(firmware_file, "rb")
for part in firmware_parts:
    outfile = open(part.name, "wb")
    f.seek(part.offset, 0)
    data = f.read(part.size)
    outfile.write(data)
    outfile.close()

