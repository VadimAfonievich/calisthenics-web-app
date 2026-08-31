import {api} from './client'
import {requireEntityID} from './entityIds'
export type RoleUser={id:string;telegram_id:number;username?:string;display_name:string;role:'user'|'coach'|'admin'|'super_admin'}
export const searchUsers=(token:string,search='')=>api<{users:RoleUser[]}>(`/super-admin/users?search=${encodeURIComponent(search)}`,{},token)
export const setCoachRole=(token:string,id:string,role:'user'|'coach',space?:{name:string;slug:string})=>api<{role:string;slug?:string}>(`/super-admin/users/${requireEntityID(id,'User')}/role`,{method:'PUT',body:JSON.stringify({role,...space})},token)
