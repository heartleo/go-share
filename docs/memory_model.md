# Go内存模型

### 一个data race例子：

```go

```

执行`go run -race main.go`会发现存在data race

**为什么出现data race?**

* 每个CPU核心拥有独立的cache
* 多个线程并行在不同的CPU核心上
* 多线程程序可能同时对同一个内存值进行读写
* 编译器优化会重排语句顺序
* CPU优化会进行指令重排（不同架构存在不同的优化策略）

**测试CPU指令重排：**

```go
package memory_model

import (
	"fmt"
	"sync"
)

func OutOfOrder() {
	var N int
	for {
		N++
		var x, y, r1, r2 int
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			x = 1
			r1 = y
			fmt.Println("g1:", x, r1)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			y = 1
			r2 = x
			fmt.Println("g2:", y, r2)
		}()

		wg.Wait()

		if r1 == 0 && r2 == 0 {
			fmt.Println("N:", N)
			break
		}
	}
}

```

*在任何顺序一致的执行中，x = 1或y = 1必须首先发生，然后另一个线程中的读取必须能够观察到它(此赋值事件)，因此r1 = 0，r2 = 0是不可能的。但是在一个TSO系统中，线程1和线程2可能会将它们的写操作排队，然后在任何一个写操作进入内存之前从内存中读取，这样两个读操作都会看到零。*

> [*Litmus Test*](https://go.dev/ref/mem)

~~~shell
Litmus Test: Write Queue (also called Store Buffer)
Can this program see r1 = 0, r2 = 0?

// Thread 1           // Thread 2
x = 1                 y = 1
r1 = y                r2 = x
On sequentially consistent hardware: no.
On x86 (or other TSO): yes!
~~~

**如何避免data race?**
需要定义一个规范，让CPU和编译器知道如何在不破坏代码顺序一致性的情况下进行优化

**什么是内存模型?**

~~内存模型分为硬件内存模型和软件内存模型
硬件内存模型 锁内存总线 锁cpu cache
软件内存模型 java c++~~

~~Sequential Consistency 顺序一致性
DRF-SC Data-Race-Free Sequential Consistency
通过同步原语对所有语句进行排序，如果能确定一个唯一的顺序， 那么这份代码就是满足 DRF-SC（并发安全）的。 编译器、CPU在运行
DRF-SC 的程序的时候应该保证顺序一致性。
内存模型是一个迄今尚无定论的领域~~

> 不同的CPU体系架构允许不同数量的指令重新排序，因此在多个处理器上并行运行的代码可以根据体系架构的不同有不同的结果。
> 黄金标准是顺序一致性，即任何执行都必须表现得好像在不同处理器上执行的程序只是以某种顺序交替在单个处理器上执行。
> 对于开发人员来说，这种模型更容易推理，但是今天没有重要的架构能够提供这种模型，因为较弱的并发保证能够带来性能提升。

内存模型像是一份程序员、编译器、CPU之间的三方契约：
程序员负责编写DRF-SC代码
编译器在优化DRF-SC代码时保证顺序一致性
CPU在执行DRF-SC代码时保证顺序一致性

**为什么要有内存模型**

**Go内存模型是什么？**

> [*Go Memory Model*](https://go.dev/ref/mem)

初始化：
Goroutine创建：
Goroutine销毁：
Channel通信：
锁：
Once：
Atomic Values：

**通过实际的例子来了解Go内存模型**

**设计哲学：**

* 不要通过共享内存来通信，而要通过通信来共享内存
* 清晰好过卖弄聪明
* 简洁好过复杂
* 好的软件应该让人难以误用
* 如果你能写出没有race的代码，那么不需要了解内存模型

**最佳实践**

1. 理解Go内存模型，尽可能写出没有data race的代码
2. 一定要使用-race来检测data race

---

### 参考资料

[*Go Memory Model 译文*](https://go-zh.org/ref/mem)
