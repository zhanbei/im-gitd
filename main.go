package main

import (
	"flag"
	"log"

	"github.com/zhanbei/im-gitd/boot"
	"github.com/zhanbei/im-gitd/httpd"
	"github.com/zhanbei/im-gitd/sshd"
)

func main() {
	gitDir := flag.String("git-dir", "", "path to git directory (.git/ or a bare repo)")
	httpAddr := flag.String("http-addr", ":8080", "http address to serve on")
	sshAddr := flag.String("ssh-addr", ":8081", "ssh address to serve on")
	flag.Parse()

	cfg := &boot.GitServerConfigs{
		*gitDir, *httpAddr, *sshAddr,
	}
	boot.StartServer(cfg)

	errc := make(chan error, 2)
	go func() {
		errc <- sshd.RunSshServer(*gitDir, *sshAddr)
	}()
	go func() {
		errc <- httpd.RunHttpServer(*gitDir, *httpAddr)
	}()
	for i := 0; i < cap(errc); i++ {
		err := <-errc
		if err != nil {
			log.Println(err)
		}
	}
}
