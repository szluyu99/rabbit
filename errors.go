package rabbit

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, code int, err error) {
	msg := err.Error()

	// add file and line number
	if log.Flags()&(log.Lshortfile|log.Llongfile) != 0 {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "???"
		}
		pos := strings.LastIndex(file, "/")
		if log.Flags()&log.Lshortfile != 0 && pos >= 0 {
			file = file[pos+1:]
		}
		err = fmt.Errorf("%s:%d: %v", file, line, err)
	}

	c.Error(err)
	c.AbortWithStatusJSON(code, gin.H{"error": msg})
}

func HandleErrorMessage(c *gin.Context, code int, msg string) {
	c.Error(fmt.Errorf(msg))
	c.AbortWithStatusJSON(code, gin.H{"error": msg})
}
