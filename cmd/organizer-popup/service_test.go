package main

import (
 "context"
 "os"
 "path/filepath"
 "strings"
 "testing"
 "github.com/wailsapp/wails/v3/pkg/application"
)
func TestMissingSourceRejectsDrag(t *testing.T) {
 svc:=NewPopupService(filepath.Join(t.TempDir(),"missing.txt"),"Documents",15)
 svc.configure("","",PopupDetails{DragEnabled:true})
 _,err:=svc.StartDrag(context.Background())
 if err==nil||!strings.Contains(err.Error(),"unavailable"){t.Fatalf("got %v",err)}
}
func TestDirectoriesCannotBeDragged(t *testing.T){
 svc:=NewPopupService(t.TempDir(),"Documents",15)
 svc.configure("","",PopupDetails{DragEnabled:true})
 _,err:=svc.StartDrag(context.Background())
 if err==nil||!strings.Contains(err.Error(),"regular"){t.Fatalf("got %v",err)}
}
func TestUnicodeFilename(t *testing.T){
 path:=filepath.Join(t.TempDir(),"relatório 日本語 final.txt")
 if err:=os.WriteFile(path,[]byte("x"),0600);err!=nil{t.Fatal(err)}
 svc:=NewPopupService(path,"Documents",15)
 if svc.Filename()!=filepath.Base(path){t.Fatal("filename was modified")}
}
func TestUnknownShortcutNotSentToWatcher(t *testing.T){
 svc:=NewPopupService(filepath.Join(t.TempDir(),"example.txt"),"Documents",15)
 svc.configure("127.0.0.1:1","test",PopupDetails{Choices:[]PopupChoice{{ID:"current",Name:"Current"}}})
 if err:=svc.Action("open_file","other");err==nil{t.Fatal("forged destination accepted")}
}

func TestDisabledDragIsRejected(t *testing.T){
 svc:=NewPopupService(filepath.Join(t.TempDir(),"example.txt"),"Documents",15)
 _,err:=svc.StartDrag(context.Background())
 if err==nil||!strings.Contains(err.Error(),"disabled"){t.Fatalf("got %v",err)}
}

func TestPopupBottomRightUsesWorkArea(t *testing.T) {
  cases := []struct {
    name string
    area application.Rect
    wantX int
    wantY int
  }{
    {"primary", application.Rect{X:0,Y:0,Width:1920,Height:1040}, 1524, 754},
    {"negative origin", application.Rect{X:-1600,Y:120,Width:1600,Height:860}, -396, 694},
    {"tiny display clamps", application.Rect{X:100,Y:200,Width:320,Height:200}, 100, 200},
  }
  for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T){
      x,y:=popupBottomRight(tc.area,popupWidth,popupHeight,popupMargin)
      if x!=tc.wantX || y!=tc.wantY {t.Fatalf("got (%d,%d), want (%d,%d)",x,y,tc.wantX,tc.wantY)}
    })
  }
}
