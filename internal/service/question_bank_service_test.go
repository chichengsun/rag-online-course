package service

import (
	"strings"
	"testing"
)

func TestMergeChoiceStemIntoStem(t *testing.T) {
	t.Parallel()
	stem := mergeChoiceStemIntoStem("下列说法正确的是？", []string{"选项一", "选项二"}, "single_choice")
	for _, w := range []string{"下列说法正确的是？", "A. 选项一", "B. 选项二"} {
		if !strings.Contains(stem, w) {
			t.Fatalf("stem %q 应包含 %q", stem, w)
		}
	}
	if mergeChoiceStemIntoStem("仅简答", []string{"x"}, "short_answer") != "仅简答" {
		t.Fatal("非选择题不应拼接 options")
	}
	if mergeChoiceStemIntoStem("无选项题", nil, "single_choice") != "无选项题" {
		t.Fatal("空 options 应保持原 stem")
	}
}
