package logger

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

type SimpleFormatter struct{}

func (s *SimpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}
	concat(b, printTime(entry))
	concat(b, " ")
	concat(b, printLevel(entry))
	//concat(b, " ")
	//concat(b, printCaller(entry))
	concat(b, " ")
	concat(b, entry.Message)
	data := printData(entry)
	if len(data) > 0 {
		concat(b, "\t- ")
		concat(b, data)
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func concat(b *bytes.Buffer, input string) {
	if _, err := fmt.Fprintf(b, input); err != nil {
		panic(err)
	}
}

func printData(entry *logrus.Entry) string {
	data := entry.Data
	if len(data) <= 0 {
		return ""
	}
	// sort data
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var s string
	for _, k := range keys {
		s += k + "=\"" + fmt.Sprintf("%v", data[k]) + "\" "
	}
	return s
}

//func printCaller(entry *logrus.Entry) string {
//	return entry.Caller.File + ":" + strconv.Itoa(entry.Caller.Line)
//}

func printLevel(entry *logrus.Entry) string {
	return strings.ToUpper("[" + entry.Level.String() + "]")
}

func printTime(entry *logrus.Entry) string {
	return entry.Time.Format("2006-01-02 15:04:05")
}
