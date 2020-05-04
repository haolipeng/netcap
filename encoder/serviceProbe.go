package encoder

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dlclark/regexp2"
	"github.com/dreadl0ck/netcap/utils"
	"github.com/mgutz/ansi"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	serviceProbes []*ServiceProbe
	ignoredProbes = map[string]struct{}{
		"pc-duo-gw": {},
		"ventrilo":  {},
		"pc-duo":    {},
		"ssl":       {},
	}
	useRE2 = true
)

type ServiceProbe struct {
	RegEx           *regexp.Regexp
	RegEx2          *regexp2.Regexp
	RegExRaw        string
	Vendor          string
	Version         string
	Info            string
	Hostname        string
	OS              string
	DeviceType      string
	CPEs            map[string]string
	CaseInsensitive bool
	IncludeNewlines bool
	Ident           string
}

func (s *ServiceProbe) String() string {

	var b strings.Builder

	b.WriteString("ServiceProbe: ")
	b.WriteString(s.Ident)
	b.WriteString("\nRegEx: ")
	b.WriteString(s.RegExRaw)
	if len(s.Vendor) > 0 {
		b.WriteString("\nVendor: ")
		b.WriteString(s.Vendor)
	}
	if len(s.Version) > 0 {
		b.WriteString("\nVersion: ")
		b.WriteString(s.Version)
	}
	if len(s.Info) > 0 {
		b.WriteString("\nInfo: ")
		b.WriteString(s.Info)
	}
	if len(s.Hostname) > 0 {
		b.WriteString("\nHostname: ")
		b.WriteString(s.Hostname)
	}
	if len(s.OS) > 0 {
		b.WriteString("\nOS: ")
		b.WriteString(s.OS)
	}
	if len(s.DeviceType) > 0 {
		b.WriteString("\nDeviceType: ")
		b.WriteString(s.DeviceType)
	}
	//b.WriteString("\nCPEs: ")
	//b.WriteString(s.CPEs)
	b.WriteString("\nCaseInsensitive: ")
	b.WriteString(strconv.FormatBool(s.CaseInsensitive))
	b.WriteString("\nIncludeNewlines: ")
	b.WriteString(strconv.FormatBool(s.IncludeNewlines))

	return b.String()
}

// only parse the match directive for now.
// match <proto> m|<regex>|<i>/<s> <meta>
// allow using $1 or $2 perl style substrings in meta section
// helpers:
// - filter unprintable chars
// - substitute strings
// - unpack unsigned int

// example data:
//match amanda m|^220 ([-.\w]+) AMANDA index server \((\d[-.\w ]+)\) ready\.\r\n| p/Amanda backup system index server/ v/$2/ o/Unix/ h/$1/ cpe:/a:amanda:amanda:$2/
//match amanda m|^501 Could not read config file [^!\r\n]+!\r\n220 ([-.\w]+) AMANDA index server \(([-\w_.]+)\) ready\.\r\n| p/Amanda backup system index server/ v/$2/ i/broken: config file not found/ h/$1/ cpe:/a:amanda:amanda:$2/
//match amanda m|^ld\.so\.1: amandad: fatal: (libsunmath\.so\.1): open failed: No such file or directory\n$| p/Amanda backup system index server/ i/broken: $1 not found/ cpe:/a:amanda:amanda/
//match amanda m|^\n\*\* \(process:\d+\): CRITICAL \*\*: GLib version too old \(micro mismatch\): Amanda was compiled with glib-[\d.]+, but linking with ([\d.]+)\n| p/Amanda backup system index server/ i/broken: GLib $1 too old/ cpe:/a:amanda:amanda/

// parseVersionInfo uses the next read byte as a delimiter
// and reads everything into a buffer until the delimiter appears again
// it returns the final buffer and an error and advances the passed in *bytes.Reader to the
func parseVersionInfo(r *bytes.Reader) (string, error) {

	var res []byte
	d, err := r.ReadByte()
	if err != nil {
		return "", err
	}

	for {
		bb, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		if bb == d {
			break
		}
		res = append(res, bb)
	}

	//fmt.Println("parsed meta", string(res))
	return string(res), nil
}

