package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
	"syscall"
	"os/signal"
)

var (
	h bool
	p string
)

func init() {
	flag.BoolVar(&h, "h", false, "help")
	flag.StringVar(&p, "p", "", "子进程启动命令")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `daemon 守护进程,自动重启
Usage: daemon [-p]

Options:
`)
	flag.PrintDefaults()
}

///
//var daemon_path = "/Users/chihuan/Documents/go_work/gin_daemon/gin_daemon"
func main() {
	flag.Parse()
	if h {
		flag.Usage()
		os.Exit(0)
	}
	if p == "" {
		fmt.Println("-p is nil")
		os.Exit(1)
	}
	daemon_path := p
	ch := make(chan string)
	go start_new_exec(daemon_path, ch) //这里子进程启动
	after := time.NewTimer(time.Second * 5)
	//defer after.Stop()
	c := make(chan os.Signal)
	//监听所有信号
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT) //监听父进程（当前进程） 退出信号
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
				fmt.Println("退出信号: ", s)
				after.Stop()
				os.Exit(0)
			default:
				fmt.Println("其他信号: ", s)
			}
		}
	}()

	for { //通过chan 获取是否需要进行重启
		select {
		case <-ch:
			fmt.Println("重新启动")
			go start_new_exec(daemon_path, ch)
			after.Reset(time.Second * 5)
		case <-after.C:
			fmt.Println("进程运行中")
			after.Reset(time.Second * 60)
		}
	}

}

func start_new_exec(daemon_path string, ch chan string) {
	fmt.Println("子进程启动   excute: ", daemon_path)
	cmd := exec.Command(daemon_path, os.Args[1:]...)
	//将其他命令传入生成出的进程
	// cmd.Stdin=nil                             //给新进程设置文件描述符，可以重定向到文件中
	// cmd.Stdout=os .Stdout
	// cmd.Stderr=os.Stderr
	err := cmd.Start()
	if err != nil {
		fmt.Println("子进程启动启动失败  error: ", err.Error())
		return
	}
	fmt.Println("子进程启动启动成功") //开始执行新进程，不等待新进程退出
	//return

	err = cmd.Wait() //堵塞父进程

	if err != nil {
		fmt.Println(" 子进程中断 error: ", err.Error())
		ch <- "启动"
	}
	return
}
