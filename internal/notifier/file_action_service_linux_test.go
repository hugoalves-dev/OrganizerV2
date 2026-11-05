//go:build linux

package notifier

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLinuxPopupActionUsesExistingRegistryAndNeverOverwrites(t *testing.T) {
    root := t.TempDir()
    source := filepath.Join(root, "source", "relatório 日本語.txt")
    destination := filepath.Join(root, "destination")
    if err:=os.MkdirAll(filepath.Dir(source),0755);err!=nil {t.Fatal(err)}
    if err:=os.MkdirAll(destination,0755);err!=nil {t.Fatal(err)}
    if err:=os.WriteFile(source,[]byte("new"),0600);err!=nil {t.Fatal(err)}
    if err:=os.WriteFile(filepath.Join(destination,filepath.Base(source)),[]byte("existing"),0600);err!=nil {t.Fatal(err)}
    registry := newNotificationEventRegistry(eventRegistryOptions{cleanupInterval:0})
    defer registry.Close()
    shortcutID := "a" // IDs are opaque; generated configuration uses SHA-256.
    resolver := newNotificationShortcutResolver(nil)
    resolver.byID[shortcutID] = resolvedShortcut{ID:shortcutID,Name:"destination",Path:destination}
    handler := &windowsNotificationActionHandler{
        registry:registry,shortcuts:&resolver,files:&linuxFileActionService{},logf:t.Logf,
    }
    event,err:=registry.Register(source,"Documents")
    if err!=nil {t.Fatal(err)}
    action:=encodeNotificationAction(actionConfirm,event.ID)
    input:=[]inputValue{{Key:destinationInputID,Value:shortcutID}}
    handler.Handle(action,input)
    handler.Handle(action,input) // One-shot: must not move again.
    if _,err:=os.Stat(source);!os.IsNotExist(err){t.Fatalf("source should be moved: %v",err)}
    original,err:=os.ReadFile(filepath.Join(destination,filepath.Base(source)))
    if err!=nil||string(original)!="existing"{t.Fatalf("collision overwritten original: %q %v",original,err)}
    duplicate,err:=os.ReadFile(filepath.Join(destination,"relatório 日本語 (2).txt"))
    if err!=nil||string(duplicate)!="new"{t.Fatalf("collision-safe target missing: %q %v",duplicate,err)}
}

func TestLinuxCopyActionKeepsSourceAndRejectsDirectory(t *testing.T) {
    root:=t.TempDir()
    source:=filepath.Join(root,"file.txt")
    dest:=filepath.Join(root,"dest")
    if err:=os.WriteFile(source,[]byte("contents"),0600);err!=nil{t.Fatal(err)}
    service:=&linuxFileActionService{}
    target,err:=service.CopyTo(source,dest)
    if err!=nil{t.Fatal(err)}
    if _,err:=os.Stat(source);err!=nil{t.Fatalf("copy removed source: %v",err)}
    if data,err:=os.ReadFile(target);err!=nil||string(data)!="contents"{t.Fatalf("bad copy: %v",err)}
    if _,err:=service.MoveTo(root,dest);err==nil{t.Fatal("directory should be rejected")}
}
