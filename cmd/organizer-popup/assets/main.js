// Wails v3 generates this module during the separate helper build.
import { Details, Ready, StartDrag, Action, Close } from './bindings/github.com/vitorhugo-java/organizerv2-popup/popupservice.js';
const el=id=>document.getElementById(id);
let timer, dragging=false;
const dismiss=()=>Close().catch(()=>{});
function schedule(seconds){clearTimeout(timer);timer=setTimeout(()=>{if(!dragging)dismiss();},seconds*1000)}
async function initialize(){
  const details=await Details();
  el('filename').textContent=details.filename;
  el('folder').textContent='In '+details.folder+' • '+details.category;
  for(const choice of details.choices){
    const opt=document.createElement('option');
    opt.value=choice.id;opt.textContent=choice.name;
    el('destination').append(opt);
  }
  if(details.choices.length<2)el('destination').disabled=true;
  for(const [enabled,code,title] of [
    [details.openLocation,'open_location','Open Location'],
    [details.openFile,'open_file','Open File'],
    [details.confirm,'confirm','Confirm']
  ]){
    if(!enabled)continue;
    const button=document.createElement('button');button.textContent=title;
    button.onclick=async()=>{
      if(dragging)return;
      button.disabled=true;
      try{await Action(code,el('destination').value);dismiss();}
      catch(e){el('status').textContent='Action failed: '+String(e);button.disabled=false;}
    };
    el('actions').append(button);
  }
  el('close').onclick=dismiss;
  if(!details.dragEnabled){
    el('drag').draggable=false;
    el('status').textContent='Drag-out disabled';
  }else{
    el('drag').addEventListener('dragstart',async e=>{
      // Avoid WebView's default HTML/text drag: hand off to the native backend.
      e.preventDefault();e.stopPropagation();
      if(dragging)return;
      dragging=true;clearTimeout(timer);
      el('status').textContent='Native drag in progress…';
      try{
        const result=await StartDrag();
        el('status').textContent=result==='dropped'
          ? 'Drop accepted — verify copy in Explorer.'
          : 'Drag cancelled. Original preserved.';
      }catch(e){el('status').textContent='Drag failed: '+String(e);}
      finally{dragging=false;schedule(details.duration);}
    });
  }
  const panel=document.querySelector('main');
  panel.addEventListener('mouseenter',()=>clearTimeout(timer));
  panel.addEventListener('mouseleave',()=>{if(!dragging)schedule(details.duration)});
  await Ready();schedule(details.duration);
}
initialize().catch(e=>{el('status').textContent='Popup error: '+String(e)});
