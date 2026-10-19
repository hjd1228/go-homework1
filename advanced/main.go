// 第二批练习题骨架（Go 语言特性方向）：指针 / Goroutine / 面向对象 / Channel / 锁机制
//
// 与 homework.go 中的 8 道算法题相互独立：
//   - homework.go 属于 package homework01，由 homework_test.go 自动评分
//   - 本文件属于 package main，自带入口，运行方式：go run ./advanced
//
// 每个函数上方的注释标明「题目」与「考察点」，函数体为待实现的 TODO。
package main

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	// ---- 指针 ----
	value := 5
	IncreaseByTen(&value)
	fmt.Println("[1] 期望 15，实际:", value)

	nums := []int{1, 2, 3}
	DoubleSlice(&nums)
	fmt.Println("[2] 期望 [2 4 6]，实际:", nums)

	// ---- Goroutine ----
	PrintOddEven()

	tasks := []func(){
		func() { time.Sleep(50 * time.Millisecond) },
		func() { time.Sleep(100 * time.Millisecond) },
		func() { time.Sleep(30 * time.Millisecond) },
	}
	durations := RunTasks(tasks)
	fmt.Println("[4] 各任务耗时:", durations)

	// ---- 面向对象 ----
	shapes := []Shape{Rectangle{Width: 3, Height: 4}, Circle{Radius: 2}}
	for _, shape := range shapes {
		fmt.Println("[5] 面积:", shape.Area(), "周长:", shape.Perimeter())
	}

	employee := Employee{Person: Person{Name: "张三", Age: 30}, EmployeeID: "E001"}
	employee.PrintInfo()

	// ---- Channel ----
	ChannelBasic()
	ChannelBuffered()

	// ---- 锁机制 ----
	fmt.Println("[9] Mutex 计数器 期望 10000，实际:", CountWithMutex(10, 1000))
	fmt.Println("[10] atomic 计数器 期望 10000，实际:", CountWithAtomic(10, 1000))
}

// ---------------------------------------------------------------------------
// 一、指针
// ---------------------------------------------------------------------------

// 题目 1：编写一个函数，接收一个整数指针作为参数，在函数内部将该指针指向的值增加 10，
// 然后在 main 函数中调用该函数并输出修改后的值。
// 考察点：指针的使用、值传递与引用传递的区别。
func IncreaseByTen(n *int) {
	*n += 10
}

// 题目 2：实现一个函数，接收一个整数切片的指针，将切片中的每个元素乘以 2。
// 考察点：指针运算、切片操作。
func DoubleSlice(nums *[]int) {
	for i := range *nums {
		(*nums)[i] *= 2
	}
}

// ---------------------------------------------------------------------------
// 二、Goroutine
// ---------------------------------------------------------------------------

// 题目 3：编写一个程序，使用 go 关键字启动两个协程，
// 一个协程打印从 1 到 10 的奇数，另一个协程打印从 2 到 10 的偶数。
// 考察点：go 关键字的使用、协程的并发执行。
func PrintOddEven() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 1; i <= 10; i += 2 {
			fmt.Println("[3] 奇数:", i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 2; i <= 10; i += 2 {
			fmt.Println("[3] 偶数:", i)
		}
	}()

	wg.Wait()
}

// 题目 4：设计一个任务调度器，接收一组任务（可以用函数表示），并使用协程并发执行这些任务，
// 同时统计每个任务的执行时间。
// 考察点：协程原理、并发任务调度。
func RunTasks(tasks []func()) []time.Duration {
	durations := make([]time.Duration, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		i, task := i, task // Go 1.20 下循环变量会被复用，必须显式复制一份
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			task()
			durations[i] = time.Since(start)
		}()
	}

	wg.Wait()
	return durations
}

// ---------------------------------------------------------------------------
// 三、面向对象
// ---------------------------------------------------------------------------

// 题目 5：定义一个 Shape 接口，包含 Area() 和 Perimeter() 两个方法。
// 然后创建 Rectangle 和 Circle 结构体，实现 Shape 接口。
// 在主函数中，创建这两个结构体的实例，并调用它们的 Area() 和 Perimeter() 方法。
// 考察点：接口的定义与实现、面向对象编程风格。

// Shape 是题目 5 要求的接口。
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Rectangle 需要实现 Shape 接口。
type Rectangle struct {
	Width  float64
	Height float64
}

// Circle 需要实现 Shape 接口。
type Circle struct {
	Radius float64
}

// Area 返回矩形面积。
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter 返回矩形周长。
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Area 返回圆面积。
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Perimeter 返回圆周长。
func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// 题目 6：使用组合的方式创建一个 Person 结构体，包含 Name 和 Age 字段，
// 再创建一个 Employee 结构体，组合 Person 结构体并添加 EmployeeID 字段。
// 为 Employee 结构体实现一个 PrintInfo() 方法，输出员工的信息。
// 考察点：组合的使用、方法接收者。

// Person 是题目 6 的基础结构体。
type Person struct {
	Name string
	Age  int
}

// Employee 通过组合 Person 复用其字段。
type Employee struct {
	Person
	EmployeeID string
}

// PrintInfo 输出员工信息。
func (e Employee) PrintInfo() {
	fmt.Printf("[6] 姓名: %s, 年龄: %d, 工号: %s\n", e.Name, e.Age, e.EmployeeID)
}

// ---------------------------------------------------------------------------
// 四、Channel
// ---------------------------------------------------------------------------

// 题目 7：编写一个程序，使用通道实现两个协程之间的通信。
// 一个协程生成从 1 到 10 的整数，并将这些整数发送到通道中，
// 另一个协程从通道中接收这些整数并打印出来。
// 考察点：通道的基本使用、协程间通信。
func ChannelBasic() {
	ch := make(chan int)

	go func() {
		for i := 1; i <= 10; i++ {
			ch <- i
		}
		close(ch)
	}()

	for v := range ch {
		fmt.Println("[7] 收到:", v)
	}
}

// 题目 8：实现一个带有缓冲的通道，生产者协程向通道中发送 100 个整数，
// 消费者协程从通道中接收这些整数并打印。
// 考察点：通道的缓冲机制。
func ChannelBuffered() {
	ch := make(chan int, 10)

	go func() {
		for i := 1; i <= 100; i++ {
			ch <- i
		}
		close(ch)
	}()

	for v := range ch {
		fmt.Println("[8] 收到:", v)
	}
}

// ---------------------------------------------------------------------------
// 五、锁机制
// ---------------------------------------------------------------------------

// 题目 9：编写一个程序，使用 sync.Mutex 来保护一个共享的计数器。
// 启动 10 个协程，每个协程对计数器进行 1000 次递增操作，最后输出计数器的值。
// 考察点：sync.Mutex 的使用、并发数据安全。
// 提示：实现时需要 import "sync"。
func CountWithMutex(goroutines, increments int) int {
	var mu sync.Mutex
	counter := 0

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return counter
}

// 题目 10：使用原子操作（sync/atomic 包）实现一个无锁的计数器。
// 启动 10 个协程，每个协程对计数器进行 1000 次递增操作，最后输出计数器的值。
// 考察点：原子操作、并发数据安全。
// 提示：实现时需要 import "sync/atomic"。
func CountWithAtomic(goroutines, increments int) int64 {
	var counter atomic.Int64

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				counter.Add(1)
			}
		}()
	}
	wg.Wait()

	return counter.Load()
}
