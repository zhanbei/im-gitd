package httpd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	"github.com/zhanbei/im-gitd/utils"
)

func RunHttpServer(dir, addr string) error {
	var router = http.NewServeMux()
	router.HandleFunc("/info/refs", httpInfoRefs(dir))
	router.HandleFunc("/git-upload-pack", httpGitUploadPack(dir))
	router.HandleFunc("/git-receive-pack", httpGitReceivePack(dir))

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	serv := &http.Server{
		Addr: addr, ErrorLog: logger,

		Handler: utils.HttpRouterLogging(logger, router),

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	logger.Println("starting http server on", addr)
	err := serv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return runHTTP(dir, addr)
}

var defer2delete = true

func runHTTP(dir, addr string) error {
	if defer2delete {
		return nil
	}
	defer2delete = false
	http.HandleFunc("/info/refs", httpInfoRefs(dir))
	http.HandleFunc("/git-upload-pack", httpGitUploadPack(dir))
	http.HandleFunc("/git-receive-pack", httpGitReceivePack(dir))

	log.Println("starting http server on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func httpInfoRefs(dir string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("httpInfoRefs %s %s", r.Method, r.URL)

		service := r.URL.Query().Get("service")
		if service != "git-upload-pack" && service != "git-receive-pack" {
			http.Error(rw, "only smart git", 403)
			return
		}

		rw.Header().Set("content-type", fmt.Sprintf("application/x-%s-advertisement", service))

		ep, err := transport.NewEndpoint("/")
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
		bfs := osfs.New(dir)
		ld := server.NewFilesystemLoader(bfs)
		svr := server.NewServer(ld)

		var sess transport.Session

		if service == "git-upload-pack" {
			sess, err = svr.NewUploadPackSession(ep, nil)
			if err != nil {
				http.Error(rw, err.Error(), 500)
				log.Println(err)
				return
			}
		} else {
			sess, err = svr.NewReceivePackSession(ep, nil)
			if err != nil {
				http.Error(rw, err.Error(), 500)
				log.Println(err)
				return
			}
		}

		ar, err := sess.AdvertisedReferencesContext(r.Context())
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
		ar.Prefix = [][]byte{
			[]byte(fmt.Sprintf("# service=%s", service)),
			pktline.Flush,
		}
		err = ar.Encode(rw)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
	}
}

func httpGitUploadPack(dir string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("httpGitUploadPack %s %s", r.Method, r.URL)

		rw.Header().Set("content-type", "application/x-git-upload-pack-result")

		upr := packp.NewUploadPackRequest()
		err := upr.Decode(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}

		ep, err := transport.NewEndpoint("/")
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
		bfs := osfs.New(dir)
		ld := server.NewFilesystemLoader(bfs)
		svr := server.NewServer(ld)

		sess, err := svr.NewUploadPackSession(ep, nil)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
		res, err := sess.UploadPack(r.Context(), upr)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}

		err = res.Encode(rw)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
	}
}

func httpGitReceivePack(dir string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("httpGitReceivePack %s %s", r.Method, r.URL)

		rw.Header().Set("content-type", "application/x-git-receive-pack-result")

		upr := packp.NewReferenceUpdateRequest()
		err := upr.Decode(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}

		ep, err := transport.NewEndpoint("/")
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
		bfs := osfs.New(dir)
		ld := server.NewFilesystemLoader(bfs)
		svr := server.NewServer(ld)

		sess, err := svr.NewReceivePackSession(ep, nil)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
		res, err := sess.ReceivePack(r.Context(), upr)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}

		err = res.Encode(rw)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			log.Println(err)
			return
		}
	}
}
