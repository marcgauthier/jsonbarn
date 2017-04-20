/*Package models - misc.go

This file contain miscillaneous functions helper.

______________________________________________________________________________

 OWLSO - Overwatch Link and Service Observer.
______________________________________________________________________________

MIT License

Copyright (c) 2014-2016 Marc Gauthier

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
package models

import (
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

/*EscDoubleQuote escape double quote inside a string.
 */
func EscDoubleQuote(s string) string {
	return strings.Replace(s, "\"", "\\\"", -1)
}

/*PrepMessageForUser prepare a message to be return to the FRONT-END, i.e. for alerting the user about suceess or
failure of the storate request.  The return byte array must contain a JSON object structure
*/
func PrepMessageForUser(msg string) []byte {
	return []byte("{ \"action\":\"message\", \"message\":\"" + EscDoubleQuote(msg) + "\"}")

}

/*UnixUTCSecs return the number of seconds since 1970
 */
func UnixUTCSecs() float64 {
	return float64(time.Now().Unix())
}

/*UnixUTCNano return the current UTC time in Unix (Nanoeconds Elaspse since)
 */
func UnixUTCNano() float64 {
	return float64(time.Now().UnixNano())
}

/*FileExists check if a specific file exists on the server harddrive.
This function os use by owlso.go to determine of the
SSL security certificate exists of if they need to be created.
*/
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

/*
bluemonday.UGCPolicy() which allows a broad selection of HTML elements and attributes that are safe for user generated content.
Note that this policy does not whitelist iframes, object, embed, styles, script, etc.
An example usage scenario would be blog post bodies where a variety of formatting is expected along with the potential
for TABLEs and IMGs.
*/

/*SanitizeHTML sanitize Data contain in a Byte Array.
 */
func SanitizeHTML(msg []byte) []byte {

	if len(msg) > 0 {
		p := bluemonday.UGCPolicy()
		return p.SanitizeBytes(msg)
	}

	return msg
}

/*SanitizeStrHTML sanitize Data contain in a String
 */
func SanitizeStrHTML(msg string) string {

	if msg != "" {
		p := bluemonday.UGCPolicy()
		return p.Sanitize(msg)
	}

	return msg

}

/*SanitizeJSONStrHTML sanitize HTML within an JSON but
do not convert single and double quote
*/
func SanitizeJSONStrHTML(json string) string {

	return strings.Replace(strings.Replace(SanitizeStrHTML(json), `&#34;`, `"`, -1), `&#39;`, `'`, -1)

}

func removeIndex(s []uint64, index int) []uint64 {
	return append(s[:index], s[index+1:]...)
}

/*IsStrInArray search a []string to confirm if a string is present
 */
func IsStrInArray(r string, rights []string) bool {
	for i := 0; i < len(rights); i++ {
		if strings.ToLower(rights[i]) == strings.ToLower(r) {
			return true
		}
	}
	return false
}

/*RandomPassword generate a password of x characters randomly
 */
func RandomPassword(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
