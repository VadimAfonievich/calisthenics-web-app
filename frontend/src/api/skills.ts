import {api} from './client'
import {requireEntityID} from './entityIds'
export type SkillStatus='locked'|'available'|'in_progress'|'mastered'
export type Skill={id:string;code:string;name:string;description:string;category:string;difficulty:string;icon:string;xp_reward:number;final_criterion_type:string;final_criterion_value:number;status:SkillStatus;current_level:number;total_levels:number;progress_percent:number}
export type SkillRequirement={skill_id:string;required_skill_id:string;requirement_type:string;requirement_value:number}
export type SkillLevel={id:string;level_number:number;name:string;description:string;criterion_type:string;criterion_value:number;status:'locked'|'available'|'in_progress'|'completed';progress_value:number;workouts:Array<{id:string;title:string;estimated_minutes:number}>}
export const listSkills=(token:string)=>api<{skills:Skill[]}>('/skills',{},token)
export const getSkillMap=(token:string)=>api<{nodes:Skill[];requirements:SkillRequirement[]}>('/skills/map',{},token)
export const getSkill=(token:string,id:string)=>api<{skill:Skill;levels:SkillLevel[]}>(`/skills/${requireEntityID(id,'Skill')}`,{},token)
export const completeSkillLevel=(token:string,id:string,level:number,value:number)=>api<{status:string;level_number:number}>(`/skills/${requireEntityID(id,'Skill')}/levels/${level}/complete`,{method:'POST',body:JSON.stringify({value})},token)
export const masterSkill=(token:string,id:string,value:number)=>api<{skill_id:string;status:string;xp_earned:number;achievement?:string;already_mastered:boolean}>(`/skills/${requireEntityID(id,'Skill')}/master`,{method:'POST',body:JSON.stringify({value})},token)
