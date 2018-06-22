package packet

import "fmt"

func Error(pre, what string, should, get interface{}) error {
	return fmt.Errorf("[%s]%s,should %v but %v", pre, what, should, get)
}