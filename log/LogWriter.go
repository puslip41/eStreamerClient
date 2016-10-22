package log

import (
	"os"
	"bufio"
	"fmt"
)

type LogWriter struct {
	File *os.File
	Writer *bufio.Writer
}

func (writer *LogWriter) WriteFormat(format string, args ...interface{}) {
	writer.Writer.WriteString(fmt.Sprintf(format, args...))
}

func (writer *LogWriter) Flush() {
	writer.Writer.Flush()
}

func (writer *LogWriter) Close () {
	writer.Writer.Flush()
	writer.File.Close()
}