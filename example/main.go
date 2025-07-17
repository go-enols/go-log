package main

import (
	"fmt"
	"time"

	"github.com/go-enols/go-log"
)

type Test struct {
	Id   int
	Name string
}

func main() {
	// 基础日志方法
	log.Debug("这是一条调试日志")
	log.Info("这是一条信息日志")
	log.Warning("这是一条警告日志")
	log.Error("这是一条错误日志")
	log.Success("这是一条成功日志")
	log.Debug("这是一条调试日志", Test{Id: 1, Name: "test"})
	log.Debug("这是一条调试日志", &Test{Id: 1, Name: "test"})
	log.Debug("这是一条调试日志", map[string]any{"map": 1, "aa": "test"})

	// 格式化日志方法
	log.Debugf("这是一条格式化的调试日志: %s", "debug")
	log.Infof("这是一条格式化的信息日志: %d", 123)
	log.Warningf("这是一条格式化的警告日志: %v", []string{"warn", "warning"})
	log.Errorf("这是一条格式化的错误日志: %t", true)
	log.Successf("这是一条格式化的成功日志: %.2f", 3.1415926)

	// 标准库兼容方法
	log.Print("这是一条标准输出日志")
	log.Printf("这是一条格式化的标准输出日志: %s", "standard")
	log.Println("这是一条标准输出日志（带换行）")

	fmt.Println("\n=== 中间件功能演示 ===")

	// 1. 添加数据库中间件示例
	dbMiddleware := log.NewDatabaseMiddleware("logs", func(entry *log.LogEntry) error {
		// 模拟数据库写入
		fmt.Printf("[DB] 保存日志到数据库: %s - %s\n", entry.Level, entry.Message)
		return nil
	})
	log.UseMiddleware(dbMiddleware)

	// 2. 添加HTTP中间件示例
	httpMiddleware := log.NewHTTPMiddleware(
		"https://api.example.com/logs",
		map[string]string{"Authorization": "Bearer token"},
		5*time.Second,
	)
	log.UseMiddleware(httpMiddleware)

	// 3. 添加文件过滤中间件（只记录ERROR和PANIC级别到特定文件）
	errorFileMiddleware := log.NewFileFilterMiddleware(
		"error.log",
		[]log.Loglevel{log.ERROR, log.PANIC},
	)
	log.UseMiddleware(errorFileMiddleware)

	// 4. 添加自定义中间件函数
	log.UseMiddlewareFunc(func(entry *log.LogEntry) bool {
		// 自定义处理逻辑
		if entry.Level == log.SUCCESS {
			fmt.Printf("[CUSTOM] 检测到成功日志: %s\n", entry.Message)
		}
		return true // 继续处理
	})

	fmt.Println("\n=== 测试中间件功能 ===")

	// 测试不同级别的日志，观察中间件的处理
	log.Info("这是一条信息日志 - 会被所有中间件处理")
	log.Error("这是一条错误日志 - 会被写入error.log文件")
	log.Success("这是一条成功日志 - 会触发自定义中间件")
	log.Warning("这是一条警告日志")

	// 等待异步操作完成
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n=== 清空中间件并继续测试 ===")

	// 清空所有中间件
	log.ClearMiddlewares()

	// 现在日志只会正常输出，不会触发中间件
	log.Info("清空中间件后的日志 - 不会触发任何中间件")
	log.Error("这条错误日志不会被写入error.log")

	// 设置日志级别
	log.SetLogLevel(log.WARNING)
	log.Debug("这条调试日志不会显示")  // 不会显示，因为级别低于 WARNING
	log.Warning("这条警告日志会显示") // 会显示
	log.Error("这条错误日志会显示")   // 会显示

	// 添加文件输出
	err := log.AddLogOutputFile("app.log")
	if err != nil {
		log.Error("添加日志文件失败:", err)
		return
	}

	// 输出到控制台和文件
	log.Info("这条日志会同时输出到控制台和文件")

	log.Debug("这条调试日志不会显示")  // 不会显示，因为级别低于 WARNING
	log.Warning("这条警告日志会显示") // 会显示
	log.Error("这条错误日志会显示")   // 会显示

	// Panic 级别日志（会触发 panic，需要注释掉后面的代码才能看到所有输出）
	// log.Panic("这是一条 panic 日志")
	// log.Panicf("这是一条格式化的 panic 日志: %s", "panic")
	// log.Panicln("这是一条 panic 日志（带换行）")

	// Fatal 级别日志（会导致程序退出，需要注释掉才能看到后面的输出）
	// log.Fatal("这是一条 fatal 日志")
	// log.Fatalf("这是一条格式化的 fatal 日志: %s", "fatal")
	// log.Fatalln("这是一条 fatal 日志（带换行）")

	// Critical/Panic 方法（会以 PANIC 级别输出）
	// log.Critical("这是一条严重错误日志")
	// log.Criticalf("这是一条格式化的严重错误日志: %s", "critical")
}