func InitProbes() error {
	// load nmap service probes
	data, err := ioutil.ReadFile("/usr/local/etc/netcap/dbs/nmap-service-probes")
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	serviceProbes = make([]*ServiceProbe, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 || line == "\n" || strings.HasPrefix(line, "#") {
			// ignore comments and blanks
			continue
		}
		if strings.HasPrefix(line, "match") {

			ident := strings.Fields(line)[1]
			if _, ok := ignoredProbes[ident]; ok {
				utils.DebugLog.Println("ignoring probe", ident)
				continue
			}

			var (
				spaces    int
				delim     byte
				regex     []byte
				r         = bytes.NewReader([]byte(line))
				checkOpts bool
				parseMeta bool
				s         = new(ServiceProbe)
			)
			s.Ident = ident

			for {
				b, err := r.ReadByte()
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}
				//fmt.Println("read", string(b))

				if unicode.IsSpace(rune(b)) && !checkOpts {
					//fmt.Println("its a space", string(b))
					spaces++

					if delim != 0 {
						// collect whitespace when parsing the regex string
						regex = append(regex, b)
					}
					continue
				}
				// last part versionInfo: m/[regex]/[opts] [meta]
				// example: p/Amanda backup system index server/ i/broken: GLib $1 too old/ cpe:/a:amanda:amanda/
				if parseMeta {

					// skip over whitespace
					if unicode.IsSpace(rune(b)) {
						//fmt.Println("parse meta: skip whitespace")
						continue
					}

					// parse a version info block
					var errParse error
					switch string(b) {
					case "p":
						s.Vendor, errParse = parseVersionInfo(r)
						if errParse != nil {
							return errParse
						}
					case "v":
						s.Version, errParse = parseVersionInfo(r)
						if errParse != nil {
							return errParse
						}
					case "i":
						s.Info, errParse = parseVersionInfo(r)
						if errParse != nil {
							return errParse
						}
					case "h":
						s.Hostname, errParse = parseVersionInfo(r)
						if errParse != nil {
							return errParse
						}
					case "o":
						s.OS, errParse = parseVersionInfo(r)
						if errParse != nil {
							return errParse
						}
					case "d":
						s.DeviceType, errParse = parseVersionInfo(r)
						if errParse != nil {
							return errParse
						}
					// TODO: handle cpe tags
					case "c":
						// ignore for now and stop parsing
						//fmt.Println("got a c, likely cpe tag. ignoring for now.")
						goto next
					}

					continue
				}
				// m/[regex]/[opts]
				// - there can be an optional i for case insensitive matching
				// - or an 's' to include newlines in the '.' specifier
				if checkOpts {
					if unicode.IsSpace(rune(b)) {

						//fmt.Println("options done!")
						// options done
						checkOpts = false
						parseMeta = true
						continue
					}
					switch string(b) {
					case "i":
						s.CaseInsensitive = true
					case "s":
						s.IncludeNewlines = true
					}
					continue
				}
				// check if delimiter was already found
				if delim != 0 {
					if b == delim {
						//fmt.Println("parsed regex", ansi.Blue, string(regex), ansi.Reset, "from line", ansi.Green, line, ansi.Reset)

						// parse options
						checkOpts = true

						continue
					}
					regex = append(regex, b)
					continue
				}
				// start of regex
				if spaces == 2 {
					if string(b) != "m" {
						return errors.New("invalid format for line: " + line)
					}

					// read delimiter
					b, err = r.ReadByte()
					if err == io.EOF {
						break
					} else if err != nil {
						return err
					}

					//fmt.Println("read delim", string(b))

					delim = b
					continue
				}
			}
		next:
			// compile regex
			var (
				errCompile error
				finalReg   = "(?m"
			)

			// To change the default matching behavior, you can add a set of flags to the beginning of a regular expression.
			// For example, the prefix "(?is)" makes the matching case-insensitive and lets . match \n. (The default matching is case-sensitive and . doesn’t match \n.)
			if s.CaseInsensitive {
				finalReg += "i"
			}
			if s.IncludeNewlines {
				finalReg += "s"
			}
			finalReg += ")" + strings.TrimSpace(string(regex))

			before := finalReg

			if useRE2 {
				finalReg = clean(finalReg)
				s.RegEx, errCompile = regexp.Compile(finalReg)
			} else {
				s.RegEx2, errCompile = regexp2.Compile(finalReg, 0) // regexp.RE2)
			}

			if errCompile != nil {
				if c.Debug {
					if useRE2 {
						fmt.Println("before:", before)
						fmt.Println("failed to compile regex:", ansi.Red, errCompile, ansi.White, finalReg, ansi.Reset) // stdlib regexp only logs the broken part of the regex. this logs the full regex string for debugging
					} else {
						fmt.Println("failed to compile regex:", ansi.Red, errCompile, ansi.Reset)
						fmt.Println(ansi.White, line, ansi.Reset)
					}
				}
			} else {
				s.RegExRaw = finalReg
				serviceProbes = append(serviceProbes, s)
			}
		}
	}

	utils.DebugLog.Println("loaded", len(serviceProbes), "nmap service probes")

	return nil
}

