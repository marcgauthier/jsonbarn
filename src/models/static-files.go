package models

import (
	"io/ioutil"
	"path"
	"strings"
	"sync"

	"time"

	"github.com/antigloss/go/logger"
)

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
