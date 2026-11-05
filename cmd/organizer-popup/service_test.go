package main

import (
 "context"
 "os"
 "path/filepath"
 "strings"
 "testing"
)
func TestMissingSourceRejectsDrag(t *testing.T) {
 svc:=NewPopupService(filepath.Join(t.TempDir(),"missing.txt"),"Documents",15)
 svc.Configure("","",PopupDetails{DragEnabled:true})
 _,err:=svc.StartDrag(context.Background())
 if err==nil||!strings.Contains(err.Error(),"unavailable"){t.Fatalf("got %v",err)}
}
func TestDirectoriesCannotBeDragged(t *testing.T){
 svc:=NewPopupService(t.TempDir(),"Documents",15)
 svc.Configure("","",PopupDetails{DragEnabled:true})
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
 svc.Configure("127.0.0.1:1","test",PopupDetails{Choices:[]PopupChoice{{ID:"current",Name:"Current"}}})
 if err:=svc.Action("open_file","other");err==nil{t.Fatal("forged destination accepted")}
}

func TestDisabledDragIsRejected(t *testing.T){
 svc:=NewPopupService(filepath.Join(t.TempDir(),"example.txt"),"Documents",15)
 _,err:=svc.StartDrag(context.Background())
 if err==nil||!strings.Contains(err.Error(),"disabled"){t.Fatalf("got %v",err)}
}
