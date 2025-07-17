package log

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// LogEntry 表示一个日志条目
type LogEntry struct {
	Level     Loglevel
	Message   string
	Timestamp time.Time
	Caller    string
	Context   context.Context
}

// Middleware 中间件接口
type Middleware interface {
	// Process 处理日志条目，返回是否继续处理
	Process(entry *LogEntry) bool
}

// MiddlewareFunc 中间件函数类型
type MiddlewareFunc func(entry *LogEntry) bool

// Process 实现 Middleware 接口
func (f MiddlewareFunc) Process(entry *LogEntry) bool {
	return f(entry)
}

// MiddlewareChain 中间件链
type MiddlewareChain struct {
	middlewares []Middleware
}

// NewMiddlewareChain 创建新的中间件链
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]Middleware, 0),
	}
}

// Use 添加中间件到链中
func (mc *MiddlewareChain) Use(middleware Middleware) {
	mc.middlewares = append(mc.middlewares, middleware)
}

// Process 执行中间件链
func (mc *MiddlewareChain) Process(entry *LogEntry) bool {
	for _, middleware := range mc.middlewares {
		if !middleware.Process(entry) {
			return false
		}
	}
	return true
}

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
	// 日志字体颜色
	levelFontColors = map[string]string{
		"DEBUG":   "\033[36m", // 青色
		"INFO":    "\033[37m", // 白色
		"WARNING": "\033[33m", // 黄色
		"ERROR":   "\033[31m", // 红色
		"PANIC":   "\033[35m", // 紫色
		"SUCCESS": "\033[32m", // 绿色
	}
	// 默认日志等级
	currentLevel *Loglevel

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
		"DEBUG":   "\033[97;46m",   // 蓝色
		"INFO":    "\033[30;47m",   // 白色背景黑色文字
		"WARNING": "\033[30;43m",   // 黑色文字配黄色背景
		"ERROR":   "\033[97;41m",   // 红色
		"PANIC":   "\033[97;45m",   // 紫色
		"SUCCESS": "\033[1;97;42m", // 绿色加粗
	}

	// 日志输出目标
	logOutputs = []logOutput{
		{writer: os.Stdout, isConsole: true},
	}
	lastLogFileDate = "" // 记录上一次写入日志的日期
	logFile         *os.File

	// 全局中间件链
	middlewareChain = NewMiddlewareChain()
)

// 日志输出结构
type logOutput struct {
	writer    *os.File
	isConsole bool
}

func init() {
	level := DEBUG
	currentLevel = &level
}

// UseMiddleware 添加中间件到全局中间件链
func UseMiddleware(middleware Middleware) {
	middlewareChain.Use(middleware)
}

// UseMiddlewareFunc 添加中间件函数到全局中间件链
func UseMiddlewareFunc(fn func(entry *LogEntry) bool) {
	middlewareChain.Use(MiddlewareFunc(fn))
}

// ClearMiddlewares 清空所有中间件
func ClearMiddlewares() {
	middlewareChain = NewMiddlewareChain()
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
	*currentLevel = level
}

// 判断是否输出该等级日志
func shouldLog(level Loglevel) bool {
	return levelPriority[level] >= levelPriority[*currentLevel]
}

// 日志格式化（带颜色，用于控制台输出）
func formatLogColored(level string, message string) string {
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

	return fmt.Sprintf("\033[32m%s\033[0m |\033[0m%s%s\033[0m| \033[35m%s\033[0m - %s%s\033[0m",
		now, levelColors[level], levelStr, caller, levelFontColors[level], message)
}

// 日志格式化（不带颜色，用于文件输出）
func formatLogPlain(level string, message string) string {
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

	return fmt.Sprintf("%s |%s| %s - %s",
		now, levelStr, caller, message)
}

