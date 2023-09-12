package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Context_addString(t *testing.T) {
	ass := assert.New(t)
	logContext := ContextForMethod(Test_Context_addString).
		AddString("un TU", "c'est trop cool").
		AddAny("deux TU", "c'est encore mieux")
	ass.Len(*logContext, 3)
	ass.Equal((*logContext)[0].Key, "method")
	ass.Equal((*logContext)[1].Key, "un TU")
	ass.Equal((*logContext)[1].Value.String(), "c'est trop cool")
	ass.Equal((*logContext)[2].Key, "deux TU")
	ass.Equal((*logContext)[2].Value.String(), "c'est encore mieux")
}
