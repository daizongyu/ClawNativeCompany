// Package expression 提供条件表达式解析和求值
package expression

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Engine 表达式引擎
type Engine struct{}

// NewEngine 创建新的表达式引擎
func NewEngine() *Engine {
	return &Engine{}
}

// Evaluate 求值表达式
// 支持的操作符：==, !=, >, <, >=, <=, contains, &&, ||, !, ()
func (e *Engine) Evaluate(expression string, context map[string]interface{}) (bool, error) {
	if strings.TrimSpace(expression) == "" {
		return true, nil // 空表达式默认通过
	}

	// 解析表达式为 AST
	ast, err := e.parse(expression)
	if err != nil {
		return false, fmt.Errorf("表达式解析失败: %w", err)
	}

	// 求值 AST
	result, err := e.eval(ast, context)
	if err != nil {
		return false, fmt.Errorf("表达式求值失败: %w", err)
	}

	return result, nil
}

// AST 节点类型
type nodeType int

const (
	nodeLiteral nodeType = 1  // 字面量
	nodeVariable nodeType = 2 // 变量
	nodeBinary  nodeType = 3  // 二元操作
	nodeUnary   nodeType = 4  // 一元操作
)

// AST 节点
type node struct {
	typ       nodeType
	value     interface{} // 字面量值或变量名
	operator  string      // 操作符
	left      *node       // 左子树
	right     *node       // 右子树
}

// 操作符优先级
var precedence = map[string]int{
	"||": 1,
	"&&": 2,
	"==": 3,
	"!=": 3,
	">":  4,
	"<":  4,
	">=": 4,
	"<=": 4,
	"contains": 4,
	"!":  5,
}

// 解析表达式为 AST
func (e *Engine) parse(expression string) (*node, error) {
	tokens, err := e.tokenize(expression)
	if err != nil {
		return nil, err
	}

	node, _, err := e.parseExpression(tokens, 0)
	return node, err
}

// tokenize 词法分析
func (e *Engine) tokenize(expression string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inString := false
	stringChar := rune(0)

	for _, ch := range expression {
		if inString {
			current.WriteRune(ch)
			if ch == stringChar {
				inString = false
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		if ch == '"' || ch == '\'' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			inString = true
			stringChar = ch
			current.WriteRune(ch)
			continue
		}

		// 跳过空白字符
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		// 括号
		if ch == '(' || ch == ')' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(ch))
			continue
		}

		// 检查双字符操作符
		// 这里简化处理，实际应该预读
		if ch == '!' || ch == '>' || ch == '<' || ch == '=' || ch == '&' || ch == '|' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(ch))
			continue
		}

		current.WriteRune(ch)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	// 处理多词操作符（如 contains）和双字符操作符
	tokens = e.mergeTokens(tokens)

	return tokens, nil
}

// mergeTokens 合并多词关键字和双字符操作符
func (e *Engine) mergeTokens(tokens []string) []string {
	var result []string
	i := 0
	for i < len(tokens) {
		// 检查双字符操作符
		if i+1 < len(tokens) {
			twoChar := tokens[i] + tokens[i+1]
			if _, ok := precedence[twoChar]; ok {
				result = append(result, twoChar)
				i += 2
				continue
			}
		}
		// 检查 contains 关键字
		if i+1 < len(tokens) &&
			(strings.ToLower(tokens[i]) == "contains" ||
				(strings.ToLower(tokens[i]) == "con" && strings.ToLower(tokens[i+1]) == "tains")) {
			result = append(result, "contains")
			if strings.ToLower(tokens[i]) == "con" {
				i += 2
			} else {
				i++
			}
			continue
		}
		result = append(result, tokens[i])
		i++
	}
	return result
}

// parseExpression 递归下降解析
func (e *Engine) parseExpression(tokens []string, minPrec int) (*node, int, error) {
	if len(tokens) == 0 {
		return nil, 0, errors.New("空表达式")
	}

	// 解析左操作数
	left, pos, err := e.parsePrimary(tokens, 0)
	if err != nil {
		return nil, pos, err
	}

	// 解析操作符链
	for pos < len(tokens) {
		op := tokens[pos]
		prec, ok := precedence[op]
		if !ok || prec < minPrec {
			break
		}

		// 二元操作符
		pos++
		right, newPos, err := e.parseExpression(tokens[pos:], prec+1)
		if err != nil {
			return nil, pos, err
		}

		left = &node{
			typ:      nodeBinary,
			operator: op,
			left:     left,
			right:    right,
		}
		pos += newPos
	}

	return left, pos, nil
}

