package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Service struct {
	Path      string `yaml:"path"`
	Base      string `yaml:"base"'`
	Cwd       string `yaml:"cwd"`
	Port      int16  `yaml:"port"`
	Command   string `yaml:"command"`
	Arguments string `yaml:"arguments"`
}

type Config struct {
	Port     int16     `yaml:"port"`
	Services []Service `yaml:"services"`
}

func readConfig() (config Config) {
	c, err := ioutil.ReadFile("multeego.yaml")

	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(c, &config)

	if err != nil {
		log.Fatal(err)
	}

	for i := range config.Services {
		service := &config.Services[i]

		if service.Base == "" {
			service.Base = service.Path
		} else if !strings.HasPrefix(service.Path, service.Base) {
			log.Fatalf("Base '%s' is not part of path '%s'", service.Base, service.Path)
		}
	}

	return
}

func createStaticHandler(service Service) (handler func(http.ResponseWriter, *http.Request)) {
	handler = func(w http.ResponseWriter, req *http.Request) {
		staticPath, err := filepath.Rel(service.Base, req.URL.Path)

		if err != nil {

			http.Error(w, "Bad Request", 400)
			return
		}

		staticPath = filepath.Join(service.Cwd, staticPath)

		info, err := os.Stat(staticPath)

		if err != nil || os.IsNotExist(err) {
			http.NotFound(w, req)
			return
		}

		if info.IsDir() {
			staticPath = filepath.Join(staticPath, "/index.html")
		}

		http.ServeFile(w, req, staticPath)
	}

	return
}

func createForwardingHandler(service Service) (handler func(http.ResponseWriter, *http.Request)) {
	proxyURL, err := url.Parse(fmt.Sprintf("http://localhost:%d/", service.Port))

	if err != nil {
		log.Fatalf("Could not create proxy for service '%s'", service.Cwd)
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyURL)

	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)

		staticPath, _ := filepath.Rel(service.Base, req.URL.Path)

		if staticPath == "." {
			staticPath = ""
		}

		req.URL.Path = fmt.Sprintf("/%s", staticPath)
	}

	handler = func(w http.ResponseWriter, req *http.Request) {
		proxy.ServeHTTP(w, req)
	}

	return
}

func launchProcess(service Service) {
	arguments := strings.Split(service.Arguments, " ")

	for i := range arguments {
		if arguments[i] == "${PORT}" {
			arguments[i] = fmt.Sprintf("%d", service.Port)
		}
	}

	cmd := exec.Command(service.Command, arguments...)
	cmd.Dir = service.Cwd
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", service.Port),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		log.Fatalf("Could not start '%s'", service.Cwd)
	}

}

func main() {
	config := readConfig()

	for _, service := range config.Services {
		if service.Port == 0 {
			log.Printf("Hosting static files from '%s' on %s", service.Cwd, service.Path)
			http.HandleFunc(service.Path, createStaticHandler(service))
		} else {
			log.Printf("Hosting '%s' as a process on %s", service.Cwd, service.Path)
			http.HandleFunc(service.Path, createForwardingHandler(service))

			if service.Command != "" {
				go launchProcess(service)
			}
		}
	}

	log.Printf("Listening on port %d", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
