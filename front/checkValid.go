package front

import (
	"unicode"
)

// validExpression проверка валидности введённого выражения
func validExpression(exp string) bool {
	if len(exp) == 0 {
		return false
	}
	if len(exp) == 1 {
		if unicode.IsDigit(rune(exp[0])) {
			return true
		} else {
			return false
		}
	}

	//певый и последний не могут быть не цифрами но могут быть скобками
	if exp[0] != byte('(') && exp[0] != byte(')') && exp[len(exp)-1] != byte('(') && exp[len(exp)-1] != byte(')') {
		if !unicode.IsDigit(rune(exp[0])) || !unicode.IsDigit(rune(exp[len(exp)-1])) {
			return false
		}
	}

	flag := false
	countBrecket := 0 //колличество открывающих и закрыающих скобок должно в конце быть 0  и положительным
	for i, e := range exp {
		//два раза подряд символы не могут идти за исключением скобок
		if !unicode.IsDigit(e) && e != '(' && e != ')' && e != '*' && e != '/' && e != '+' && e != '-' {
			return false
		}
		//variant ++
		if e == '+' || e == '-' || e == '*' || e == '/' {
			if flag == true {
				return false
			}
			flag = true
		} else {
			// variant +)
			if e == ')' {
				if flag == true {
					return false
				}
			}
			//variant 2(
			if e == '(' {
				if i != 0 && i != len(exp)-1 && flag == false {
					return false
				}
				flag = true
			}
			flag = false
			//variant )2
			if i != 0 && unicode.IsDigit(e) && exp[i-1] == byte(')') {
				return false
			}
		}
		if e == '(' {
			countBrecket++
		}
		if e == ')' {
			countBrecket--
		}
		if countBrecket < 0 {
			return false
		}
	}

	if countBrecket != 0 {
		return false
	}

	return true
}
