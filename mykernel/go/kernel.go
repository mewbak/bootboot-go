/*
 * mykernel/go/kernel.go
 *
 * Copyright (C) 2017 - 2021 bzt (bztsrc@gitlab)
 *
 * Permission is hereby granted, free of charge, to any person
 * obtaining a copy of this software and associated documentation
 * files (the "Software"), to deal in the Software without
 * restriction, including without limitation the rights to use, copy,
 * modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be
 * included in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 * EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 * MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 * NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
 * HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 *
 * This file is part of the BOOTBOOT Protocol package.
 * @brief A sample BOOTBOOT compatible kernel
 *
 */

package main

import "unsafe"

/* Go sucks big time:
 * 1. no include (WTF they dare to call this a C killer language without a pre-compiler?)
 * 2. no union support nor pre-compiler conditionals for the arch specific struct part (no comment...)
 * 3. as soon as you declare a type struct (not use it, just declare) Go will generate tons of unresolved runtime references
 * 4. importing font from another object file? Forget it... neither CGO nor .syso work for non-function labels
 * 5. accessing a linker defined label? Forget it... Use constants and keep them synced with the linker script
 * 6. even the "official" bare-metal-gophers example on github miserably fails to compile with the latest Go compiler...
 * 7. if you finally manage to compile it, the resulting file is going to be twice the size of the C version!
 *
 * So we do dirty hacks here, only using pointers to access bootboot struct members, addresses by constants
 * instead of linker provided labels, font embedded in a string, and we need an Assembly runtime too...
 */

const (
	BOOTBOOT_MMIO = 0xfffffffff8000000 /* memory mapped IO virtual address */
	BOOTBOOT_FB   = 0xfffffffffc000000 /* frame buffer virtual address */
	BOOTBOOT_INFO = 0xffffffffffe00000 /* bootboot struct virtual address */
	BOOTBOOT_ENV  = 0xffffffffffe01000 /* environment string virtual address */
	BOOTBOOT_CORE = 0xffffffffffe02000 /* core loadable segment start */
)

