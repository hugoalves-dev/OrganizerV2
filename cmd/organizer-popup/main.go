package main

import (
 "embed"
 "flag"
 "log"
 "path/filepath"
 "os"

 "github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed assets
var assets embed.FS

func main(){
 file:=flag.String("file","","absolute file path")
 category:=flag.String("category","","category")
 duration:=flag.Int("duration",15,"duration seconds")
 flag.Parse()
 if !filepath.IsAbs(*file){log.Print("invalid file path");os.Exit(2)}
 details:=NewPopupService(*file,*category,*duration)
 app:=application.New(application.Options{Name:"OrganizerV2 Popup",Services:[]application.Service{application.NewService(details)},Assets:application.AssetOptions{Handler:application.BundledAssetFileServer(assets)}})
 details.SetApplication(app)
 app.Window.NewWithOptions(application.WebviewWindowOptions{
    Title:"OrganizerV2",Width:380,Height:230,URL:"/",Frameless:true,
    AlwaysOnTop:true,DisableResize:true,
    Windows:application.WindowsWindow{HiddenOnTaskbar:true,DisableMenu:true},
 })
 if err:=app.Run();err!=nil{log.Fatal(err)}
}
