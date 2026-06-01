package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"genericagent-admin-go/internal/api"
	"genericagent-admin-go/internal/config"
	"genericagent-admin-go/internal/modelconfig"
	"genericagent-admin-go/internal/service"
)

//go:embed web/dist
var webFS embed.FS

func main() {
	cwd, err := appRoot()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Chdir(cwd); err != nil {
		log.Fatalf("chdir %s failed: %v", cwd, err)
	}
	cfgStore := config.NewStore(cwd)
	if err := cfgStore.Load(); err != nil {
		log.Printf("load config: %v", err)
	}
	svc := service.NewManager(cfgStore.Cfg.GARoot, cfgStore.Cfg.BufferLines)
	models := modelconfig.NewStore(cwd)
	static, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		log.Fatal(err)
	}
	srv := api.New(cfgStore, svc, models, static)
	petState := newPetActivityState()
	srv.PetEvent = petState.handle
	addr := fmt.Sprintf("%s:%d", cfgStore.Cfg.Host, cfgStore.Cfg.Port)
	url := "http://" + addr
	server := &http.Server{Addr: addr, Handler: srv.Routes()}
	go srv.StartAutostartServices()
	go func() { time.Sleep(500 * time.Millisecond); openBrowser(url) }()
	go func() {
		log.Printf("GenericAgent Admin Go listening on %s", url)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen %s failed: %v; if the port is occupied, edit config.local.json and change port", addr, err)
		}
	}()
	stopPet := startDesktopPet()
	runTray(url,
		func() { openBrowser(url) },
		func() { openBrowser(url + "/chat") },
		func() { showDesktopPet() },
		func() { hideDesktopPet() },
		func() { srv.StopManagedServices() },
		func() {
			stopPet()
			srv.ShutdownCleanup()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = server.Shutdown(ctx)
		},
	)
}

type petActivityState struct {
	mu      sync.Mutex
	running map[string]bool
}

func newPetActivityState() *petActivityState {
	return &petActivityState{running: map[string]bool{}}
}

func (p *petActivityState) handle(event string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch event {
	case "chat:start":
		p.markRunning("chat", true)
		setDesktopPetAction(petActionRunning)
	case "goal:start", "goal:active":
		p.markRunning("goal", true)
		setDesktopPetAction(petActionRunning)
	case "react:start":
		p.markRunning("react", true)
		setDesktopPetAction(petActionRunning)
	case "chat:done", "goal:review", "react:review":
		p.markDomainDone(event)
		p.restoreBaseLocked()
		setDesktopPetActionForTicks(petActionReview, 30)
	case "chat:cancel", "goal:stop", "react:stop":
		p.markDomainDone(event)
		p.restoreBaseLocked()
		setDesktopPetActionForTicks(petActionWaiting, 20)
	case "goal:idle":
		p.markRunning("goal", false)
		p.restoreBaseLocked()
	case "service:start", "tmwebdriver:start":
		setDesktopPetActionForTicks(petActionRunningRight, 18)
	case "service:stop", "service:stop_all":
		setDesktopPetActionForTicks(petActionRunningLeft, 18)
	case "service:autostart", "tmwebdriver:ready":
		setDesktopPetActionForTicks(petActionWaving, 20)
	case "service:error", "react:error", "goal:error", "chat:error", "tmwebdriver:error":
		p.markDomainDone(event)
		p.restoreBaseLocked()
		setDesktopPetActionForTicks(petActionFailed, 24)
	}
}

func (p *petActivityState) markRunning(domain string, running bool) {
	if running {
		p.running[domain] = true
	} else {
		delete(p.running, domain)
	}
}

func (p *petActivityState) markDomainDone(event string) {
	for _, domain := range []string{"chat", "goal", "react"} {
		if len(event) >= len(domain) && event[:len(domain)] == domain {
			p.markRunning(domain, false)
			return
		}
	}
}

func (p *petActivityState) restoreBaseLocked() {
	if len(p.running) > 0 {
		setDesktopPetAction(petActionRunning)
		return
	}
	setDesktopPetAction(petActionIdle)
}

func appRoot() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if exe != "" {
		return filepath.Dir(exe), nil
	}
	return os.Getwd()
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	hideChildWindow(cmd)
	_ = cmd.Start()
}