/* our "font", Go simply can't import variables from another object (and it can't embed byte arrays either, only strings...) */
const glyphs = "\x00\x00\xda\x02\x80\x82\x02\x80\x82\x02\x80\xb6\x00\x00\x00\x00\x00\x00\x7e\x81\xa5\x81\x81\xbd\x99\x81\x81\x7e\x00\x00\x00\x00\x00\x00\x7e\xff\xdb\xff\xff\xc3\xe7\xff\xff\x7e\x00\x00\x00\x00\x00\x00\x00\x00\x6c\xfe\xfe\xfe\xfe\x7c\x38\x10\x00\x00\x00\x00\x00\x00\x00\x00\x10\x38\x7c\xfe\x7c\x38\x10\x00\x00\x00\x00\x00\x00\x00\x00\x18\x3c\x3c\xe7\xe7\xe7\x18\x18\x3c\x00\x00\x00\x00\x00\x00\x00\x18\x3c\x7e\xff\xff\x7e\x18\x18\x3c\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x18\x3c\x3c\x18\x00\x00\x00\x00\x00\x00\xff\xff\xff\xff\xff\xff\xe7\xc3\xc3\xe7\xff\xff\xff\xff\xff\xff\x00\x00\x00\x00\x00\x3c\x66\x42\x42\x66\x3c\x00\x00\x00\x00\x00\xff\xff\xff\xff\xff\xc3\x99\xbd\xbd\x99\xc3\xff\xff\xff\xff\xff\x00\x00\x1e\x0e\x1a\x32\x78\xcc\xcc\xcc\xcc\x78\x00\x00\x00\x00\x00\x00\x3c\x66\x66\x66\x66\x3c\x18\x7e\x18\x18\x00\x00\x00\x00\x00\x00\x3f\x33\x3f\x30\x30\x30\x30\x70\xf0\xe0\x00\x00\x00\x00\x00\x00\x7f\x63\x7f\x63\x63\x63\x63\x67\xe7\xe6\xc0\x00\x00\x00\x00\x00\x00\x18\x18\xdb\x3c\xe7\x3c\xdb\x18\x18\x00\x00\x00\x00\x00\x80\xc0\xe0\xf0\xf8\xfe\xf8\xf0\xe0\xc0\x80\x00\x00\x00\x00\x00\x02\x06\x0e\x1e\x3e\xfe\x3e\x1e\x0e\x06\x02\x00\x00\x00\x00\x00\x00\x18\x3c\x7e\x18\x18\x18\x7e\x3c\x18\x00\x00\x00\x00\x00\x00\x00\x66\x66\x66\x66\x66\x66\x66\x00\x66\x66\x00\x00\x00\x00\x00\x00\x7f\xdb\xdb\xdb\x7b\x1b\x1b\x1b\x1b\x1b\x00\x00\x00\x00\x00\x7c\xc6\x60\x38\x6c\xc6\xc6\x6c\x38\x0c\xc6\x7c\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xfe\xfe\xfe\xfe\x00\x00\x00\x00\x00\x00\x18\x3c\x7e\x18\x18\x18\x7e\x3c\x18\x7e\x00\x00\x00\x00\x00\x00\x18\x3c\x7e\x18\x18\x18\x18\x18\x18\x18\x00\x00\x00\x00\x00\x00\x18\x18\x18\x18\x18\x18\x18\x7e\x3c\x18\x00\x00\x00\x00\x00\x00\x00\x00\x00\x18\x0c\xfe\x0c\x18\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x30\x60\xfe\x60\x30\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xc0\xc0\xc0\xfe\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x28\x6c\xfe\x6c\x28\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x10\x38\x38\x7c\x7c\xfe\xfe\x00\x00\x00\x00\x00\x00\x00\x00\x00\xfe\xfe\x7c\x7c\x38\x38\x10\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x18\x3c\x3c\x3c\x18\x18\x18\x00\x18\x18\x00\x00\x00\x00\x00\x66\x66\x66\x24\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x6c\x6c\xfe\x6c\x6c\x6c\xfe\x6c\x6c\x00\x00\x00\x00\x18\x18\x7c\xc6\xc2\xc0\x7c\x06\x06\x86\xc6\x7c\x18\x18\x00\x00\x00\x00\x00\x00\xc2\xc6\x0c\x18\x30\x60\xc6\x86\x00\x00\x00\x00\x00\x00\x38\x6c\x6c\x38\x76\xdc\xcc\xcc\xcc\x76\x00\x00\x00\x00\x00\x30\x30\x30\x20\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x0c\x18\x30\x30\x30\x30\x30\x30\x18\x0c\x00\x00\x00\x00\x00\x00\x30\x18\x0c\x0c\x0c\x0c\x0c\x0c\x18\x30\x00\x00\x00\x00\x00\x00\x00\x00\x00\x66\x3c\xff\x3c\x66\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x18\x18\x7e\x18\x18\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x18\x18\x18\x30\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xfe\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x18\x18\x00\x00\x00\x00\x00\x00\x00\x00\x02\x06\x0c\x18\x30\x60\xc0\x80\x00\x00\x00\x00\x00\x00\x38\x6c\xc6\xc6\xd6\xd6\xc6\xc6\x6c\x38\x00\x00\x00\x00\x00\x00\x18\x38\x78\x18\x18\x18\x18\x18\x18\x7e\x00\x00\x00\x00\x00\x00\x7c\xc6\x06\x0c\x18\x30\x60\xc0\xc6\xfe\x00\x00\x00\x00\x00\x00\x7c\xc6\x06\x06\x3c\x06\x06\x06\xc6\x7c\x00\x00\x00\x00\x00\x00\x0c\x1c\x3c\x6c\xcc\xfe\x0c\x0c\x0c\x1e\x00\x00\x00\x00\x00\x00\xfe\xc0\xc0\xc0\xfc\x06\x06\x06\xc6\x7c\x00\x00\x00\x00\x00\x00\x38\x60\xc0\xc0\xfc\xc6\xc6\xc6\xc6\x7c\x00\x00\x00\x00\x00\x00\xfe\xc6\x06\x06\x0c\x18\x30\x30\x30\x30\x00\x00\x00\x00\x00\x00\x7c\xc6\xc6\xc6\x7c\xc6\xc6\xc6\xc6\x7c\x00\x00\x00\x00\x00\x00\x7c\xc6\xc6\xc6\x7e\x06\x06\x06\x0c\x78\x00\x00\x00\x00\x00\x00\x00\x00\x18\x18\x00\x00\x00\x18\x18\x00\x00\x00\x00\x00\x00\x00\x00\x00\x18\x18\x00\x00\x00\x18\x18\x30\x00\x00\x00\x00\x00\x00\x00\x06\x0c\x18\x30\x60\x30\x18\x0c\x06\x00\x00\x00\x00\x00\x00\x00\x00\x00\x7e\x00\x00\x7e\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x60\x30\x18\x0c\x06\x0c\x18\x30\x60\x00\x00\x00\x00\x00\x00\x7c\xc6\xc6\x0c\x18\x18\x18\x00\x18\x18\x00\x00\x00\x00\x00\x00\x00\x7c\xc6\xc6\xde\xde\xde\xdc\xc0\x7c\x00\x00\x00\x00\x00\x00\x10\x38\x6c\xc6\xc6\xfe\xc6\xc6\xc6\xc6\x00\x00\x00\x00\x00\x00\xfc\x66\x66\x66\x7c\x66\x66\x66\x66\xfc\x00\x00\x00\x00\x00\x00\x3c\x66\xc2\xc0\xc0\xc0\xc0\xc2\x66\x3c\x00\x00\x00\x00\x00\x00\xf8\x6c\x66\x66\x66\x66\x66\x66\x6c\xf8\x00\x00\x00\x00\x00\x00\xfe\x66\x62\x68\x78\x68\x60\x62\x66\xfe\x00\x00\x00\x00\x00\x00\xfe\x66\x62\x68\x78\x68\x60\x60\x60\xf0\x00\x00\x00\x00\x00\x00\x3c\x66\xc2\xc0\xc0\xde\xc6\xc6\x66\x3a\x00\x00\x00\x00\x00\x00\xc6\xc6\xc6\xc6\xfe\xc6\xc6\xc6\xc6\xc6\x00\x00\x00\x00\x00\x00\x3c\x18\x18\x18\x18\x18\x18\x18\x18\x3c\x00\x00\x00\x00\x00\x00\x1e\x0c\x0c\x0c\x0c\x0c\xcc\xcc\xcc\x78\x00\x00\x00\x00\x00\x00\xe6\x66\x66\x6c\x78\x78\x6c\x66\x66\xe6\x00\x00\x00\x00\x00\x00\xf0\x60\x60\x60\x60\x60\x60\x62\x66\xfe\x00\x00\x00\x00\x00\x00\xc6\xee\xfe\xfe\xd6\xc6\xc6\xc6\xc6\xc6\x00\x00\x00\x00\x00\x00\xc6\xe6\xf6\xfe\xde\xce\xc6\xc6\xc6\xc6\x00\x00\x00\x00\x00\x00\x7c\xc6\xc6\xc6\xc6\xc6\xc6\xc6\xc6\x7c\x00\x00\x00\x00\x00\x00\xfc\x66\x66\x66\x7c\x60\x60\x60\x60\xf0\x00\x00\x00\x00\x00\x00\x7c\xc6\xc6\xc6\xc6\xc6\xc6\xd6\xde\x7c\x0c\x0e\x00\x00\x00\x00\xfc\x66\x66\x66\x7c\x6c\x66\x66\x66\xe6\x00\x00\x00\x00\x00\x00\x7c\xc6\xc6\x60\x38\x0c\x06\xc6\xc6\x7c\x00\x00\x00\x00\x00\x00\x7e\x7e\x5a\x18\x18\x18\x18\x18\x18\x3c\x00\x00\x00\x00\x00\x00\xc6\xc6\xc6\xc6\xc6\xc6\xc6\xc6\xc6\x7c\x00\x00\x00\x00\x00\x00\xc6\xc6\xc6\xc6\xc6\xc6\xc6\x6c\x38\x10\x00\x00\x00\x00\x00\x00\xc6\xc6\xc6\xc6\xd6\xd6\xd6\xfe\xee\x6c\x00\x00\x00\x00\x00\x00\xc6\xc6\x6c\x7c\x38\x38\x7c\x6c\xc6\xc6\x00\x00\x00\x00\x00\x00\x66\x66\x66\x66\x3c\x18\x18\x18\x18\x3c\x00\x00\x00\x00\x00\x00\xfe\xc6\x86\x0c\x18\x30\x60\xc2\xc6\xfe\x00\x00\x00\x00\x00\x00\x3c\x30\x30\x30\x30\x30\x30\x30\x30\x3c\x00\x00\x00\x00\x00\x00\x00\x80\xc0\xe0\x70\x38\x1c\x0e\x06\x02\x00\x00\x00\x00\x00\x00\x3c\x0c\x0c\x0c\x0c\x0c\x0c\x0c\x0c\x3c\x00\x00\x00\x00\x10\x38\x6c\xc6\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff\x00\x00\x30\x30\x18\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x78\x0c\x7c\xcc\xcc\xcc\x76\x00\x00\x00\x00\x00\x00\xe0\x60\x60\x78\x6c\x66\x66\x66\x66\x7c\x00\x00\x00\x00\x00\x00\x00\x00\x00\x7c\xc6\xc0\xc0\xc0\xc6\x7c\x00\x00\x00\x00\x00\x00\x1c\x0c\x0c\x3c\x6c\xcc\xcc\xcc\xcc\x76\x00\x00\x00\x00\x00\x00\x00\x00\x00\x7c\xc6\xfe\xc0\xc0\xc6\x7c\x00\x00\x00\x00\x00\x00\x38\x6c\x64\x60\xf0\x60\x60\x60\x60\xf0\x00\x00\x00\x00\x00\x00\x00\x00\x00\x76\xcc\xcc\xcc\xcc\xcc\x7c\x0c\xcc\x78\x00\x00\x00\xe0\x60\x60\x6c\x76\x66\x66\x66\x66\xe6\x00\x00\x00\x00\x00\x00\x18\x18\x00\x38\x18\x18\x18\x18\x18\x3c\x00\x00\x00\x00\x00\x00\x06\x06\x00\x0e\x06\x06\x06\x06\x06\x06\x66\x66\x3c\x00\x00\x00\xe0\x60\x60\x66\x6c\x78\x78\x6c\x66\xe6\x00\x00\x00\x00\x00\x00\x38\x18\x18\x18\x18\x18\x18\x18\x18\x3c\x00\x00\x00\x00\x00\x00\x00\x00\x00\xec\xfe\xd6\xd6\xd6\xd6\xc6\x00\x00\x00\x00\x00\x00\x00\x00\x00\xdc\x66\x66\x66\x66\x66\x66\x00\x00\x00\x00\x00\x00\x00\x00\x00\x7c\xc6\xc6\xc6\xc6\xc6\x7c\x00\x00\x00\x00\x00\x00\x00\x00\x00\xdc\x66\x66\x66\x66\x66\x7c\x60\x60\xf0\x00\x00\x00\x00\x00\x00\x76\xcc\xcc\xcc\xcc\xcc\x7c\x0c\x0c\x1e\x00\x00\x00\x00\x00\x00\xdc\x76\x66\x60\x60\x60\xf0\x00\x00\x00\x00\x00\x00\x00\x00\x00\x7c\xc6\x60\x38\x0c\xc6\x7c\x00\x00\x00\x00\x00\x00\x10\x30\x30\xfc\x30\x30\x30\x30\x36\x1c\x00\x00\x00\x00\x00\x00\x00\x00\x00\xcc\xcc\xcc\xcc\xcc\xcc\x76\x00\x00\x00\x00\x00\x00\x00\x00\x00\x66\x66\x66\x66\x66\x3c\x18\x00\x00\x00\x00\x00\x00\x00\x00\x00\xc6\xc6\xd6\xd6\xd6\xfe\x6c\x00\x00\x00\x00\x00\x00\x00\x00\x00\xc6\x6c\x38\x38\x38\x6c\xc6\x00\x00\x00\x00\x00\x00\x00\x00\x00\xc6\xc6\xc6\xc6\xc6\xc6\x7e\x06\x0c\xf8\x00\x00\x00\x00\x00\x00\xfe\xcc\x18\x30\x60\xc6\xfe\x00\x00\x00\x00\x00\x00\x0e\x18\x18\x18\x70\x18\x18\x18\x18\x0e\x00\x00\x00\x00\x00\x00\x18\x18\x18\x18\x18\x18\x18\x18\x18\x18\x18\x18\x00\x00\x00\x00\x70\x18\x18\x18\x0e\x18\x18\x18\x18\x70\x00\x00\x00\x00\x00\x00\x76\xdc\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x10\x38\x6c\xc6\xc6\xc6\xfe\x00\x00\x00\x00\x00"

