package models

/*Package models - static-files.go

This file contain miscillaneous functions helper.
______________________________________________________________________________

 Ecureuil - Web framework for real-time javascript app.
_____________________________________________________________________________

MIT License

Copyright (c) 2014-2017 Marc Gauthier

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

______________________________________________________________________________


Revision:
	01 Nov 2016 - Clean code, audit.

______________________________________________________________________________

*/

import (
	"io/ioutil"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/antigloss/go/logger"
)

var cacheIsEnabled = false

type cachefile struct {
	validuntil uint64
	buf        []byte
}

var lock = sync.RWMutex{}

var filecache map[string]cachefile

/*InitFileCache create the map for storing the files into a basic cache system.
 */
func InitFileCache() {
	filecache = make(map[string]cachefile)
}

/*GetStaticFile the webserver can serve files to HTTPS request.  The webserver will only serve files that are
  present in the public folder.
  The file extension is also return so that a proper content-descriptor can be set.

  In there future we can add file caching and filtering here
*/
func GetStaticFile(filepath string) ([]byte, string, error) {

	ext := strings.ToLower(path.Ext(filepath))

	if cacheIsEnabled {

		lock.RLock()
		c, ok := filecache[filepath]
		lock.RUnlock()

		//do something here

		if ok {

			// key exists

			if c.validuntil >= uint64(time.Now().UTC().Unix()) {

				logger.Trace("Serving cache file thru HTTPS " + filepath)
				return c.buf, ext, nil
			}

			// cache exists but has expired
			lock.Lock()
			delete(filecache, filepath)
			lock.Unlock()

		}
	}

	var cache cachefile
	var err error

	cache.validuntil = uint64(time.Now().Add(time.Duration(5 * time.Minute)).UTC().Unix())
	cache.buf, err = ioutil.ReadFile("public/" + filepath)

	if err != nil {
		logger.Error(err.Error())
	}

	// add to cache
	lock.Lock()
	filecache[filepath] = cache
	lock.Unlock()

	logger.Trace("Serving file thru HTTPS " + filepath)
	return cache.buf, ext, err
}
