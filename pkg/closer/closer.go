// Замыкает на себе логику регистрации и вызова функций завершения, ожидания их отработки.
package closer

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

// Содержит методы регистрации callback-функций, их отложенного вызова и ожидания завершения их отработки.
type Closer struct {
	mu        sync.Mutex
	once      sync.Once
	funcs     []CloseFunc
	funcsDone chan struct{}
	shutdown  chan os.Signal
	notify    chan os.Signal
}

// Callback-функция завершения.
type CloseFunc func() error

// Создает Closer, который при перехвате одного из заданных сигналов вызовет переданные через Add функции завершения.
func NewCloser(signals ...os.Signal) *Closer {
	c := &Closer{
		funcsDone: make(chan struct{}),
		shutdown:  make(chan os.Signal, 1),
		notify:    make(chan os.Signal, 1),
	}

	if len(signals) > 0 {
		go func() {
			signal.Notify(c.shutdown, signals...)
			<-c.shutdown
			close(c.notify)
			signal.Stop(c.shutdown)
			c.CloseAll()
		}()
	}

	return c
}

// Сигнализирует о необходимости вызова функций завершения.
func (c *Closer) Signal() {
	close(c.shutdown)
}

// Регистрирует callback-функцию для выполнения при вызове CloseAll.
func (c *Closer) Add(f ...CloseFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.funcs = append(c.funcs, f...)
}

// Ожидает завершения всех функций, добавленных через Add.
func (c *Closer) Wait() {
	<-c.funcsDone
}

// Возвращает канал, который закроется при получении сигнала на завершение работы.
func (c *Closer) Notify() <-chan os.Signal {
	return c.notify
}

// Возвращает канал, который закроется при завершении всех функций, добавленных через Add.
func (c *Closer) Done() chan struct{} {
	done := make(chan struct{})
	go func() {
		c.Wait()
		close(done)
	}()

	return done
}

// Завершает работу всех функций, добавленных через Add.
func (c *Closer) CloseAll() {
	c.once.Do(func() {
		defer close(c.funcsDone)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		errs := make(chan error, len(funcs))

		var wg sync.WaitGroup

		for _, f := range funcs {
			wg.Add(1)

			closeFunc := func(f CloseFunc) {
				defer wg.Done()

				if err := f(); err != nil {
					errs <- err
				}
			}
			go closeFunc(f)
		}

		go func() {
			wg.Wait()
			close(errs)
		}()

		for err := range errs {
			if err != nil {
				log.Println("error returned from Closer:", err) // !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			}
		}
	})
}
