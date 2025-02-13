package data

import (
	"go-stock/backend/logger"
	"testing"
)

// TestRemoveNonPrintable tests the RemoveAllBlankChar function.
func TestRemoveNonPrintable(t *testing.T) {
	//tests := []struct {
	//	input    string
	//	expected string
	//}{
	//	{"新 希 望", "新希望"},
	//	{"", ""},
	//	{"Hello, World!", "Hello, World!"},
	//	{"\x00\x01\x02", ""},
	//	{"Hello\x00World", "HelloWorld"},
	//	{"\x1F\x20\x7E\x7F", " \x7E"},
	//}

	//for _, test := range tests {
	//	actual := RemoveAllBlankChar(test.input)
	//	if actual != test.expected {
	//		t.Errorf("RemoveAllBlankChar(%q) = %q; expected %q", test.input, actual, test.expected)
	//	}
	//}
	txt := "新 希 望"
	txt2 := RemoveAllBlankChar(txt)
	logger.SugaredLogger.Infof("RemoveAllBlankChar(%s)", txt2)
	logger.SugaredLogger.Infof("RemoveAllBlankChar(%s)", txt)

}
