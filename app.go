package main

import (
	"context"
	"fmt"
	"headofseo/backend"
	"log"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) StartCrawl(urls string, userAgent string) {
	backend.StartCrawl(a.ctx, urls, userAgent)
}

func (a *App) CancelFetch() {
	backend.CancelFetch(a.ctx)
}

func (a *App) SaveFile(data []backend.Crawl) {
	if len(data) == 0 {
		log.Println("SaveFile: no data to save")
		return
	}

	selection, _ := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Select Directory",
		DefaultFilename: fmt.Sprintf("crawl-result-%s.csv", time.Now().Format("2006-01-02")),
	})

	csvContent, err := gocsv.MarshalString(&data)
	if err != nil {
		// handle error
		log.Println("SaveFile: error ", err)
	}
	f, err := os.Create(selection)
	if err != nil {
		fmt.Println(err)
		return
	}
	l, err := f.WriteString(csvContent)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println("SaveFile: selection ", selection)
	log.Println("SaveFile: err ", err)
}
