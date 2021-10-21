# 0 开始之前

``` shell

源码版本：
go 1.17
应用场景：
1) 上下文信息传递
2) 控制goroutine的执行(超时&取消)
```

## 1 详细说明

### 1.1 interface

``` go
type Context interface {
    // Deadline 返回context的deadline(超时则会被取消)
    Deadline() (deadline time.Time, ok bool)
    // Done 返回chan对象
    Done() <-chan struct{}
    // Err 如果Done未被close, 返回nil; 若被close, 返回被close的原因
    Err() error
    // Value 获取key对应的值
    Value(key interface{}) interface{}
}
```

实际使用时，一般是将Context作为函数第一个参数, 来接收调用方传入Context的实例。

``` go
func (srv *Server) Shutdown(ctx context.Context) error {
    timer := time.NewTimer(nextPollInterval())
    defer timer.Stop()
    for {
        if srv.closeIdleConns() && srv.numListeners() == 0 {
            return lnerr
        }
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-timer.C:
            timer.Reset(nextPollInterval())
        }
    }
}
```

### 1.2 创建Context

#### 1.2.1 创建不带控制的Context

使用Background或者TODO来创建, 从源码来看这俩其实是等效的。

``` go
type emptyCtx int

var (
    background = new(emptyCtx)
    todo       = new(emptyCtx)
)

func Background() Context {
    return background
}

func TODO() Context {
    return todo
}
```

当被调函数需要Context类型的参数时, 可以通过context.TODO()或者context.Background()来创建。

``` go
Shutdown(context.TODO())
Shutdown(context.Background())
```

#### 1.2.2 创建带取消控制的Context

使用context.WithCancel()创建带取消控制的ctx。context.WithCancel()会返回context和cancle函数, 通过调用cancle函数则可以完成对应context的cancle。

``` gp
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
    if parent == nil {
        panic("cannot create context from nil parent")
    }
    c := newCancelCtx(parent)
    propagateCancel(parent, &c)
    return &c, func() { c.cancel(true, Canceled) }
}

func newCancelCtx(parent Context) cancelCtx {
    return cancelCtx{Context: parent}
}
```

实际使用context.WithCancel()创建带取消控制的ctx时, 通过调用cancle函数则可以完成对应context的cancle, 由于关闭了context的Done对应的channel, 因此这里会输出对应的错误信息。

``` go
cancelCtx, cancel := context.WithCancel(ctx)
cancel()
select {
case <-cancelCtx.Done():
    fmt.Println("cancelCtx done, err:", cancelCtx)
default:
}

OUTPUT:
cancelCtx done, err: context.Background.WithValue(type string, val value).WithCancel
```

#### 1.2.3 创建带超时控制的Context

使用context.WithTimeout()或者context.WithDeadline()创建。实际上，WithTimeout其实就是对WithDeadline的封装。

``` go
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
    return WithDeadline(parent, time.Now().Add(timeout))
}

func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
    if parent == nil {
        panic("cannot create context from nil parent")
    }
    if cur, ok := parent.Deadline(); ok && cur.Before(d) {
        // The current deadline is already sooner than the new one.
        return WithCancel(parent)
    }
    c := &timerCtx{
        cancelCtx: newCancelCtx(parent),
        deadline:  d,
    }
    propagateCancel(parent, c)
    dur := time.Until(d)
    if dur <= 0 {
        c.cancel(true, DeadlineExceeded) // deadline has already passed
        return c, func() { c.cancel(false, Canceled) }
    }
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.err == nil {
        c.timer = time.AfterFunc(dur, func() {
        c.cancel(true, DeadlineExceeded)
        })
    }
    return c, func() { c.cancel(true, Canceled) }
}
```

这里的示例是通过WithTimeout创建一个有效期为10s的context。

``` go
ctx, _ = context.WithTimeout(ctx, 10*time.Second)
select {
case <-ctx.Done():
    fmt.Println("ctx->Done,", ctx.Err())
}
```

## 3 值传递

使用context.WithValue()来设置key val到context里。并且WithValue()会返回一个valueCtx类型的context。

``` go
type valueCtx struct {
    Context
    key, val interface{}
}

func WithValue(parent Context, key, val interface{}) Context {
    if parent == nil {
        panic("cannot create context from nil parent")
    }
    if key == nil {
        panic("nil key")
    }
    if !reflectlite.TypeOf(key).Comparable() {
        panic("key is not comparable")
    }
    return &valueCtx{parent, key, val}
}

```

使用Value()来获取对应的val, 这里是做的链式查找, 写法还是很妙的。

``` go
func (c *valueCtx) Value(key interface{}) interface{} {
    if c.key == key {
        return c.val
    }
    return c.Context.Value(key)
}
```

## 3 总结

context提供了值传递和协程控制两个功能。对于协程控制，使用channel提供了超时取消和主动取消两种能力。因此，检测是否超时需要搭配select来使用。
