package main
import (
 "context"
 "github.com/wailsapp/wails/v3/pkg/application"
)
type PopupService struct {path string; category string; duration int; app *application.App}
func NewPopupService(path,category string,duration int)*PopupService{return &PopupService{path:path,category:category,duration:duration}}
func(s *PopupService) SetApplication(app *application.App){s.app=app}
func(s *PopupService) StartDrag(ctx context.Context)(string,error){return "cancelled",nil}
