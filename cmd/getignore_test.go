/* The MIT License (MIT)

Copyright © 2021 Christian Korneck <christian@korneck.de>
Copyright © 2014 Ben Johnson

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE. */

package cmd_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/christian-korneck/getignore/cmd"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

const cannedTreeResponse = `
{
  "sha": "218a941be92679ce67d0484547e3e142b2f5f6f0",
  "url": "https://api.github.com/repos/github/gitignore/git/trees/218a941be92679ce67d0484547e3e142b2f5f6f0",
  "tree": [
    {
		"path": "Go.gitignore",
		"mode": "100644",
		"type": "blob",
		"sha": "66fd13c903cac02eb9657cd53fb227823484401d",
		"size": 269,
		"url": "https://api.github.com/repos/github/gitignore/git/blobs/66fd13c903cac02eb9657cd53fb227823484401d"
	  },
    {
      "path": "community/Golang",
      "mode": "040000",
      "type": "tree",
      "sha": "9359186308beec2e179c9932a7becb6bf5a54b81",
      "url": "https://api.github.com/repos/github/gitignore/git/trees/9359186308beec2e179c9932a7becb6bf5a54b81"
    },
    {
      "path": "community/Golang/Hugo.gitignore",
      "mode": "100644",
      "type": "blob",
      "sha": "37fa330e4fc32f1cd62b825c27ae30fd6d3a0ba2",
      "size": 125,
      "url": "https://api.github.com/repos/github/gitignore/git/blobs/37fa330e4fc32f1cd62b825c27ae30fd6d3a0ba2"
    }
  ],
  "truncated": false
}
`

const cannedGitignore = `
something-something/*
*.exe

dist/
`

const expectedGitignore = `

# --- start Go --- 


something-something/*
*.exe

dist/


# --- end Go ---`

func TestRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.String() == "/repos/github/gitignore/git/trees/main?recursive=1" {
			rw.Write([]byte(cannedTreeResponse))
		}
		if req.URL.String() == "/github/gitignore/main/Go.gitignore" {
			rw.Write([]byte(cannedGitignore))
		}
	}))

	defer server.Close()

	rc := cmd.RestClient{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	body, err := rc.Run([]string{"go"})
	ok(t, err)
	equals(t, expectedGitignore, body)

}
