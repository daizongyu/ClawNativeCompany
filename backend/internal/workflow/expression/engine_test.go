package expression

import (
	"testing"
)

func TestEngine_Evaluate(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
		want       bool
		wantErr    bool
	}{
		// 基本比较
		{
			name:       "相等比较",
			expression: `status == "active"`,
			context:    map[string]interface{}{"status": "active"},
			want:       true,
		},
		{
			name:       "不等比较",
			expression: `status != "inactive"`,
			context:    map[string]interface{}{"status": "active"},
			want:       true,
		},
		// 数字比较
		{
			name:       "大于比较",
			expression: `amount > 100`,
			context:    map[string]interface{}{"amount": 150},
			want:       true,
		},
		{
			name:       "小于等于比较",
			expression: `count <= 10`,
			context:    map[string]interface{}{"count": 10},
			want:       true,
		},
		// 逻辑与
		{
			name:       "逻辑与",
			expression: `status == "active" && level > 5`,
			context:    map[string]interface{}{"status": "active", "level": 10},
			want:       true,
		},
		{
			name:       "逻辑与失败",
			expression: `status == "active" && level > 5`,
			context:    map[string]interface{}{"status": "active", "level": 3},
			want:       false,
		},
		// 逻辑或
		{
			name:       "逻辑或",
			expression: `role == "admin" || role == "manager"`,
			context:    map[string]interface{}{"role": "manager"},
			want:       true,
		},
		// 非运算
		{
			name:       "非运算",
			expression: `!isDeleted`,
			context:    map[string]interface{}{"isDeleted": false},
			want:       true,
		},
		{
			name:       "非运算2",
			expression: `!isDeleted`,
			context:    map[string]interface{}{"isDeleted": true},
			want:       false,
		},
		// 括号
		{
			name:       "括号优先级",
			expression: `(a > 0 && b < 20) || c > 100`,
			context:    map[string]interface{}{"a": 5, "b": 15, "c": 50},
			want:       true,
		},
		{
			name:       "括号改变优先级",
			expression: `a > 0 && (b < 20 || c > 100)`,
			context:    map[string]interface{}{"a": 5, "b": 25, "c": 150},
			want:       true,
		},
		// contains
		{
			name:       "字符串包含",
			expression: `tags contains "urgent"`,
			context:    map[string]interface{}{"tags": "high urgent bug"},
			want:       true,
		},
		{
			name:       "字符串不包含",
			expression: `tags contains "critical"`,
			context:    map[string]interface{}{"tags": "high urgent bug"},
			want:       false,
		},
		// 空表达式
		{
			name:       "空表达式",
			expression: "",
			context:    map[string]interface{}{},
			want:       true,
		},
		// 布尔值
		{
			name:       "布尔值 true",
			expression: `isValid`,
			context:    map[string]interface{}{"isValid": true},
			want:       true,
		},
		{
			name:       "布尔值 false",
			expression: `isValid`,
			context:    map[string]interface{}{"isValid": false},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Evaluate(tt.expression, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEngine_Evaluate_Errors(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name       string
		expression string
		context    map[string]interface{}
	}{
		{
			name:       "未定义变量",
			expression: `undefinedVar == "test"`,
			context:    map[string]interface{}{},
		},
		{
			name:       "未闭合括号",
			expression: `(a > 0 && b < 20`,
			context:    map[string]interface{}{"a": 5, "b": 15},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.Evaluate(tt.expression, tt.context)
			if err == nil {
				t.Errorf("Evaluate() expected error but got none")
			}
		})
	}
}

func BenchmarkEngine_Evaluate(b *testing.B) {
	engine := NewEngine()
	expression := `status == "active" && (level > 5 || role == "admin")`
	context := map[string]interface{}{
		"status": "active",
		"level":  10,
		"role":   "user",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Evaluate(expression, context)
	}
}
