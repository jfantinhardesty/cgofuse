/*
 * host_test.go
 *
 * Copyright 2017-2022 Bill Zissimopoulos
 */
/*
 * This file is part of Cgofuse.
 *
 * It is licensed under the MIT license. The full license text can be found
 * in the License.txt file at the root of this project.
 */

package fuse

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

type testfs struct {
	FileSystemBase
	init, dstr int
}

func (fs *testfs) Init() {
	fs.init++
}

func (fs *testfs) Destroy() {
	fs.dstr++
}

func (fs *testfs) Getattr(path string, stat *Stat_t, fh uint64) (errc int) {
	switch path {
	case "/":
		stat.Mode = S_IFDIR | 0555
		return 0
	default:
		return -ENOENT
	}
}

func (fs *testfs) Readdir(path string,
	fill func(name string, stat *Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64, flags uint32) (errc int) {
	fill(".", nil, 0)
	fill("..", nil, 0)
	return 0
}

func testHost(t *testing.T, unmount bool) {
	path, err := os.MkdirTemp("", "test")
	if nil != err {
		panic(err)
	}
	defer os.Remove(path)
	mntp := filepath.Join(path, "m")
	if runtime.GOOS != "windows" {
		err = os.Mkdir(mntp, os.FileMode(0755))
		if nil != err {
			panic(err)
		}
		defer os.Remove(mntp)
	}
	done := make(chan bool)
	tmch := time.After(3 * time.Second)
	tstf := &testfs{}
	host := NewFileSystemHost(tstf)
	mres := false
	ures := false
	go func() {
		mres = host.Mount(mntp, nil)
		done <- true
	}()
	<-tmch
	if unmount {
		ures = host.Unmount()
	} else {
		ures = sendInterrupt()
	}
	<-done
	if !mres {
		t.Error("Mount failed")
	}
	if !ures {
		t.Error("Unmount failed")
	}
	if tstf.init != 1 {
		t.Errorf("Init() called %v times; expected 1", tstf.init)
	}
	if tstf.dstr != 1 {
		t.Errorf("Destroy() called %v times; expected 1", tstf.dstr)
	}
}

func TestUnmount(t *testing.T) {
	testHost(t, true)
}

func TestSignal(t *testing.T) {
	if runtime.GOOS != "windows" {
		testHost(t, false)
	}
}