// parsePrimary 解析基本单元
func (e *Engine) parsePrimary(tokens []string, pos int) (*node, int, error) {
	if pos >= len(tokens) {
		return nil, pos, errors.New("意外的表达式结束")
	}

	token := tokens[pos]

	// 括号表达式
	if token == "(" {
		pos++
		// 找到匹配的右括号
		endPos := pos
		depth := 1
		for endPos < len(tokens) && depth > 0 {
			if tokens[endPos] == "(" {
				depth++
			} else if tokens[endPos] == ")" {
				depth--
			}
			endPos++
		}
		if depth > 0 {
			return nil, pos, errors.New("未闭合的括号")
		}

		// 解析括号内的表达式
		innerTokens := tokens[pos : endPos-1]
		node, _, err := e.parseExpression(innerTokens, 0)
		if err != nil {
			return nil, pos, err
		}
		return node, endPos, nil
	}

	// 一元操作符 !
	if token == "!" {
		pos++
		operand, newPos, err := e.parsePrimary(tokens, pos)
		if err != nil {
			return nil, pos, err
		}
		return &node{
			typ:      nodeUnary,
			operator: "!",
			right:    operand,
		}, newPos, nil
	}

	// 字符串字面量
	if strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"") {
		value := strings.Trim(token, "\"")
		return &node{
			typ:   nodeLiteral,
			value: value,
		}, pos + 1, nil
	}

	// 数字字面量
	if num, err := strconv.ParseFloat(token, 64); err == nil {
		return &node{
			typ:   nodeLiteral,
			value: num,
		}, pos + 1, nil
	}

	// 布尔字面量
	if token == "true" {
		return &node{
			typ:   nodeLiteral,
			value: true,
		}, pos + 1, nil
	}
	if token == "false" {
		return &node{
			typ:   nodeLiteral,
			value: false,
		}, pos + 1, nil
	}

	// 变量（标识符）
	if e.isIdentifier(token) {
		return &node{
			typ:   nodeVariable,
			value: token,
		}, pos + 1, nil
	}

	return nil, pos, fmt.Errorf("意外的 token: %s", token)
}

// isIdentifier 检查是否为合法标识符
func (e *Engine) isIdentifier(s string) bool {
	match, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, s)
	return match
}

// eval 求值 AST
func (e *Engine) eval(n *node, context map[string]interface{}) (bool, error) {
	switch n.typ {
	case nodeLiteral:
		return e.toBool(n.value), nil

	case nodeVariable:
		val, ok := context[n.value.(string)]
		if !ok {
			return false, fmt.Errorf("变量未定义: %s", n.value)
		}
		return e.toBool(val), nil

	case nodeUnary:
		if n.operator == "!" {
			val, err := e.eval(n.right, context)
			if err != nil {
				return false, err
			}
			return !val, nil
		}
		return false, fmt.Errorf("未知的一元操作符: %s", n.operator)

	case nodeBinary:
		return e.evalBinary(n, context)

	default:
		return false, fmt.Errorf("未知的节点类型: %d", n.typ)
	}
}

// evalBinary 求值二元操作
func (e *Engine) evalBinary(n *node, context map[string]interface{}) (bool, error) {
	switch n.operator {
	case "&&":
		left, err := e.eval(n.left, context)
		if err != nil {
			return false, err
		}
		if !left {
			return false, nil // 短路
		}
		return e.eval(n.right, context)

	case "||":
		left, err := e.eval(n.left, context)
		if err != nil {
			return false, err
		}
		if left {
			return true, nil // 短路
		}
		return e.eval(n.right, context)

	case "==", "!=", ">", "<", ">=", "<=":
		return e.evalComparison(n, context)

	case "contains":
		return e.evalContains(n, context)

	default:
		return false, fmt.Errorf("未知的二元操作符: %s", n.operator)
	}
}

// evalComparison 求值比较操作
func (e *Engine) evalComparison(n *node, context map[string]interface{}) (bool, error) {
	leftVal, err := e.getValue(n.left, context)
	if err != nil {
		return false, err
	}
	rightVal, err := e.getValue(n.right, context)
	if err != nil {
		return false, err
	}

	// 类型转换
	leftNum, leftIsNum := e.toNumber(leftVal)
	rightNum, rightIsNum := e.toNumber(rightVal)

	// 数字比较
	if leftIsNum && rightIsNum {
		switch n.operator {
		case "==":
			return leftNum == rightNum, nil
		case "!=":
			return leftNum != rightNum, nil
		case ">":
			return leftNum > rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		}
	}

	// 字符串比较
	leftStr := e.toString(leftVal)
	rightStr := e.toString(rightVal)

	switch n.operator {
	case "==":
		return leftStr == rightStr, nil
	case "!=":
		return leftStr != rightStr, nil
	case ">":
		return leftStr > rightStr, nil
	case "<":
		return leftStr < rightStr, nil
	case ">=":
		return leftStr >= rightStr, nil
	case "<=":
		return leftStr <= rightStr, nil
	}

	return false, fmt.Errorf("未知的比较操作符: %s", n.operator)
}

// evalContains 求值 contains 操作
func (e *Engine) evalContains(n *node, context map[string]interface{}) (bool, error) {
	leftVal, err := e.getValue(n.left, context)
	if err != nil {
		return false, err
	}
	rightVal, err := e.getValue(n.right, context)
	if err != nil {
		return false, err
	}

	// 字符串包含
	leftStr := e.toString(leftVal)
	rightStr := e.toString(rightVal)
	return strings.Contains(leftStr, rightStr), nil
}

// getValue 获取节点的值
func (e *Engine) getValue(n *node, context map[string]interface{}) (interface{}, error) {
	switch n.typ {
	case nodeLiteral:
		return n.value, nil
	case nodeVariable:
		val, ok := context[n.value.(string)]
		if !ok {
			return nil, fmt.Errorf("变量未定义: %s", n.value)
		}
		return val, nil
	default:
		return nil, fmt.Errorf("非字面量或变量节点")
	}
}

// toBool 转换为布尔值
func (e *Engine) toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != "" && val != "false" && val != "0"
	case float64:
		return val != 0
	case int:
		return val != 0
	case int64:
		return val != 0
	default:
		return true
	}
}

// toNumber 转换为数字
func (e *Engine) toNumber(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			return num, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toString 转换为字符串
func (e *Engine) toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}
