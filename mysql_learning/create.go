package mysql_learning

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// 用户结构体
type Users struct {
	UserId   int    `db:"user_id"`
	Username string `db:"username"`
	Sex      string `db:"sex"`
	Email    string `db:"email"`
}

// 数据库指针
var db *sqlx.DB

// 初始化数据库连接，init() 方法会在 main 方法之前执行
func init() {
	database, err := sqlx.Open("mysql", "root:yezichao@tcp(127.0.0.1:3306)/mytest")
	if err != nil {
		fmt.Println("open mysql failed,", err)
	}
	db = database
}

// 测试时需要在 main 包下运行
func Main() {
	sql := "insert into user(username,sex, email)values (?,?,?)"
	value := [3]string{"user02", "woman", "user02@163.com"}

	//执行SQL语句
	r, err := db.Exec(sql, value[0], value[1], value[2])
	if err != nil {
		fmt.Println("exec failed,", err)
		return
	}

	//查询最后一天用户ID，判断是否插入成功
	id, err := r.LastInsertId()
	if err != nil {
		fmt.Println("exec failed,", err)
		return
	}
	fmt.Println("insert succ", id)

}
