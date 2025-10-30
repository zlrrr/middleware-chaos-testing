package core

import (
	"testing"
)

// TestDummy 占位符测试，验证测试框架正常工作
func TestDummy(t *testing.T) {
	// 简单的断言测试
	if 1+1 != 2 {
		t.Error("Math is broken!")
	}
}

// TestExample 示例表驱动测试
func TestExample(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"positive numbers", 1, 2, 3},
		{"negative numbers", -1, -2, -3},
		{"zero", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a + tt.b
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