/******************************************
 * Entry point, called by BOOTBOOT Loader *
 ******************************************/
// Here is why each pragma below is needed:
//
// cgo_export_static - Exports the symbol for the external linker to see
//
// linkname - Go symbols are a combination of their package and func name.
// So this func is actually main._start This creates a link to this function
// just as _start so that the exported C name knows which functions to call.
//
// nosplit - Go has growable stacks. Since we haven't told the runtime how
// big the limine provided stack is disable the stack growth check.
//
//go:cgo_export_static _start _start
//go:linkname _start _start
//go:nosplit
func _start() {
	/*** NOTE: this code runs on all cores in parallel ***/
	var x, y int
	var w int = (int)(*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_INFO + 0x34)))) // bootboot.fb_width
	var h int = (int)(*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_INFO + 0x38)))) // bootboot.fb_height
	var s int = (int)(*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_INFO + 0x3c)))) // bootboot.fb_scanline

	if s > 0 {
		// cross-hair to see screen dimension detected correctly
		for y = 0; y < h; y++ {
			*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_FB) + uintptr(s*y+w*2))) = 0x00FFFFFF
		}
		for x = 0; x < w; x++ {
			*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_FB) + uintptr(s*y/2+x*4))) = 0x00FFFFFF
		}

		// red, green, blue boxes in order
		for y = 0; y < 20; y++ {
			for x = 0; x < 20; x++ {
				*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_FB) + uintptr(s*(y+20)+(x+20)*4))) = 0x00FF0000
			}
		}
		for y = 0; y < 20; y++ {
			for x = 0; x < 20; x++ {
				*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_FB) + uintptr(s*(y+20)+(x+50)*4))) = 0x0000FF00
			}
		}
		for y = 0; y < 20; y++ {
			for x = 0; x < 20; x++ {
				*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_FB) + uintptr(s*(y+20)+(x+80)*4))) = 0x000000FF
			}
		}

		// say hello
		puts("Hello from a simple BOOTBOOT kernel")
	}

	// hang for now
	for {
	}
}

/**************************
 * Display text on screen *
 **************************/
func puts(s string) {
	var i, x, y, kx, line, mask, offs int
	kx = 0
	for i = 0; i < len(s); i++ {
		offs = kx * 9 * 4
		for y = 0; y < 16; y++ {
			line = offs
			mask = 1 << 7
			for x = 0; x < 8; x++ {
				if (int(glyphs[int(s[i])*16+y]) & mask) == mask {
					*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_FB) + uintptr(line))) = 0x00FFFFFF
				} else {
					*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_FB) + uintptr(line))) = 0
				}
				line = line + 4
				mask >>= 1
			}
			*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_FB) + uintptr(line))) = 0
			offs = offs + (int)(*(*uint32)(unsafe.Pointer(uintptr(BOOTBOOT_INFO + 0x3c))))
		}
		kx = kx + 1
	}
}

// All go programs expect there to be a main function even though this is never called
func main() {}
