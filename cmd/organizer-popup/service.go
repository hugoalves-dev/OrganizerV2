package main
import (
 "context"
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "sync"
 dragout "github.com/eaxum/wails-dragout"
 wailsdrag "github.com/eaxum/wails-dragout/wails"
 "github.com/wailsapp/wails/v3/pkg/application"
)
type PopupService struct {path string; category string; duration int; app *application.App; mu sync.Mutex}
func NewPopupService(path,category string,duration int)*PopupService{return &PopupService{path:path,category:category,duration:duration}}
func(s *PopupService) SetApplication(app *application.App){s.app=app}
func(s *PopupService) Filename()string{return filepath.Base(s.path)}
func(s *PopupService) Category()string{return s.category}
func(s *PopupService) Duration()int{return s.duration}
func(s *PopupService) Close(){if s.app!=nil{go s.app.Quit()}}
func(s *PopupService) StartDrag(ctx context.Context)(string,error){
 if !s.mu.TryLock(){return "",errors.New("drag already running")}
 defer s.mu.Unlock()
 if !filepath.IsAbs(s.path){return "",errors.New("path must be absolute")}
 info,err:=os.Stat(s.path)
 if err!=nil{return "",fmt.Errorf("file unavailable: %w",err)}
 if !info.Mode().IsRegular(){return "",errors.New("only regular files supported")}
 if !dragout.Available{return "",dragout.ErrUnsupported}
 w,ok:=ctx.Value(application.WindowKey).(*application.WebviewWindow)
 if !ok||w==nil{return "",errors.New("Wails window unavailable")}
 result,err:=wailsdrag.Start(ctx,w,[]string{s.path})
 if err!=nil{return "",err}
 return string(result),nil
}
