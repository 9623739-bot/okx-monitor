package services

import (
	"sync"
	"time"
)

// ErrorEntry 去重错误条目
type ErrorEntry struct {
	Message   string    `json:"message"`
	Count     int       `json:"count"`
	FirstTime time.Time `json:"first_time"`
	LastTime  time.Time `json:"last_time"`
}

var (
	errMu     sync.Mutex
	errMap    = make(map[string]*ErrorEntry) // key: 去重后的错误消息
	errList   []*ErrorEntry                   // 按首次出现排序
	maxErrors = 100                           // 最多保留 100 种不同错误
)

// LogError 记录错误（相同消息去重，只增加计数）
func LogError(msg string) {
	if msg == "" {
		return
	}

	errMu.Lock()
	defer errMu.Unlock()

	if entry, ok := errMap[msg]; ok {
		entry.Count++
		entry.LastTime = time.Now()
		return
	}

	entry := &ErrorEntry{
		Message:   msg,
		Count:     1,
		FirstTime: time.Now(),
		LastTime:  time.Now(),
	}
	errMap[msg] = entry
	errList = append(errList, entry)

	// 超过上限时清除最旧的
	if len(errList) > maxErrors {
		delete(errMap, errList[0].Message)
		errList = errList[1:]
	}
}

// GetErrors 获取所有错误（按最近发生排序）
func GetErrors() []*ErrorEntry {
	errMu.Lock()
	defer errMu.Unlock()

	// 按 LastTime 降序排列
	result := make([]*ErrorEntry, len(errList))
	copy(result, errList)

	// 简单插入排序（数量不超过100，可接受）
	for i := 1; i < len(result); i++ {
		j := i
		for j > 0 && result[j].LastTime.After(result[j-1].LastTime) {
			result[j], result[j-1] = result[j-1], result[j]
			j--
		}
	}

	return result
}
