package main

import (
 "context"
 "sync"
 "time"
)

// Config структура для хранения конфигурации флуд-контроля
type Config struct {
 Interval  time.Duration // Интервал времени, за который считаем запросы
 MaxChecks int           // Максимальное количество запросов за интервал времени
}

// FloodControl интерфейс, который нужно реализовать
type FloodControl interface {
 // Check возвращает false если достигнут лимит максимально разрешенного
 // кол-ва запросов согласно заданным правилам флуд контроля.
 Check(ctx context.Context, userID int64) (bool, error)
}

// floodControl реализация интерфейса FloodControl
type floodControl struct {
 config     Config
 checks     map[int64][]time.Time
 checksLock sync.RWMutex
}

// NewFloodControl создает новый экземпляр FloodControl с заданной конфигурацией
func NewFloodControl(config Config) FloodControl {
 return &floodControl{
  config: config,
  checks: make(map[int64][]time.Time),
 }
}

// Check реализация метода интерфейса FloodControl
func (fc *floodControl) Check(ctx context.Context, userID int64) (bool, error) {
 fc.checksLock.Lock()
 defer fc.checksLock.Unlock()

 now := time.Now()
 checks, ok := fc.checks[userID]
 if !ok {
  fc.checks[userID] = []time.Time{now}
  return true, nil
 }

 // Удаляем все запросы, которые были до интервала
 for len(checks) > 0 && now.Sub(checks[0]) > fc.config.Interval {
  checks = checks[1:]
 }
 fc.checks[userID] = append(checks, now)

 if len(fc.checks[userID]) > fc.config.MaxChecks {
  return false, nil
 }

 return true, nil
}

func main() {
 // Пример использования
 config := Config{
  Interval:  time.Second * 10,
  MaxChecks: 5,
 }
 fc := NewFloodControl(config)

 ctx := context.Background()
 userID := int64(123)

 for i := 0; i < 6; i++ {
  ok, err := fc.Check(ctx, userID)
  if err != nil {
   panic(err)
  }
  if ok {
   println("Request passed flood control check")
  } else {
   println("Request failed flood control check")
  }
  time.Sleep(time.Second)
 }
}
