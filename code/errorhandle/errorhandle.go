package main

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

// dao 模拟数据层返回错误，需要Wrap下错误，携带堆栈、相关dao层信息，帮助开发同学定位问题
func dao() error {
	return errors.Wrap(sql.ErrNoRows, "dao failed")
}

// logic 模拟业务服务逻辑处理，可以将自己这层相关信息带回调用方
func logic() error {
	return errors.WithMessage(dao(), "logic failed")
}

// 使用 errors.Cause(err) == sql.ErrNoRows 或者 errors.Is(err, sql.ErrNoRows)
// 来做错误判断，通过%+v来打印堆栈内容
func main() {
	err := logic()
	if errors.Cause(err) == sql.ErrNoRows {
		fmt.Printf("%+v\n",  err)
		return
	}
	// if errors.Is(err, sql.ErrNoRows) {
	// 	fmt.Printf("%+v\n", err)
	// 	return
	// }
}

/*Output:
root@VM-1-130-centos errorhandle [main] $ ./errorhandle
sql: no rows in result set
dao failed
main.dao
        /root/source/github/go-dev/code/errorhandle/errorhandle.go:12
main.logic
        /root/source/github/go-dev/code/errorhandle/errorhandle.go:17
main.main
        /root/source/github/go-dev/code/errorhandle/errorhandle.go:21
runtime.main
        /usr/local/go/src/runtime/proc.go:255
runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:1581
logic failed
*/
