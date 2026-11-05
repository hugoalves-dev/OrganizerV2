package main

import (
  "bytes"
  "encoding/json"
  "errors"
  "fmt"
  "net/http"
  "time"
)

func (s *PopupService) Ready() error {
  if s.window == nil {
    return errors.New("popup window unavailable")
  }
  // Show only after the JS has rendered the filename, destinations and buttons.
  // A failed JS import keeps the helper hidden and triggers genuine fallback.
  s.place()
  s.window.Show()
  if err := s.request("/ready", nil); err != nil {
    s.window.Hide()
    s.Close()
    return err
  }
  return nil
}
func (s *PopupService) Action(action, selection string) error {
  if action!="open_location" && action!="open_file" && action!="confirm" {
    return errors.New("unsupported action")
  }
  valid:=false
  for _,c:=range s.options.Choices {if c.ID==selection {valid=true;break}}
  if !valid{return errors.New("unsupported selection")}
  return s.request("/action",map[string]string{"action":action,"selection":selection})
}
func (s *PopupService) request(endpoint string,data any) error {
  encoded,err:=json.Marshal(data)
  if err!=nil{return err}
  req,err:=http.NewRequest(http.MethodPost,"http://"+s.callback+endpoint,bytes.NewReader(encoded))
  if err!=nil{return err}
  req.Header.Set("X-Organizer-Key",s.authKey)
  client:=&http.Client{Timeout:3*time.Second}
  response,err:=client.Do(req)
  if err!=nil{return err}
  defer response.Body.Close()
  if response.StatusCode!=http.StatusNoContent{return fmt.Errorf("notifier rejected request: %d",response.StatusCode)}
  return nil
}
