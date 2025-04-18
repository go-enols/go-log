package log

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// 自定义日志格式
type CustomLogger struct {
	logger *log.Logger
}

var levelColors map[string]string = map[string]string{
	"DEBUG":    "[97;46m", // 蓝色
	"INFO":     "[97;42m", // 绿色
	"WARNING":  "[97;43m", // 黄色
	"ERROR":    "[97;41m", // 红色
	"CRITICAL": "[97;45m", // 红色加粗
	"SUCCESS":  "[97;42m", // 绿色加粗
}

// 初始化自定义日志
func NewCustomLogger() *CustomLogger {
	return &CustomLogger{
		logger: log.New(os.Stdout, "", 0), // 输出到标准输出，不带前缀
	}
}

// 自定义日志格式
func (l *CustomLogger) formatLog(level string, message string) string {
	// 获取当前时间
	now := time.Now().Format("2006-01-02 15:04:05.000")

	// 获取调用栈信息
	pc, file, line, ok := runtime.Caller(3) // 跳过三层调用栈
	caller := "unknown"
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		caller = fmt.Sprintf("%s:%d", file, line)
		if index := strings.LastIndex(funcName, "."); index != -1 {
			caller = fmt.Sprintf("%s.%s:%d", funcName[:index], funcName[index+1:], line)
		}
	}

	// 构建日志格式
	return fmt.Sprintf("[32m%s[0m |%s%s[0m| [35m%s[0m - [33m%s[0m",
		now, levelColors[level], level, caller, message)
}

// 日志输出函数
func (l *CustomLogger) Log(level string, message string) {
	logLine := l.formatLog(level, message)
	l.logger.Println(logLine)
}

// 不同级别的日志函数
func (l *CustomLogger) Info(message string) {
	l.Log("INFO", message)
}

func (l *CustomLogger) Infof(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	l.Log("INFO", message)
}

func (l *CustomLogger) Error(message string) {
	l.Log("ERROR", message)
}

func (l *CustomLogger) Errorf(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	l.Log("ERROR", message)
}

func (l *CustomLogger) Debug(message string) {
	l.Log("DEBUG", message)
}

func (l *CustomLogger) Debugf(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	l.Log("DEBUG", message)
}
