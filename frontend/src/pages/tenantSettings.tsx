import {useEffect,useState} from 'react'
import {useMutation,useQuery,useQueryClient} from '@tanstack/react-query'
import {copyShareLink,getCoachSpace,tenantShareLink,updateCoachSpace,updateCoachSpaceAvatar,updateCoachSpaceSlug,type Tenant} from '../api/tenants'
import {addExternalMedia} from '../api/coach'
import {useSessionStore} from '../store/session'

const slugPattern=/^[a-z0-9]+(?:-[a-z0-9]+)*$/

export function TenantSettingsPage(){
  const token=useSessionStore(s=>s.accessToken)!,updateCurrentTenant=useSessionStore(s=>s.updateCurrentTenant),qc=useQueryClient()
  const q=useQuery({queryKey:['coach-space'],queryFn:()=>getCoachSpace(token)})
  const [name,setName]=useState(''),[description,setDescription]=useState(''),[slug,setSlug]=useState(''),[confirmSlug,setConfirmSlug]=useState(false),[copied,setCopied]=useState(false),[saved,setSaved]=useState(false)
  useEffect(()=>{if(q.data){setName(q.data.tenant.name);setDescription(q.data.tenant.description??'');setSlug(q.data.tenant.slug)}},[q.data])
  const accept=(tenant:Tenant)=>{qc.setQueryData(['coach-space'],{tenant});updateCurrentTenant(tenant)}
  const save=useMutation({mutationFn:()=>updateCoachSpace(token,{name:name.trim(),description:description.trim()}),onSuccess:data=>{accept(data.tenant);setSaved(true)}})
  const changeSlug=useMutation({mutationFn:()=>updateCoachSpaceSlug(token,slug),onSuccess:data=>{accept(data.tenant);setSlug(data.tenant.slug);setConfirmSlug(false);setCopied(false)}})
  const avatar=useMutation({mutationFn:async(file?:File)=>{if(!file)return updateCoachSpaceAvatar(token);if(!['image/jpeg','image/png','image/webp'].includes(file.type)||file.size>10*1024*1024)throw new Error('format');const url=await new Promise<string>((resolve,reject)=>{const reader=new FileReader();reader.onload=()=>resolve(String(reader.result));reader.onerror=()=>reject(reader.error);reader.readAsDataURL(file)});const created=await addExternalMedia(token,{type:'image',url,original_filename:file.name,mime_type:file.type,size_bytes:file.size});return updateCoachSpaceAvatar(token,created.media.id)},onSuccess:data=>accept(data.tenant)})
  if(!q.data)return <div className="notice skeleton">Загружаем пространство…</div>
  const t=q.data.tenant,dirty=name.trim()!==t.name||description.trim()!==(t.description??''),slugDirty=slug!==t.slug,slugValid=slug.length>=2&&slug.length<=63&&slugPattern.test(slug),nextLink=tenantShareLink(slug)
  return <div className="stack tenant-settings">
    <div><p className="eyebrow">МОЯ ШКОЛА</p><h2>Настройки школы</h2></div>
    <section className="card stack school-avatar-settings"><h3>Аватар школы</h3>{t.avatar_url?<img className="tenant-avatar settings-avatar" src={t.avatar_url} alt="Аватар школы"/>:<span className="tenant-avatar tenant-avatar-fallback settings-avatar">{t.avatar_initials}</span>}<label className="button-file"><span>Загрузить фото</span><input type="file" accept="image/jpeg,image/png,image/webp" onChange={e=>{const file=e.target.files?.[0];if(file)avatar.mutate(file)}}/></label>{t.telegram_avatar_url&&t.avatar_media_id&&<button className="secondary-button" onClick={()=>avatar.mutate(undefined)}>Использовать фото Telegram</button>}{t.avatar_media_id&&<button className="secondary-button" onClick={()=>avatar.mutate(undefined)}>Удалить фото</button>}<small className="muted">JPEG, PNG или WebP, до 10 МБ. Фото будет обрезано по центру.</small>{avatar.isError&&<span className="field-error">Не удалось сохранить фото.</span>}</section>
    <section className="card stack settings-form">
      <label><span>Название школы</span><input className="editable-control" value={name} maxLength={120} onChange={e=>{setName(e.target.value);setSaved(false)}} placeholder="Название школы"/></label>
      <label><span>Описание для учеников</span><textarea className="editable-control" value={description} maxLength={1000} onChange={e=>{setDescription(e.target.value);setSaved(false)}} placeholder="Расскажите ученикам о тренировках, подходе или целях школы."/></label>
      <small className="muted">Коротко расскажите о своей школе или мотивируйте учеников. Этот текст будет виден ученикам на главной странице.</small>
      <button disabled={save.isPending||!dirty||name.trim().length<2} onClick={()=>save.mutate()}>Сохранить изменения</button>
      {saved&&<span role="status">Настройки школы сохранены</span>}
    </section>
    <section className="card stack school-address">
      <label><span>Адрес школы</span><input className="editable-control" value={slug} maxLength={63} onChange={e=>setSlug(e.target.value.toLowerCase().replace(/^@/,'').replace(/\s+/g,'-'))} aria-describedby="school-address-preview"/></label>
      <div id="school-address-preview" className="readonly-value"><small>Ссылка для учеников:</small><span>{nextLink}</span></div>
      {slugDirty&&!slugValid&&<span className="field-error">Используйте строчные латинские буквы, цифры и дефисы.</span>}
      <button className="secondary-button" disabled={!slugDirty||!slugValid||changeSlug.isPending} onClick={()=>setConfirmSlug(true)}>Изменить адрес</button>
    </section>
    {confirmSlug&&<section className="modal-backdrop" role="presentation"><div className="card confirm-dialog" role="alertdialog" aria-modal="true" aria-labelledby="slug-confirm-title"><h3 id="slug-confirm-title">Изменить адрес школы?</h3><p>После изменения старые ссылки для учеников перестанут работать.</p><div className="readonly-value"><small>Новый адрес:</small><span>{nextLink}</span></div>{changeSlug.isError&&<p className="field-error">Не удалось изменить адрес. Проверьте, что он свободен и не зарезервирован.</p>}<div className="form-actions"><button className="secondary-button" onClick={()=>setConfirmSlug(false)}>Отмена</button><button disabled={changeSlug.isPending} onClick={()=>changeSlug.mutate()}>Изменить адрес</button></div></div></section>}
    <section className="card share-card"><p className="eyebrow">ССЫЛКА ДЛЯ УЧЕНИКОВ</p><div className="readonly-value"><span>{tenantShareLink(t.slug)}</span></div><button onClick={async()=>{await copyShareLink(t.slug);setCopied(true)}}>Скопировать ссылку</button>{copied&&<span role="status">Ссылка скопирована</span>}</section>
  </div>
}
