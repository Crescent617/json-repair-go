package test

import (
	"testing"

	"github.com/Crescent617/json-repair-go/jsonrepair"
)

func TestQuoteEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple quotes", `{"name": "John "Doe""}`, `{"name":"John \"Doe\""}`},
		{"Already quotes", `{"name": "John \"Doe\""}`, `{"name":"John \"Doe\""}`},
		{"Nested quotes", `{"message": "He said, "Hello, World!""}`, `{"message":"He said, \"Hello, World!\""}`},
		{"Escaped quotes", `{"quote": "This is a \"test\" message."}`, `{"quote":"This is a \"test\" message."}`},
		{"Multiple quotes", `{"text": "She replied, "Yes, I agree.""}`, `{"text":"She replied, \"Yes, I agree.\""}`},
		{"Quotes in array", `["Item 1", "Item "2"", "Item 3"]`, `["Item 1","Item \"2\"","Item 3"]`},
		{"Quotes in key", `{"user"name": "test"}`, `{"user\"name":"test"}`},
		{"Mixed quotes", `{"greeting": "Hello, "User"!"}`, `{"greeting":"Hello, \"User\"!"}`},
		{"Quotes with special chars", `{"path": "C:"Program Files"App"}`, `{"path":"C:\"Program Files\"App"}`},
		{"Quotes in nested object", `{"person": {"name": "Alice "Smith""}}`, `{"person":{"name":"Alice \"Smith\""}}`},
		{"Quotes in multiline string", `{"text": "Line 1\nLine 2 with "quotes"\nLine 3"}`, `{"text":"Line 1\nLine 2 with \"quotes\"\nLine 3"}`},
		{"Quotes with punctuation", `{"sentence": "Is this "correct"?"}`, `{"sentence":"Is this \"correct\"?"}`},
		{"Quotes in URL", `{"url": "https://example.com/"test"page"}`, `{"url":"https://example.com/\"test\"page"}`},
		{"Complex JSON with quotes", `{"data": {"items": ["Item "A"", "Item "B""], "count": 2}}`, `{"data":{"count":2,"items":["Item \"A\"","Item \"B\""]}}`},
		{"Quotes in boolean context", `{"status": "The value is "true""}`, `{"status":"The value is \"true\""}`},
		{"Quotes in numeric context", `{"value": "The number is "42""}`, `{"value":"The number is \"42\""}`},
		{"Chinese names with quotes", `{"name": "张三"李四"王五"}`, `{"name":"张三\"李四\"王五"}`},
		{"Chinese message with quotes", `{"message": "这是"测试"消息"}`, `{"message":"这是\"测试\"消息"}`},
		{"Chinese key with quotes", `{"用户"名": "test"}`, `{"用户\"名":"test"}`},
		{"Mixed Chinese and English", `{"text": "Hello"世界"World"}`, `{"text":"Hello\"世界\"World"}`},
		{"Chinese array with quotes", `["水果"苹果", "颜色"红色"]`, `["水果\"苹果","颜色\"红色"]`},
		{"Traditional Chinese", `{"traditional": "這是"測試"訊息"}`, `{"traditional":"這是\"測試\"訊息"}`},
		{"Chinese punctuation with quotes", `{"punct": "这是"测试".包含，多个"符号""}`, `{"punct":"这是\"测试\".包含，多个\"符号\""}`},
		{"Special Chinese characters", `{"special": "①②③"测试"★☆"}`, `{"special":"①②③\"测试\"★☆"}`},
		{"Chinese URL with quotes", `{"url": "https://example.com/测试"页面".html"}`, `{"url":"https://example.com/测试\"页面\".html"}`},
		{"Complex Chinese JSON", `{"用户信息": {"姓名": "张"三", "年龄": 25, "地址": "北京"朝阳区"}}`, `{"用户信息":{"地址":"北京\"朝阳区","姓名":"张\"三","年龄":25}}`},
		{"Multiline Chinese text", `{"text": "第一行"文本"，第二行"更多"文本"}`, `{"text":"第一行\"文本\"，第二行\"更多\"文本"}`},
		{"Chinese brackets with quotes", `{"brackets": "（这是"测试"）"}`, `{"brackets":"（这是\"测试\"）"}`},
		{"Chinese book title with quotes", `{"book": "《这是"测试"书》"}`, `{"book":"《这是\"测试\"书》"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonrepair.RepairJSON(tt.input)
			if err != nil {
				t.Errorf("RepairJSON() error = %v", err)
				return
			}

			// Verify result matches expected output
			if result != tt.expected {
				t.Errorf("RepairJSON() result mismatch\\nExpected: %s\\nGot:      %s", tt.expected, result)
			}

			// Verify the result is valid JSON
			if _, parseErr := jsonrepair.RepairJSONToValue(result); parseErr != nil {
				t.Errorf("Repaired JSON is not valid: %v\\nResult: %s", parseErr, result)
			}
		})
	}
}
