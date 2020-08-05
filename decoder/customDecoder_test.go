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

package decoder

import "testing"

func TestInitCustomDecoders(t *testing.T) {
	c = &Config{
		Out:                     "",
		Source:                  "",
		CustomRegex:             "",
		MemProfile:              "",
		IncludeDecoders:         "",
		ExcludeDecoders:         "",
		FileStorage:             "",
		ConnFlushInterval:       0,
		MemBufferSize:           0,
		FlowTimeOut:             0,
		StreamDecoderBufSize:    0,
		CloseInactiveTimeOut:    0,
		FlushEvery:              0,
		HarvesterBannerSize:     0,
		BannerSize:              0,
		ClosePendingTimeOut:     0,
		FlowFlushInterval:       0,
		ConnTimeOut:             0,
		UseRE2:                  false,
		StopAfterHarvesterMatch: false,
		Buffer:                  false,
		WriteIncomplete:         false,
		WriteChan:               false,
		CSV:                     false,
		AddContext:              false,
		WaitForConnections:      false,
		HexDump:                 false,
		Debug:                   false,
		AllowMissingInit:        false,
		IgnoreFSMerr:            false,
		CalculateEntropy:        false,
		SaveConns:               false,
		TCPDebug:                false,
		NoOptCheck:              false,
		Checksum:                false,
		DefragIPv4:              false,
		Export:                  false,
		IncludePayloads:         false,
		Compression:             false,
		IgnoreDecoderInitErrors: false,
	}
	decoders, err := InitCustomDecoders(c)
	if err != nil {
		t.Fatal(err)
	}

	if len(decoders) == 0 {
		t.Fatal("no custom decoders after initialization")
	}
}

func TestCustomDecoder_Decode(t *testing.T) {
}

func TestCustomDecoder_Destroy(t *testing.T) {
}