// 通用日志输出
func logWithLevel(level string, v ...any) {
	lv := Loglevel(level)
	if !shouldLog(lv) {
		return
	}

	var parts []string
	for _, arg := range v {
		val := reflect.ValueOf(arg)
		kind := val.Kind()
		if kind == reflect.Ptr {
			if val.IsNil() {
				parts = append(parts, "<nil>")
				continue
			}
			val = val.Elem()
			kind = val.Kind()
		}

		if err, ok := arg.(error); ok {
			// 特殊处理 error 接口类型
			parts = append(parts, err.Error())
		} else if kind == reflect.Struct || kind == reflect.Map || kind == reflect.Array {
			jsonData, err := json.MarshalIndent(arg, "", "  ") // 使用 MarshalIndent 并设置缩进
			if err == nil {
				parts = append(parts, val.Type().Name(), " "+string(jsonData)) // 在 JSON 前添加换行符
			} else {
				// 序列化失败，回退到 Sprint
				parts = append(parts, fmt.Sprint(arg))
			}
		} else {
			// 不是结构体，使用 Sprint
			parts = append(parts, fmt.Sprint(arg))
		}
	}
	message := strings.Join(parts, " ") // 使用空格连接各个部分

	// 创建日志条目
	entry := createLogEntry(lv, message)
	
	// 通过中间件链处理
	if !middlewareChain.Process(entry) {
		return
	}

	coloredLogLine := formatLogColored(level, message)
	plainLogLine := formatLogPlain(level, message)
	writeLogAllTargets(coloredLogLine, plainLogLine)
}

// createLogEntry 创建日志条目
func createLogEntry(level Loglevel, message string) *LogEntry {
	now := time.Now()
	pc, file, line, ok := runtime.Caller(4)
	caller := "unknown"
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		caller = fmt.Sprintf("%s:%d", file, line)
		if index := strings.LastIndex(funcName, "."); index != -1 {
			caller = fmt.Sprintf("%s.%s:%d", funcName[:index], funcName[index+1:], line)
		}
	}

	return &LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: now,
		Caller:    caller,
		Context:   context.Background(),
	}
}

// 支持格式化输出
func logWithLevelf(level string, format string, v ...any) {
	lv := Loglevel(level)
	if !shouldLog(lv) {
		return
	}
	message := fmt.Sprintf(format, v...)
	
	// 创建日志条目
	entry := createLogEntry(lv, message)
	
	// 通过中间件链处理
	if !middlewareChain.Process(entry) {
		return
	}
	
	coloredLogLine := formatLogColored(level, message)
	plainLogLine := formatLogPlain(level, message)
	writeLogAllTargets(coloredLogLine, plainLogLine)
}

// Logger 结构体，用于兼容 外部的 Logger 接口
//
// 已兼容 Wails V2 版本的 Logger 接口
//
// TODO: 后续考虑对外进行扩展
type Logger struct{}

// NewLogger 创建一个新的 Logger 实例
func NewLogger() *Logger {
	return &Logger{}
}

// Print 实现 Logger 接口的 Print 方法
func (l *Logger) Print(message string) {
	logWithLevel(string(INFO), message) // 使用 INFO 级别作为默认 Print 级别
}

// Trace 实现 Logger 接口的 Trace 方法 (映射到 DEBUG)
func (l *Logger) Trace(message string) {
	logWithLevel(string(DEBUG), message)
}

// Debug 实现 Logger 接口的 Debug 方法
func (l *Logger) Debug(message string) {
	logWithLevel(string(DEBUG), message)
}

// Info 实现 Logger 接口的 Info 方法
func (l *Logger) Info(message string) {
	logWithLevel(string(INFO), message)
}

// Warning 实现 Logger 接口的 Warning 方法
func (l *Logger) Warning(message string) {
	logWithLevel(string(WARNING), message)
}

// Error 实现 Logger 接口的 Error 方法
func (l *Logger) Error(message string) {
	logWithLevel(string(ERROR), message)
}

// Fatal 实现 Logger 接口的 Fatal 方法 (映射到 PANIC 并退出)
func (l *Logger) Fatal(message string) {
	logWithLevel(string(PANIC), message)
	os.Exit(1)
}

// 标准库兼容方法
func Print(v ...any) {
	logWithLevel(string(INFO), v...)
}

func Printf(format string, v ...any) {
	logWithLevelf("INFO", format, v...)
}

func Println(v ...any) {
	logWithLevel(string(INFO), v...)
}

func Fatal(v ...any) {
	logWithLevel(string(ERROR), v...)
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	logWithLevelf(string(ERROR), format, v...)
	os.Exit(1)
}

func Fatalln(v ...any) {
	logWithLevel(string(ERROR), v...)
	os.Exit(1)
}

func Panic(v ...any) {
	s := fmt.Sprint(v...)
	logWithLevel(string(PANIC), s)
	panic(s)
}

func Panicf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	logWithLevel(string(PANIC), s)
	panic(s)
}

func Panicln(v ...any) {
	s := fmt.Sprintln(v...)
	logWithLevel(string(PANIC), s)
	panic(s)
}

