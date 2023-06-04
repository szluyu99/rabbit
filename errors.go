package rabbit

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

type Error struct {
	Code int
	Msg  string
}

func (e Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

var (
	ErrPermissionDenied = &Error{Code: http.StatusForbidden, Msg: "permission denied"}
)

func HandleTheError(c *gin.Context, e *Error) {
	c.AbortWithStatusJSON(e.Code, gin.H{"error": e.Msg})
}

func HandleError(c *gin.Context, code int, err error) {
	if e, ok := err.(*Error); ok {
		code = e.Code
	}

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

func HandleErrorMsg(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(code, gin.H{"error": msg})
}
