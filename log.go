package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Loglevel string

const (
	DEBUG    Loglevel = "DEBUG"
	INFO     Loglevel = "INFO"
	WARNING  Loglevel = "WARNING"
	ERROR    Loglevel = "ERROR"
	CRITICAL Loglevel = "CRITICAL"
	SUCCESS  Loglevel = "SUCCESS"
)

var (
	// 默认日志等级
	currentLevel = DEBUG

	// 日志等级优先级
	levelPriority = map[Loglevel]int{
		DEBUG:    1,
		INFO:     2,
		WARNING:  3,
		ERROR:    4,
		CRITICAL: 5,
		SUCCESS:  6,
	}

	// 日志颜色
	levelColors = map[string]string{
		"DEBUG":    "[97;46m", // 蓝色
		"INFO":     "[97;42m", // 绿色
		"WARNING":  "[97;43m", // 黄色
		"ERROR":    "[97;41m", // 红色
		"CRITICAL": "[97;45m", // 红色加粗
		"SUCCESS":  "[97;42m", // 绿色加粗
	}

	// 日志输出目标
	logOutputs = []logOutput{
		{writer: os.Stdout},
	}
	lastLogFileDate = "" // 记录上一次写入日志的日期
	logFile         *os.File
)

// 日志输出结构
type logOutput struct {
	writer *os.File
}

// 添加日志输出目标（如文件）
func AddLogOutputFile(filepath string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logOutputs = append(logOutputs, logOutput{writer: file})
	return nil
}

// 每次写日志前检查日期，切换文件
func ensureLogFile() {
	today := time.Now().Format("2006-01-02")
	if lastLogFileDate == today && logFile != nil {
		return
	}
	// 关闭旧文件
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	// 创建log目录
	_ = os.MkdirAll("log", 0755)
	logPath := filepath.Join("log", today+".log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		logFile = file
		// 检查是否已存在，避免重复添加
		found := false
		for _, out := range logOutputs {
			if out.writer == file {
				found = true
				break
			}
		}
		if !found {
			logOutputs = append(logOutputs, logOutput{writer: file})
		}
		lastLogFileDate = today
	}
}

// 日志输出到所有目标（自动切换文件）
func writeLogAllTargets(msg string) {
	ensureLogFile()
	for _, out := range logOutputs {
		fmt.Fprintln(out.writer, msg)
	}
}

// 设置日志等级
func SetLogLevel(level Loglevel) {
	currentLevel = level
}

// 判断是否输出该等级日志
func shouldLog(level Loglevel) bool {
	return levelPriority[level] >= levelPriority[currentLevel]
}

// 日志格式化
func formatLog(level string, message string) string {
	now := time.Now().Format("2006-01-02 15:04:05.000")
	pc, file, line, ok := runtime.Caller(3)
	caller := "unknown"
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		caller = fmt.Sprintf("%s:%d", file, line)
		if index := strings.LastIndex(funcName, "."); index != -1 {
			caller = fmt.Sprintf("%s.%s:%d", funcName[:index], funcName[index+1:], line)
		}
	}
	return fmt.Sprintf("[32m%s[0m |%s%s[0m| [35m%s[0m - [33m%s[0m",
		now, levelColors[level], level, caller, message)
}

// 通用日志输出
func logWithLevel(level string, v ...any) {
	lv := Loglevel(level)
	if !shouldLog(lv) {
		return
	}
	message := fmt.Sprint(v...)
	logLine := formatLog(level, message)
	writeLogAllTargets(logLine)
}

// 支持格式化输出
func logWithLevelf(level string, format string, v ...any) {
	lv := Loglevel(level)
	if !shouldLog(lv) {
		return
	}
	message := fmt.Sprintf(format, v...)
	logLine := formatLog(level, message)
	writeLogAllTargets(logLine)
}

// 公开API
func Debug(v ...any)    { logWithLevel("DEBUG", v...) }
func Info(v ...any)     { logWithLevel("INFO", v...) }
func Warning(v ...any)  { logWithLevel("WARNING", v...) }
func Error(v ...any)    { logWithLevel("ERROR", v...) }
func Critical(v ...any) { logWithLevel("CRITICAL", v...) }
func Success(v ...any)  { logWithLevel("SUCCESS", v...) }

func Debugf(format string, v ...any)    { logWithLevelf("DEBUG", format, v...) }
func Infof(format string, v ...any)     { logWithLevelf("INFO", format, v...) }
func Warningf(format string, v ...any)  { logWithLevelf("WARNING", format, v...) }
func Errorf(format string, v ...any)    { logWithLevelf("ERROR", format, v...) }
func Criticalf(format string, v ...any) { logWithLevelf("CRITICAL", format, v...) }
func Successf(format string, v ...any)  { logWithLevelf("SUCCESS", format, v...) }