// DumpServiceProbes prints all loaded probes as JSON
func DumpServiceProbes() {
	for _, p := range serviceProbes {
		data, err := json.MarshalIndent(p, " ", "  ")
		if err == nil {
			fmt.Println(string(data))
		}
	}
}

func clean(in string) string {
	var (
		r                   = bytes.NewReader([]byte(in))
		out                 []byte
		check               bool
		ignore              bool
		firstQuestionMark   = true
		stopCnt, startCount = -1, -1
		resetEscaped, escaped bool
		lastchar byte
	)
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				break
			}
		}
		debug := func(args ...interface{}) {
			// TODO: make debug mode configurable
			//fmt.Println(string(lastchar), ansi.Blue, string(b), ansi.Red, startCount, stopCnt, ansi.Green, string(out), ansi.White, args, ansi.Reset, in)
		}
		if string(b) == "\\" && !escaped {
			debug("set escaped to true")
			escaped = true
		} else {
			if escaped {
				if resetEscaped {
					debug("reset escaped")
					// reset
					escaped = false
					resetEscaped = false
				} else {
					// reset escaped next round
					resetEscaped = true
				}
			}
		}
		if ignore {
			if string(b) == ")" {

				if !escaped {

					stopCnt++
					debug("stopCnt++")

					if startCount == stopCnt {

						debug("stop ignore")

						ignore = false
						out = append(out, byte(')'))
						check = false

						stopCnt = 0
						startCount = 0

						lastchar = b
						continue
					}
				} else {
					debug("ignoring because escaped")
				}
			}
			if string(b) == "(" {
				if !escaped {
					startCount++
					debug("startCount++")
				}
			}
			debug("ignore")

			lastchar = b
			continue
		}
		if string(b) == "(" {

			debug("got parentheses")
			if !escaped {
				startCount++
				debug("startCount++")
			}

			out = append(out, b)
			check = true

			lastchar = b
			continue
		}
		if check {
			if string(b) == "?" && lastchar == '(' {

				debug("got ?")
				if firstQuestionMark {
					firstQuestionMark = false
				} else {
					debug("write .*")
					out = append(out, byte('.'))
					out = append(out, byte('*'))
					ignore = true

					lastchar = b
					continue
				}
			}
		}
		if string(b) == ")" {
			if !escaped {
				stopCnt++
				debug("stopCnt++")
			}
		}
		if string(b) == "(" {
			if !escaped {
				startCount++
				debug("startCount++")
			}
		}
		debug("collect")
		out = append(out, b)
		lastchar = b
	}
	return string(out)
}