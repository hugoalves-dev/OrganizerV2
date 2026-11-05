package main

import (
    "embed"
    "encoding/json"
    "flag"
    "log"
    "os"
    "path/filepath"

    "github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed assets
var assets embed.FS

func main(){
    file:=flag.String("file","","organized file path")
    category:=flag.String("category","","category")
    duration:=flag.Int("duration",15,"timeout in seconds")
    event:=flag.String("event","","event identifier")
    callback:=flag.String("callback","","callback address")
    token:=flag.String("token","","callback key")
    choicesRaw:=flag.String("choices","[]","destination selections")
    dragEnabled:=flag.Bool("drag-enabled",true,"drag permission")
    openFile:=flag.Bool("open-file",true,"open file action")
    openLocation:=flag.Bool("open-location",true,"open folder action")
    confirm:=flag.Bool("confirm",true,"confirm action")
    flag.Parse()
    if !filepath.IsAbs(*file)||*event==""||*callback==""||*token==""||*duration<=0||*duration>3600 {
        log.Print("popup requires a valid OrganizerV2 notification")
        os.Exit(2)
    }
    var choices []PopupChoice
    if err:=json.Unmarshal([]byte(*choicesRaw),&choices);err!=nil{log.Fatal(err)}
    details:=PopupDetails{
        Filename:filepath.Base(*file),Folder:filepath.Base(filepath.Dir(*file)),
        Category:*category,Duration:*duration,Choices:choices,
        OpenFile:*openFile,OpenLocation:*openLocation,Confirm:*confirm,DragEnabled:*dragEnabled,
    }
    service:=NewPopupService(*file,*category,*duration)
    service.configure(*callback,*token,details)
    app:=application.New(application.Options{
        Name:"OrganizerV2 Popup",
        Services:[]application.Service{application.NewService(service)},
        Assets:application.AssetOptions{Handler:application.BundledAssetFileServer(assets)},
    })
    service.setApplication(app)
    window:=app.Window.NewWithOptions(application.WebviewWindowOptions{
        Title:"OrganizerV2",Width:380,Height:230,URL:"/",
        Frameless:true,AlwaysOnTop:true,DisableResize:true,
        HideOnEscape:true,DefaultContextMenuDisabled:true,
        Windows:application.WindowsWindow{HiddenOnTaskbar:true,DisableMenu:true},
    })
    service.setWindow(window)
    if err:=app.Run();err!=nil{log.Fatal(err)}
}
