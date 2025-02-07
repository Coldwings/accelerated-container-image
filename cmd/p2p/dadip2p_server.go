/*
   Copyright The Accelerated Container Image Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/alibaba/accelerated-container-image/pkg/p2p"

	log "github.com/sirupsen/logrus"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, " ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	root       arrayFlags
	port       int
	prefetch   int
	cachesize  int64
	maxentry   int64
	loglevel   string
	media      string
	nodeip     string
	detectaddr string
)

func getOutbondIP() net.IP {
	conn, err := net.Dial("udp", detectaddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func init() {
	flag.Var(&root, "r", "Root list")
	flag.StringVar(&nodeip, "h", "", "Current node IP Address")
	flag.StringVar(&detectaddr, "d", "8.8.8.8:80", "Address for detecting current node address")
	flag.IntVar(&port, "p", 19145, "Listening port")
	flag.Int64Var(&cachesize, "m", 8*1024*1024*1024, "Cache size")
	flag.Int64Var(&maxentry, "e", 1*1024*1024*1024, "Max cache entry")
	flag.IntVar(&prefetch, "pre", 64, "Prefetch workers")
	flag.StringVar(&loglevel, "l", "info", "Log level, debug | info | warn | error | panic")
	flag.StringVar(&media, "c", "/tmp/cache", "Cache media path")
}

func main() {
	flag.Parse()
	level, err := log.ParseLevel(loglevel)
	if err != nil {
		level = log.InfoLevel
		log.Warnf("Log level argument %s not recognized", loglevel)
	}
	log.SetLevel(level)
	rand.Seed(time.Now().Unix())
	if len(root) == 0 {
		log.Info("P2P Root")
	} else {
		log.Info("P2P Agent")
	}
	cache := p2p.NewCachePool(&p2p.CacheConfig{
		MaxEntry:   maxentry,
		CacheSize:  cachesize,
		CacheMedia: media,
	})
	hp := p2p.NewHostPicker(root, cache)
	fs := p2p.NewP2PFS(&p2p.FSConfig{
		CachePool:       cache,
		HostPicker:      hp,
		APIKey:          "dadip2p",
		PrefetchWorkers: prefetch,
	})
	if nodeip == "" {
		nodeip = getOutbondIP().String()
	}
	server := p2p.NewP2PServer(&p2p.ServerConfig{
		MyAddr: fmt.Sprintf("http://%s:%d", nodeip, port),
		Fs:     fs,
		APIKey: "dadip2p",
	})
	http.ListenAndServe(fmt.Sprintf(":%d", port), server)
}
