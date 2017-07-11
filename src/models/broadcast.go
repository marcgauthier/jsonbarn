/*
______________________________________________________________________________

 Ecureuil - Web framework for real-time javascript app.
_____________________________________________________________________________

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


This file contain functions to create a Queue of messages that need to be
broadcasted to all the users.  This file only contain the Queue list it
does not care what method is used to send messages to users.

______________________________________________________________________________


Revision:


	01 Nov 2016 - Clean code, audit.

______________________________________________________________________________

*/

package models

import "sync"

/* declare a type to hold all the messages with sync capabillity. */

type tMessageQueue struct {
	sync.RWMutex
	queue [][]byte
}

/* declare a container to hold all the messages with sync capabillity. */

var messages tMessageQueue

/* function to extract the next message from the queue return nil if queue is empty */

/*BroadcastGet extract queue message
 */
func BroadcastGet() []byte {

	// declare return object.
	var item []byte

	/* lock the message object so we can safely delete the queue */
	messages.Lock()
	defer messages.Unlock()

	/* if there is at least one message return it */
	if len(messages.queue) > 0 {
		item = messages.queue[0]
		// remove item 0 from the queue.
		messages.queue = append(messages.queue[:0], messages.queue[1:]...) // or a = a[:i+copy(a[i:], a[i+1:])]
	}

	/* return next Queue item or nil if queue is empty */
	return item
}

/*BroadcastPut add a message to the broadcaster, in the form bucket:message
 */
func BroadcastPut(bucket, message string) error {

	// lock queue before we can insert data.
	messages.Lock()
	defer messages.Unlock()

	// insert data
	messages.queue = append(messages.queue, []byte(bucket+":"+message))

	// return no error
	return nil

	// defered unlock is executed here.
}
