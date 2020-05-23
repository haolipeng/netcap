/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017-2020 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package encoder

import (
	"strings"
	"testing"
	"time"
)

var softwareTests = []regexTest{
	{
		name:     "Windows Netcat",
		input:    "Test123\nMicrosoft Windows [Version 10.0.10586]\n(c) 2015 Microsoft Corporation. All rights reserved. \nC:\\cygwin\\netcat>",
		expected: "Microsoft Windows [Version 10.0.10586]",
	},
	{
		name:     "Apache Test",
		input:    "Hello dears,\nfor our hosting we will use Apache 2.4.29\nThere are other options,howver,\nlike Lighttp 2.3.4",
		expected: "for our hosting we will use Apache 2.4.29-like Lighttp 2.3.4",
	},
	{
		name:     "NginX Test",
		input:    "We will test\ncan we detect NginX v2.3.4\nI hope so\nwe'll see",
		expected: "can we detect NginX v2.3.4",
	},
	{
		name:     "NginX Test",
		input:    "We will test\ncan we detect NginX version 2.3.4\nI hope so\nwe'll see",
		expected: "can we detect NginX version 2.3.4",
	},
}

func (r regexTest) testSoftwareHarvester(t *testing.T) {
	s := softwareHarvester([]byte(r.input), "", time.Now(), "test", "test", []string{})

	parts := strings.Split(r.expected, "-")

	if len(s) != len(parts) {
		t.Fatal("Expected:", len(parts), " found: ", len(s), " results")
	}

	for i, c := range parts {
		if c != s[i].Notes {
			t.Fatal("Expected: ", c, " Received: ", s[i].Notes)
		}
	}
}

func TestGenericVersionHarvester(t *testing.T) {
	for _, r := range softwareTests {
		r.testSoftwareHarvester(t)
	}
}