// 扩展方法
func Debug(v ...any)    { logWithLevel("DEBUG", v...) }
func Info(v ...any)     { logWithLevel(string(INFO), v...) }
func Warning(v ...any)  { logWithLevel("WARNING", v...) }
func Error(v ...any)    { logWithLevel(string(ERROR), v...) }
func Critical(v ...any) { logWithLevel("PANIC", v...) }
func Success(v ...any)  { logWithLevel("SUCCESS", v...) }

func Debugf(format string, v ...any)    { logWithLevelf(string(DEBUG), format, v...) }
func Infof(format string, v ...any)     { logWithLevelf("INFO", format, v...) }
func Warningf(format string, v ...any)  { logWithLevelf(string(WARNING), format, v...) }
func Errorf(format string, v ...any)    { logWithLevelf(string(ERROR), format, v...) }
func Criticalf(format string, v ...any) { logWithLevelf(string(PANIC), format, v...) }
func Successf(format string, v ...any)  { logWithLevelf(string(SUCCESS), format, v...) }

// ==================== 常用中间件示例 ====================

// FilterMiddleware 日志级别过滤中间件
type FilterMiddleware struct {
	MinLevel Loglevel
}

// NewFilterMiddleware 创建级别过滤中间件
func NewFilterMiddleware(minLevel Loglevel) *FilterMiddleware {
	return &FilterMiddleware{MinLevel: minLevel}
}

// Process 实现 Middleware 接口
func (f *FilterMiddleware) Process(entry *LogEntry) bool {
	return levelPriority[entry.Level] >= levelPriority[f.MinLevel]
}

// DatabaseMiddleware 数据库日志中间件示例
type DatabaseMiddleware struct {
	TableName string
	Callback  func(entry *LogEntry) error
}

// NewDatabaseMiddleware 创建数据库中间件
func NewDatabaseMiddleware(tableName string, callback func(entry *LogEntry) error) *DatabaseMiddleware {
	return &DatabaseMiddleware{
		TableName: tableName,
		Callback:  callback,
	}
}

// Process 实现 Middleware 接口
func (d *DatabaseMiddleware) Process(entry *LogEntry) bool {
	if d.Callback != nil {
		if err := d.Callback(entry); err != nil {
			// 数据库写入失败，可以选择记录错误或继续处理
			fmt.Printf("Database middleware error: %v\n", err)
		}
	}
	return true // 继续处理其他中间件
}

// HTTPMiddleware HTTP请求日志中间件示例
type HTTPMiddleware struct {
	URL     string
	Headers map[string]string
	Timeout time.Duration
}

// NewHTTPMiddleware 创建HTTP中间件
func NewHTTPMiddleware(url string, headers map[string]string, timeout time.Duration) *HTTPMiddleware {
	return &HTTPMiddleware{
		URL:     url,
		Headers: headers,
		Timeout: timeout,
	}
}

// Process 实现 Middleware 接口
func (h *HTTPMiddleware) Process(entry *LogEntry) bool {
	// 这里可以实现HTTP请求发送逻辑
	// 为了避免引入额外依赖，这里只是示例
	go func() {
		// 异步发送HTTP请求，避免阻塞日志输出
		// 实际实现中可以使用 net/http 包
		fmt.Printf("Sending log to %s: [%s] %s\n", h.URL, entry.Level, entry.Message)
	}()
	return true
}

// FileFilterMiddleware 文件输出过滤中间件
type FileFilterMiddleware struct {
	FilePath   string
	LevelFilter map[Loglevel]bool
}

// NewFileFilterMiddleware 创建文件过滤中间件
func NewFileFilterMiddleware(filePath string, allowedLevels []Loglevel) *FileFilterMiddleware {
	filter := make(map[Loglevel]bool)
	for _, level := range allowedLevels {
		filter[level] = true
	}
	return &FileFilterMiddleware{
		FilePath:    filePath,
		LevelFilter: filter,
	}
}

// Process 实现 Middleware 接口
func (f *FileFilterMiddleware) Process(entry *LogEntry) bool {
	if f.LevelFilter[entry.Level] {
		// 异步写入文件
		go func() {
			file, err := os.OpenFile(f.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return
			}
			defer file.Close()
			
			logLine := fmt.Sprintf("%s |%-8s| %s - %s\n",
				entry.Timestamp.Format("2006-01-02 15:04:05"),
				entry.Level,
				entry.Caller,
				entry.Message)
			file.WriteString(logLine)
		}()
	}
	return true
}
