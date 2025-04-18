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
	DEBUG   Loglevel = "DEBUG"
	INFO    Loglevel = "INFO"
	WARNING Loglevel = "WARNING"
	ERROR   Loglevel = "ERROR"
	PANIC   Loglevel = "PANIC"
	SUCCESS Loglevel = "SUCCESS"
)

var (
	// 默认日志等级
	currentLevel = DEBUG

	// 日志等级优先级
	levelPriority = map[Loglevel]int{
		DEBUG:   1,
		INFO:    2,
		WARNING: 3,
		ERROR:   4,
		PANIC:   5,
		SUCCESS: 6,
	}

	// 日志颜色
	levelColors = map[string]string{
		"DEBUG":   "\033[97;46m", // 蓝色
		"INFO":    "\033[97;42m", // 绿色
		"WARNING": "\033[97;43m", // 黄色
		"ERROR":   "\033[97;41m", // 红色
		"PANIC":   "\033[97;45m", // 紫色
		"SUCCESS": "\033[97;42m", // 绿色加粗
	}

	// 日志输出目标
	logOutputs = []logOutput{
		{writer: os.Stdout, isConsole: true},
	}
	lastLogFileDate = "" // 记录上一次写入日志的日期
	logFile         *os.File
)

// 日志输出结构
type logOutput struct {
	writer    *os.File
	isConsole bool
}

// 添加日志输出目标（如文件）
func AddLogOutputFile(filepath string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logOutputs = append(logOutputs, logOutput{writer: file, isConsole: false})
	return nil
}

// 每次写日志前检查日期，切换文件
func ensureLogFile() {
	today := time.Now().Format("2006-01-02")
	if lastLogFileDate == today && logFile != nil {
		return
	}
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	_ = os.MkdirAll("log", 0755)
	logPath := filepath.Join("log", today+".log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		logFile = file
		found := false
		for _, out := range logOutputs {
			if out.writer == file {
				found = true
				break
			}
		}
		if !found {
			logOutputs = append(logOutputs, logOutput{writer: file, isConsole: false})
		}
		lastLogFileDate = today
	}
}

// 日志输出到所有目标
func writeLogAllTargets(coloredMsg string, plainMsg string) {
	ensureLogFile()
	for _, out := range logOutputs {
		if out.isConsole {
			fmt.Fprintln(out.writer, coloredMsg)
		} else {
			fmt.Fprintln(out.writer, plainMsg)
		}
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
func formatLog(level string, message string, withColor bool) string {
	now := time.Now().Format("2006-01-02 15:04:05")
	pc, file, line, ok := runtime.Caller(3)
	caller := "unknown"
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		caller = fmt.Sprintf("%s:%d", file, line)
		if index := strings.LastIndex(funcName, "."); index != -1 {
			caller = fmt.Sprintf("%s.%s:%d", funcName[:index], funcName[index+1:], line)
		}
	}

	// 统一日志级别的宽度为8个字符
	levelStr := fmt.Sprintf("%-8s", level)

	if withColor {
		return fmt.Sprintf("\033[32m%s\033[0m |\033[0m%s%s\033[0m| \033[35m%s\033[0m - \033[33m%s\033[0m",
			now, levelColors[level], levelStr, caller, message)
	}
	return fmt.Sprintf("%s |%s| %s - %s",
		now, levelStr, caller, message)
}

// 通用日志输出
func logWithLevel(level string, v ...any) {
	lv := Loglevel(level)
	if !shouldLog(lv) {
		return
	}
	message := fmt.Sprint(v...)
	coloredLogLine := formatLog(level, message, true)
	plainLogLine := formatLog(level, message, false)
	writeLogAllTargets(coloredLogLine, plainLogLine)
}

// 支持格式化输出
func logWithLevelf(level string, format string, v ...any) {
	lv := Loglevel(level)
	if !shouldLog(lv) {
		return
	}
	message := fmt.Sprintf(format, v...)
	coloredLogLine := formatLog(level, message, true)
	plainLogLine := formatLog(level, message, false)
	writeLogAllTargets(coloredLogLine, plainLogLine)
}

// 标准库兼容方法
func Print(v ...any) {
	logWithLevel("INFO", v...)
}

func Printf(format string, v ...any) {
	logWithLevelf("INFO", format, v...)
}

func Println(v ...any) {
	logWithLevel("INFO", v...)
}

func Fatal(v ...any) {
	logWithLevel("ERROR", v...)
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	logWithLevelf("ERROR", format, v...)
	os.Exit(1)
}

func Fatalln(v ...any) {
	logWithLevel("ERROR", v...)
	os.Exit(1)
}

func Panic(v ...any) {
	s := fmt.Sprint(v...)
	logWithLevel("PANIC", s)
	panic(s)
}

func Panicf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	logWithLevel("PANIC", s)
	panic(s)
}

func Panicln(v ...any) {
	s := fmt.Sprintln(v...)
	logWithLevel("PANIC", s)
	panic(s)
}

// 扩展方法
func Debug(v ...any)    { logWithLevel("DEBUG", v...) }
func Info(v ...any)     { logWithLevel("INFO", v...) }
func Warning(v ...any)  { logWithLevel("WARNING", v...) }
func Error(v ...any)    { logWithLevel("ERROR", v...) }
func Critical(v ...any) { logWithLevel("PANIC", v...) }
func Success(v ...any)  { logWithLevel("SUCCESS", v...) }

func Debugf(format string, v ...any)    { logWithLevelf("DEBUG", format, v...) }
func Infof(format string, v ...any)     { logWithLevelf("INFO", format, v...) }
func Warningf(format string, v ...any)  { logWithLevelf("WARNING", format, v...) }
func Errorf(format string, v ...any)    { logWithLevelf("ERROR", format, v...) }
func Criticalf(format string, v ...any) { logWithLevelf("PANIC", format, v...) }
func Successf(format string, v ...any)  { logWithLevelf("SUCCESS", format, v...) }
