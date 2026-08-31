import {api} from './client'
import type {TenantContext} from './auth'
export type Tenant=TenantContext&{description?:string;status:string;owner_name?:string;created_at?:string;students?:number}
export const botUsername=(import.meta.env.VITE_TELEGRAM_BOT_USERNAME??'calisthenics_app_bot').replace(/^@/,'')
export const tenantShareLink=(slug:string)=>`https://t.me/${botUsername}?startapp=${slug}`
export const listTenants=(token:string)=>api<{tenants:Tenant[]}>('/super-admin/tenants',{},token)
export const getCoachSpace=(token:string)=>api<{tenant:Tenant}>('/coach/space',{},token)
export const updateCoachSpace=(token:string,input:{name:string;description:string})=>api<{tenant:Tenant}>('/coach/space',{method:'PUT',body:JSON.stringify(input)},token)
export const updateCoachSpaceSlug=(token:string,slug:string)=>api<{tenant:Tenant}>('/coach/space/slug',{method:'PUT',body:JSON.stringify({slug})},token)
export const updateCoachSpaceAvatar=(token:string,media_id?:string)=>api<{tenant:Tenant}>('/coach/space/avatar',{method:'PUT',body:JSON.stringify({media_id:media_id??null})},token)
export async function copyShareLink(slug:string){const value=tenantShareLink(slug);if(navigator.clipboard?.writeText){await navigator.clipboard.writeText(value);return}const area=document.createElement('textarea');area.value=value;area.style.position='fixed';area.style.opacity='0';document.body.appendChild(area);area.select();document.execCommand('copy');area.remove()}
