package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Context_addSomeFields(t *testing.T) {
	ass := assert.New(t)
	logContext := ContextForMethod(Test_Context_addSomeFields).
		AddString("un TU", "c'est trop cool").
		AddAny("deux TU", "c'est encore mieux")
	ass.Len(*logContext, 3)
	ass.Equal((*logContext)["method"].Key, "method")
	ass.Equal((*logContext)["un TU"].Key, "un TU")
	ass.Equal((*logContext)["un TU"].Value.String(), "c'est trop cool")
	ass.Equal((*logContext)["deux TU"].Key, "deux TU")
	ass.Equal((*logContext)["deux TU"].Value.String(), "c'est encore mieux")
}

func Test_Context_removeField(t *testing.T) {
	ass := assert.New(t)
	logContext := ContextForMethod(Test_Context_addSomeFields).
		AddAny("to remove", fake.Lorem().Word()).
		Remove("to remove")
	ass.Len(*logContext, 1)
	ass.Equal((*logContext)["method"].Key, "method")
}

func Test_Context_addSameFieldTwiceGetOnlyLastValue(t *testing.T) {
	ass := assert.New(t)
	logContext := ContextForMethod(Test_Context_addSomeFields).
		AddAny("un test", fake.Lorem().Word()).
		AddString("un test", "c'est bien")
	ass.Len(*logContext, 2)
	ass.Equal((*logContext)["un test"].Key, "un test")
	ass.Equal((*logContext)["un test"].Value.String(), "c'est bien")
}

func Test_Context_clone(t *testing.T) {
	ass := assert.New(t)
	originalContext := ContextForMethod(Test_Context_addSomeFields).
		AddString("original", "c'est bien")

	clonedContext := originalContext.Clone().AddString("clone", "c'est pas bien de copier")
	ass.NotEqual(*originalContext, *clonedContext)
	ass.NotSame(originalContext, clonedContext)

	ass.Len(*originalContext, 2)
	ass.Equal((*originalContext)["original"].Key, "original")
	ass.Equal((*originalContext)["original"].Value.String(), "c'est bien")

	ass.Len(*clonedContext, 3)
	ass.Equal((*clonedContext)["original"].Key, "original")
	ass.Equal((*clonedContext)["original"].Value.String(), "c'est bien")
	ass.Equal((*clonedContext)["clone"].Key, "clone")
	ass.Equal((*clonedContext)["clone"].Value.String(), "c'est pas bien de copier")

}
