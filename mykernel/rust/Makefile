#
# mykernel/rust/Makefile
#
# Copyright (C) 2017 - 2021 Vinay Chandra
#
# Permission is hereby granted, free of charge, to any person
# obtaining a copy of this software and associated documentation
# files (the "Software"), to deal in the Software without
# restriction, including without limitation the rights to use, copy,
# modify, merge, publish, distribute, sublicense, and/or sell copies
# of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be
# included in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
# EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
# MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
# NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
# HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
# WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
# DEALINGS IN THE SOFTWARE.
#
# This file is part of the BOOTBOOT Protocol package.
# @brief An example Makefile for sample kernel
#
#

all: mykernel.x86_64.elf

# Kernel build
mykernel.x86_64.elf: src/**
	objcopy -O elf64-x86-64 -B i386 -I binary font.psf font.o
	cargo xbuild --target ./triplets/mykernel-x86.json
	# Using this causes the kernel to exceed the 2 MB limit
	# cargo build -Z build-std=core,alloc --target ./triplets/$mykernel-x86.json
	cp ./target/mykernel-x86/debug/mykernel-rust mykernel.x86_64.elf
