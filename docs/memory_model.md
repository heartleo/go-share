# Go内存模型

### data race例子：

```
func main() {
	var count int64

	var wg sync.WaitGroup
	wg.Add(10000)

	for i := 0; i < 10000; i++ {
		go func() {
			count++
			wg.Done()
		}()
	}

	wg.Wait()
	fmt.Println(count)
}
```

~~~
var (
	mu sync.Mutex
	d  *Data
)

type Data struct {
	Name string
}

func getData() (*Data, error) {
	if d == nil {
		mu.Lock()
		defer mu.Unlock()
		if d == nil {
			d = &Data{
				Name: "haha",
			}
		}
	}
	return d, nil
}

func main() {
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = getData()
		}()
	}
	for {
		time.Sleep(time.Second)
	}
}

~~~

执行`go run -race main.go`会发现存在data race

> 数据竞争定义为对内存位置的写入与对同一位置的另一次读取或写入同时发生

> [*Go竞态检测器*](https://go.dev/blog/race-detector)

*在任何顺序一致的执行中，x = 1或y = 1必须首先发生，然后另一个线程中的读取必须能够观察到它(此赋值事件)，因此r1 = 0，r2 =
0是不可能的。但是在一个TSO系统中，线程1和线程2可能会将它们的写操作排队，然后在任何一个写操作进入内存之前从内存中读取，这样两个读操作都会看到零。*

> [*Litmus Test*](https://diy.inria.fr/www/index.html?record=x86)

~~~shell
Litmus Test: Write Queue (also called Store Buffer)
Can this program see r1 = 0, r2 = 0?

// Thread 1           // Thread 2
x = 1                 y = 1
r1 = y                r2 = x
On sequentially consistent hardware: no.
On x86 (or other TSO): yes!
~~~

### 为什么出现这种情况?

* 编译器和CPU为了优化和提升性能通常会改变程序原本定义的执行顺序，这包括：编译器指令重排、CPU的乱序执行等
  除此之外，由于缓存的关系，在多核CPU下，一个CPU核心的写结果仅发生在该核心最近的缓存下，
  要想被另一个CPU读到则必须等待内存被置换回低级缓存再置换到另一个核心后才能被读到。

由上所知，在程序被写成后，将经过编译器的转换与优化、所运行操作系统或虚拟机等动态优化器的优化，以及CPU硬件平台对指令流的优化才最终得以被执行。
这个过程意味着，对于某一个内存的读取与写入操作，可能被这个过程中任何一个中间步骤进行调整，从而偏离程序员在程序中所指定的原有顺序。
没有内存模型的保障，就无法正确的推演程序在最终被执行时的正确性。

### 如何解决?

> 不同的CPU体系架构允许不同数量的指令重新排序，因此在多个处理器上并行运行的代码可以根据体系架构的不同有不同的结果。
> 黄金标准是顺序一致性，即任何执行都必须表现得好像在不同处理器上执行的程序只是以某种顺序交替在单个处理器上执行。
> 对于开发人员来说，这种模型更容易推理，但是今天没有重要的架构能够提供这种模型，因为较弱的并发保证能够带来性能提升。

* Sequential Consistency (SC) 顺序一致性 问题：平台兼容性 性能低 限制硬件和编译器优化
* Data-Race-Free Sequential Consistency (DRF-SC) 无竞争下的顺序一致性 大部分平台均支持
  通过同步原语对所有语句进行排序，如果能确定一个唯一的顺序，那么这份代码就是满足DRF-SC（并发安全）的。

### 什么是内存模型?

内存模型，是一份语言用户与语言自身、语言自身与所在的操作系统平台、 所在操作系统平台与硬件平台之间的契约。
它定义了并行状态下拥有确定读取和写入的时序的条件，并回答了一个共享变量是否具有足够的同步机制来保障一个线程的写入能否发生在另一个线程的读取之前这个问题

简而言之：
程序员负责编写DRF-SC代码
编译器和CPU在优化DRF-SC代码时保证顺序一致性

**Go的内存模型始于以下建议:**

* 修改由多个goroutine同时访问的数据的程序必须串行化这些访问。
* 为了实现串行访问, 需要使用channel操作或其他同步原语(如sync和sync/atomic包中的原语)来保护数据。
* 如果你必须阅读本文的其余部分才能理解你的程序的行为，那你太聪明了。
* 别自作聪明。

### Go内存模型是什么？

> [*Go Memory Model*](https://go.dev/ref/mem)

happens-before是一种严格偏序，它是针对程序的一次执行定义的顺序，该次执行满足所有的内存操作都是原子的，并且以程序顺序执行。
对于这样一次执行，我们称不同处理器上的两个操作具有happens-before关系，仅当它们之间存在一个同步操作。

* 初始化：main.init < main.main
* Goroutine 创建: go() < Goroutine 开始执行
* Goroutine 销毁: 不做任何保证

---
如果 ch 是一个 buffered channel，则 ch<-val < val <- ch
如果 ch 是一个 buffered channel 则 close(ch) < val <- ch & val == isZero(val)
如果 ch 是一个 unbuffered channel 则，ch<-val > val <- ch
如果 ch 是一个容量 len(ch) == C 的 buffered channel，则 从 channel 中收到第 k 个值 < k+C 个值的发送完成

*A send on a channel is synchronized before the completion of the corresponding receive from that channel.*

*如果 ch 是一个 buffered channel，则: ch<-val < val <- ch*

~~~
var c = make(chan int, 10)
var a string

func f() {
	a = "hello, world" // 1
	c <- 0             // 2
}

func main() {
	go f()
	<-c      // 3
	print(a) // 4
	// is guaranteed to print "hello, world".
	// 1<2 2<3 3<4 => 1<4
}
~~~

*The closing of a channel is synchronized before a receive that returns a zero value because the channel is closed.*

*如果 ch 是一个 buffered channel 则: close(ch) < val <- ch & val == isZero(val)*

~~~
var c = make(chan int, 10)
var a string

func f() {
	a = "hello, world" // 1
	close(c)           // 2
}

func main() {
	go f()
	<-c      // 3
	print(a) // 4
	// is guaranteed to print "hello, world".
	// 1<2 2<3 3<4 => 1<4
}
~~~

*A receive from an unbuffered channel is synchronized before the completion of the corresponding send on that channel.*

*如果 ch 是一个 unbuffered channel 则：val <- ch < ch<-val*

~~~
var c = make(chan int)
var a string

func f() {
	a = "hello, world" // 1
	<-c                // 2
}

func main() {
	go f()
	c <- 0   // 3
	print(a) // 4
	// is guaranteed to print "hello, world"
	// 1<2 2<3 3<4 => 1<4
}
~~~

*The kth receive on a channel with capacity C is synchronized before the completion of the k+Cth send from that channel
completes.*

*如果 ch 是一个容量 cap(ch) == C 的 buffered channel，则：从 channel 中接收第k个值 < k+C 个值的发送完成*

~~~
var limit = make(chan int, 3)

func main() {
	for _, w := range work {
		go func(w func()) {
			limit <- 1
			w()
			<-limit
		}(w)
	}
	select{}
}
~~~

---

* mutex: 如果对于 sync.Mutex/sync.RWMutex 的锁 l 有 n < m, 则第 n 次调用 l.Unlock() < 第 m 次调用 l.Lock() 的返回
* mutex: 任何发生在 sync.RWMutex 上的调用 l.RLock, 存在一个 n 使得 l.RLock > 第 n 次调用 l.Unlock，且与之匹配的
  l.RUnlock < 第 n+1 次调用 l.Lock
* once: f() 在 once.Do(f) 中的调用 < once.Do(f) 的返回

### 总结

1. 尽可能写出没有data race的代码（使用actor模型、消息队列等）
2. 理解Go内存模型, 正确运用同步原语
3. 并发代码一定要使用-race来检测data race

### 引用

[*Go内存模型*](https://go.dev/ref/mem)
[*Go竞态检测器*](https://go.dev/blog/race-detector)
[*Litmus Test在线*](https://diy.inria.fr/www/index.html?record=x86)
[*Litmus Test工具*](https://github.com/herd/herdtools7)
